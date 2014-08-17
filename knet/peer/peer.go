// Remote bitcoin network peer/node
package peer

import (
    "net"
    "sync"
    "time"
    "errors"
    "github.com/oxfeeefeee/kaiju/knet/btcmsg"
)

const SendQueueSize = 2

const defaultSendMsgTimeout = time.Second * 20

const defaultExpectMsgTimeout = time.Second * 20

// Returns (isTheMessageSwallowed, shouldWeStopExpectingMessage)
type MsgFilter func(btcmsg.Message) (accept bool, stop bool)


// Descriptor of message being sent
type msgSent struct {
    msg         btcmsg.Message
    timeout     time.Duration
    errChan     chan error
}

// Descriptor of message being expected
type msgExpector struct {
    filter      MsgFilter
    timeout     time.Duration
    retChan     chan struct{btcmsg.Message; Error error}
}

// Peer represents and communicates to a remote bitcoin node.
//
// Note that the ip of the remote node is used as the unique Handle of the peer
type Peer struct {
    handle          Handle
    // Standard bitcoin protocol peer info 
    info            *btcmsg.PeerInfo
    // Is this an outgoing or incoming connection? the handshaking differs
    outgoing        bool
    // Network connection to remote node
    conn            net.Conn 
    // For outgoing message
    sendChan        chan *msgSent
    // Expectors expects specific messages
    expectors       []*msgExpector
    // Mutex for expectors
    expMutex        sync.Mutex
    // For tell sender when chan closed
    done            chan struct{}
    // Used to clean up this peer
    onceCleanUp     *sync.Once
    // Embeds monitors
    *monitors

}

func Launch(info *btcmsg.PeerInfo, conn net.Conn, outgoing bool, moni []Monitor) (Handle, error) {
    p := &Peer{
        handle: InvalidHandle,
        info: info,
        outgoing: outgoing,
        conn: conn,
        sendChan: make(chan *msgSent, SendQueueSize),
        expectors: make([]*msgExpector, 0, 2),
        done: make(chan struct{}),
        onceCleanUp: new(sync.Once),
        monitors : new(monitors),
    } 
    err := p.addMonitors(moni)
    if err != nil {
        panic("Failed to add monitor to newly created peer")
    }
    err = p.start()
    if err != nil {
        return InvalidHandle, err
    }
    return peerMgr.addPeer(p)
}

func (p *Peer) Handle() Handle {
    return p.handle
}

func (p *Peer) BtcInfo() *btcmsg.PeerInfo {
    return p.info
}

// Send a bitcoin message to remote peer
// SendMsg mustn't block for Pool to work properly
func (p *Peer) sendMsg(m btcmsg.Message, timeout time.Duration, ch chan error) {
    if timeout <= 0 {
        timeout = defaultSendMsgTimeout
    }
    msent := &msgSent{m, timeout, ch}
    select {
        case p.sendChan <- msent:
        default:
            if msent.errChan != nil {
                msent.errChan <- errors.New("Peer sending queue full.")    
            } 
    }
}

func (p *Peer) expectMsg(f MsgFilter, timeout time.Duration, ch chan struct{btcmsg.Message; Error error}) {
    if timeout <= 0 {
        timeout = defaultExpectMsgTimeout
    }
    exp := &msgExpector{f, timeout, ch}
    p.expMutex.Lock()
    p.expectors = append(p.expectors, exp)
    p.expMutex.Unlock()

    // Remove the expector at time out
    go func(){
        <-time.After(exp.timeout)
        p.expMutex.Lock()
        defer p.expMutex.Unlock()
        exps := p.expectors
        for i, e := range exps {
            if e == exp {
                e.retChan <-struct{btcmsg.Message; Error error}{
                    nil, errors.New("Peer ExpectMsg timeout")}
                // Delete the expector
                p.expectors = append(exps[:i], exps[i+1:]...)
                return
            }
        }
    }()
}

func (p *Peer) start() error{
    err := p.versionHankshake()
    if err != nil {
        p.conn.Close()
        return err;
    }
    go p.onPeerUp(p)
    go p.loopSendMsg()
    go p.loopReceiveMsg()
    // Send getaddr
    err = btcmsg.WriteMsg(p.conn, btcmsg.NewGetAddrMsg())
    if err != nil {
        p.conn.Close()
        return err
    }
    return nil
}

// OK to be called simultaneously by multiple goroutines
func (p *Peer) kill() {
    // This leads to read error, which will end the loop
    p.conn.Close()
}

func (p *Peer) cleanUp() {
    if p.onceCleanUp == nil {
        return
    }
    p.onceCleanUp.Do( 
        func() {
            // Remove onceCleanup in case this instance get reused
            p.onceCleanUp = nil
            close(p.done)
            p.conn.Close()
            // Notify monitors
            p.onPeerDown(p)

            peerMgr.peerDie(p.handle)
       })
}

func (p *Peer) loopSendMsg() {
    // This "running" is used to fix a bug about "break"
    // so if you break in "select", you cannot get out of "for".
    running := true
    for running {
        select {
        case m := <-p.sendChan:
            // "ch" is used do timeout 
            ch := make(chan error, 1)
            go func() { ch <- btcmsg.WriteMsg(p.conn, m.msg) } ()
            select {
            case err := <-ch:
                if m.errChan != nil {
                    m.errChan <- err
                }
                if err != nil {
                    running = false
                } 
            case <-time.After(m.timeout):
                if m.errChan != nil {
                    m.errChan <- errors.New("Peer send message timeout.")
                }
                if m.timeout >= defaultSendMsgTimeout {
                    running = false
                }
            case <-p.done:
                running = false
            }
        case <-p.done:
            running = false
        }
    }
    //log.Debugf("               PEER SEND exit %d", p.handle)
    p.cleanUp()
}

func (p *Peer) loopReceiveMsg() {
    for {
        msg, err := btcmsg.ReadMsg(p.conn) 
        if err != nil {
            //log.Infof("loopReceiveMsg error: %s", err.Error())
            break
        } else {
            if !p.handleMessage(msg) {
                p.onPeerMsg(p.handle, msg)
            }
        }
    }
    //log.Debugf("               PEER RECEIVE exit %d", p.handle)
    p.cleanUp()
}

// Some of the messages are handled here instead of being sent to upper level
// Returns true if we don't want send this message to upper level
func (p *Peer) handleMessage(msg btcmsg.Message) bool {
    switch msg.Command() {
    case "ping":
        ping := msg.(*btcmsg.Message_ping)
        pong := btcmsg.NewPongMsg().(*btcmsg.Message_pong)
        pong.Nonce = ping.Nonce
        p.sendMsg(pong, 0, nil)
        return true
    default:
        p.expMutex.Lock()
        defer p.expMutex.Unlock()
        exps := p.expectors
        for i, e := range exps {
            if accept, stop := e.filter(msg); accept {
                e.retChan <-struct{btcmsg.Message; Error error}{msg, nil}
                // Delete the expector
                if stop {
                    p.expectors = append(exps[:i], exps[i+1:]...)    
                }
                return true
            }
        }
    }
    return false
}

// The first thing to do after a connection is established is to exchange version.
// For kaiju it goes as follows:
// 1. Send out my version message, if outgoing == true
// 2. Expect the remote peer send it's version message
// 3. Once the message from remote peer is received, send verack
// 4. Send out my version message, if outgoing == false
// 5. Done, knet.Peer starts to work
func (p *Peer) versionHankshake() error {
    // Step 1
    if p.outgoing {
        err := p.sendVersionMsg()
        if err != nil {
            return err
        }  
    }
    // Step 2
    msg, err := btcmsg.ReadMsg(p.conn) 
    if err != nil {
        return err
    } else {
        if ver, ok := msg.(*btcmsg.Message_version); ok {
            // TODO: more check
            p.info = ver.Addr_from
        } else {
            return errors.New("Wrong message type when doing versionHankshake")
        }
    }
    // Step 3
    vamsg := btcmsg.NewVerAckMsg()
    err = btcmsg.WriteMsg(p.conn, vamsg)
    if err != nil {
        return err
    }
    // Step 4 
    if !p.outgoing {
        err := p.sendVersionMsg()
        if err != nil {
            return err
        }  
    }
    return nil
}

func (p *Peer) sendVersionMsg() error {
    vmsg := btcmsg.NewLocalVersionMsg(p.info)
    return btcmsg.WriteMsg(p.conn, vmsg) 
}

