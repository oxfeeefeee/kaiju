// Remote bitcoin network peer/node

package peer

import (
    "net"
    "sync"
    "errors"
    "github.com/oxfeeefeee/kaiju/log"
    "github.com/oxfeeefeee/kaiju/kio/btcmsg"
)

// Peer represents and communicates to a remote bitcoin node.
//
// Note that the ip of the remote node is used as the unique ID of the peer
type Peer struct {
    MyID            ID
    // Standard bitcoin protocol peer info 
    info            *btcmsg.PeerInfo
    // Is this an outgoing or incoming connection? the handshaking differs
    outgoing        bool
    // Network connection to remote node
    conn            net.Conn 
    // For outgoing message
    sendChan        chan btcmsg.Message
    // For tell sender when chan closed
    done            chan struct{}
    // Used to clean up this peer
    onceCleanUp     *sync.Once
    // embeds monitors
    *monitors
}

func NewPeer(id ID, conn net.Conn, info *btcmsg.PeerInfo, outgoing bool) *Peer {
    return &Peer{
        MyID: id,
        info: info,
        outgoing: outgoing,
        conn: conn,
        sendChan: make(chan btcmsg.Message),
        done: make(chan struct{}),
        onceCleanUp: new(sync.Once),
        monitors : new(monitors),
    }
}

func (p *Peer) Go() error{
    err := p.versionHankshake()
    if err != nil {
        p.conn.Close()
        return err;
    }

    go p.OnPeerUp(p)
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

// Send a bitcoin message to remote peer
func (p *Peer) SendMsg(msg btcmsg.Message) {
    select {
        case p.sendChan <- msg:
        case <-p.done:
    }
}

// OK to be called simultaneously by multiple goroutines
func (p *Peer) Kill() {
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
            p.OnPeerDown(p)
       })
}

func (p *Peer) loopSendMsg() {
    // This "running" is used to fix a bug about "break"
    // so if you break in "select", you cannot get out of "for".
    running := true
    for running {
        select {
        case msg := <-p.sendChan:
            err := btcmsg.WriteMsg(p.conn, msg)
            if err != nil {
                logger().Printf("loopSendMsg error: %s", err.Error())
                running = false
            }
        case <-p.done:
            running = false
        }
    }
    logger().Debugf("PEER SEND exit")
    p.cleanUp()
}

func (p *Peer) loopReceiveMsg() {
    for {
        msg, err := btcmsg.ReadMsg(p.conn) 
        if err != nil {
            logger().Printf("loopReceiveMsg error: %s", err.Error())
            break
        } else {
            if !p.handleMessage(msg) {
                p.OnPeerRecevieMsg(p.MyID, msg)
            }
        }
    }
    logger().Debugf("PEER RECEIVE exit")
    p.cleanUp()
}

// Some of the messages are handled here instead of sending to upper level
// Returns true if we don't want send this message to upper level
func (p *Peer) handleMessage(msg btcmsg.Message) bool {
    switch msg.Command() {
    case "ping":
        ping := msg.(*btcmsg.Message_ping)
        pong := btcmsg.NewPongMsg().(*btcmsg.Message_pong)
        pong.Nonce = ping.Nonce
        p.SendMsg(pong)
        //logger().Debugf("PONG!!!!!! %v", pong.Nonce)
        return true
    }
    return false
}

// The first thing to do after a connection is established is to exchange version.
// For kaiju it goes as follows:
// 1. Send out my version message, if outgoing == true
// 2. Expect the remote peer send it's version message
// 3. Once the message from remote peer is received, send verack
// 4. Send out my version message, if outgoing == false
// 5. Done, kio.Peer starts to work
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

// Handy function
func logger() *log.Logger {
    return log.KioPeerLogger
}


