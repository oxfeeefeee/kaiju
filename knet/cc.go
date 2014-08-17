// CC for Crowd Control as in CCP Games. CC decides if we connect to a specific remote peer or not
package knet

import (
    "net"
    "time"
    "math"
    "github.com/oxfeeefeee/kaiju"
    "github.com/oxfeeefeee/kaiju/knet/peer"
    "github.com/oxfeeefeee/kaiju/knet/btcmsg"
)

type CC struct {
    ap              *addrPool
    // To control how many dailing in progress
    dialControl     chan struct{}
}

func newCC() *CC {
    return &CC{
        newAddrPool(), 
        make(chan struct{}, kaiju.MaxDialConcurrency), 
    }
}

func (cc *CC) addSeedAddr(ipstr string, port int) {
    ip := net.ParseIP(ipstr)
    peerInfo := btcmsg.NewPeerInfo()
    peerInfo.IP = btcmsg.FromNetIP(&ip)
    peerInfo.Port = uint16(port)
    cc.ap.addAddr(true, peerInfo, 0, 0)
}

func (cc *CC) start(peerMonitors []peer.Monitor) {
    go func() {
        for {
            // Flow control for dial:
            // dialControl is a buffered channel of size MaxDialConcurrency
            cc.dialControl <- struct{}{}
            go cc.doConnect(peerMonitors)      
        }
    }()
}

// Member of peer.Monitor interface
func (cc *CC) ListenTypes() []string {
    return []string{"addr",}
}

// Member of peer.Monitor interface
func (cc *CC) OnPeerUp(p *peer.Peer) {
    // Do nothing
}

// Member of peer.Monitor interface
func (cc *CC) OnPeerDown(p *peer.Peer) {
    // Put it back to the address list
    now := time.Now().Unix()
    p.BtcInfo().Time = uint32(now) // Update last available time to "now"
    cc.ap.addAddr(true, p.BtcInfo(), 1, now)
}

// Member of peer.Monitor interface
func (cc *CC) OnPeerMsg(_ peer.Handle, msg btcmsg.Message) {
    // TODO: addAddr is very slow, gorountine is a workaround for that
    go func (){
        addrMsg := msg.(*btcmsg.Message_addr)
        for _, addr := range addrMsg.Addresses {
            cc.ap.addAddr(false, addr, 0, 0)
        }
    }()
}

func (cc *CC) doConnect(monitors []peer.Monitor) {
    addr, timesFailed := cc.ap.pickBest()
    if addr == nil {
        // Wait for half a second before retry
        time.Sleep(500 * time.Millisecond)
        _ = <- cc.dialControl
        return
    }

    // Lower the frequency of trying to connect for failed peers
    time.Sleep(time.Millisecond * 1000 * time.Duration(math.Exp2(float64(timesFailed))))
    
    // Try to connect to currently best candidate
    // 1. Remove it from the list, then add it back in if failed to start the peer
    cchan := dialAddr(addr)
    conn := <- cchan
    if ok := createPeer(addr, conn, true, monitors); !ok {
        cc.ap.addAddr(true, addr, timesFailed + 1, time.Now().Unix())    
    }
    _ = <- cc.dialControl
}

func createPeer(addr *btcmsg.PeerInfo, conn net.Conn, outgoing bool, monitors []peer.Monitor) bool {
    if conn == nil {
        return false
    }
    _, err := peer.Launch(addr, conn, outgoing, monitors)
    if err != nil {
        return false
    }
    return true
}

func dialAddr(a *btcmsg.PeerInfo) <-chan net.Conn {
    cchan := make(chan net.Conn)
    go func () {
        conn, _ := net.DialTimeout("tcp", a.ToTCPAddr().String(), kaiju.DialTimeout * time.Millisecond)
        cchan <- conn
    }()
    return cchan
}

// Listen needs a public IP to test, do it later
/*
func listen(connChan chan net.Conn, laddr string) {
    go func () {
        listener, err := net.Listen("tcp", laddr)
        if err != nil {
            log.Infof("Error listening: %s", err.Error())
            return;
        }
        defer listener.Close()
 
        for {
            conn, err := listener.Accept()
            if err != nil {
                log.Infof("Error accept: %s", err.Error())
                continue
            }
            connChan <- conn
        }
    }()
}
*/