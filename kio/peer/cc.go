// CC for Crowd Control as in CCP Games. CC decides if we connect to a specific remote peer or not
package peer

import (
    "net"
    "time"
    "github.com/oxfeeefeee/kaiju/cst"
    "github.com/oxfeeefeee/kaiju/kio/btcmsg"
)

type CC struct {
    ap              *addrPool
    // To control how many dailing in progress
    dialControl     chan struct{}
    // Embed the global IDManager for convenience  
    *IDManager
}

func NewCC(im *IDManager) *CC {
    return &CC{
        newAddrPool(im), 
        make(chan struct{}, cst.MaxDialConcurrency), 
        im,
    }
}

// Member of Monitor interface
func (cc *CC) InterestedMsgTypes() []string {
    return []string{"addr",}
}

// Member of Monitor interface
func (cc *CC) OnPeerUp(p *Peer) {
    // Do nothing
}

// Member of Monitor interface
func (cc *CC) OnPeerDown(p *Peer) {
    // Put it back to the address list
    now := time.Now().Unix()
    p.info.Time = uint32(now) // Update last available time to "now"
    cc.ap.addAddr(true, p.info, 1, now)
}

// Member of Monitor interface
func (cc *CC) OnPeerRecevieMsg(id ID, msg btcmsg.Message) {
    addrMsg := msg.(*btcmsg.Message_addr)
    for _, addr := range addrMsg.Addresses {
        cc.ap.addAddr(false, addr, 0, 0)
    }
    //cc.ap.dump()
}

func (cc *CC) AddSeedAddr(ipstr string, port int) {
    ip := net.ParseIP(ipstr)
    peerInfo := btcmsg.NewPeerInfo()
    peerInfo.IP = btcmsg.FromNetIP(&ip)
    peerInfo.Port = uint16(port)
    cc.ap.addAddr(true, peerInfo, 0, 0)
}

func (cc *CC) Go(peerMonitors []Monitor) {
    go func() {
        for {
            // Flow control for dial:
            // dialControl is a buffered channel of size MaxDialConcurrency
            cc.dialControl <- struct{}{}
            go cc.doConnect(peerMonitors)      
        }
    }()
}

func (cc *CC) doConnect(peerMonitors []Monitor) {
    addr, timesFailed := cc.ap.pickOutBest()
    if addr == nil {
        // Wait for half a second before retry
        time.Sleep(500 * time.Millisecond)
        _ = <- cc.dialControl
        return
    }

    // Lower the frequency of trying to connect for failed peers
    time.Sleep(100 * time.Duration(timesFailed) * time.Millisecond)
    
    // Try to connect to currently best candidate
    // 1. Remove it from the list, then add it back in whether failed to connect or not
    cchan := dialAddr(addr)
    conn := <- cchan
    if conn != nil {
        // Spawn a peer
        id := cc.GetID(addr.IP)
        peer := NewPeer(id, conn, addr, true)
        peer.AddMonitors(peerMonitors)
        err := peer.Go()
        if err != nil {
            cc.ap.addAddr(true, addr, timesFailed + 1, time.Now().Unix())
            //logger().Debugf("Failed to do handshake with peer %s: %s", addr.IP.ToNetIP(), err.Error())        
        } else {
            logger().Debugf("Connected to peer %s", addr.IP.ToNetIP())        
        }
    } else {
        cc.ap.addAddr(true, addr, timesFailed + 1, time.Now().Unix())
        //logger().Debugf("Failed to connect to peer %s", addr.IP.ToNetIP())   
    }
    _ = <- cc.dialControl
}

func dialAddr(a *btcmsg.PeerInfo) <-chan net.Conn {
    cchan := make(chan net.Conn)
    go func () {
        conn, _ := net.DialTimeout("tcp", a.ToTCPAddr().String(), cst.DialTimeout * time.Millisecond)
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
            logger().Printf("Error listening: %s", err.Error())
            return;
        }
        defer listener.Close()
 
        for {
            conn, err := listener.Accept()
            if err != nil {
                logger().Printf("Error accept: %s", err.Error())
                continue
            }
            connChan <- conn
        }
    }()
}
*/