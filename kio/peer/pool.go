// Pool manages all remote bitcoin nodes, represented by kio.Peer's
package peer

import (
    "github.com/oxfeeefeee/kaiju/kio/btcmsg"
)

type peerMsg struct {
    peerID      ID
    msg         btcmsg.Message
}

type Pool struct {
    // All Peers
    peers           map[ID]*Peer
    // Channel for requesting to add a peer
    addReqChan      chan *Peer
    // Channel for requesting to remove a peer
    removeReqChan   chan ID
    // Channel for requesting to send a to a peer
    sendMsgReqChan  chan *peerMsg
    // Channel for receive bitcoin message
    receiveMsgChan  chan *peerMsg
    // embed the global IDManager for convenience  
    *IDManager
} 

func NewPool(im *IDManager) *Pool {
    return &Pool {
        peers: make(map[ID]*Peer),
        addReqChan: make(chan *Peer),
        removeReqChan: make(chan ID),
        sendMsgReqChan: make(chan *peerMsg),
        receiveMsgChan: make(chan *peerMsg),
        IDManager: im,
    }
}

// OK to be called simultaneously by multiple goroutines
func (pool *Pool) SendMsg(pmsg *peerMsg) {
    pool.sendMsgReqChan <- pmsg
}

// OK to be called simultaneously by multiple goroutines
func (pool *Pool) RecevieMsg(pmsg *peerMsg) {
    pool.receiveMsgChan <- pmsg
}

func (pool *Pool) Go() {
    go pool.loopManagePeers()
    go pool.loopReceiveMsg()
}

// Member of Monitor interface
func (pool *Pool) InterestedMsgTypes() []string {
    return []string{"1","2"}
}

// Member of Monitor interface
func (pool *Pool) OnPeerUp(p *Peer) {
    pool.addReqChan <- p
}

// Member of Monitor interface
func (pool *Pool) OnPeerDown(p *Peer) {
    pool.removeReqChan <- p.MyID
}

// Member of Monitor interface
func (pool *Pool) OnPeerRecevieMsg(id ID, msg btcmsg.Message) {
    pool.receiveMsgChan <- &peerMsg{id, msg}
}

func (pool *Pool) loopReceiveMsg() {
    for pmsg := range pool.receiveMsgChan {
        logger().Debugf("Got Message, type: %s", pmsg.msg.Command())
    }
}

func (pool *Pool) loopManagePeers() {
    for {
        select {
        case p := <- pool.addReqChan:
            pool.addPeer(p)
        case id := <- pool.removeReqChan:
            pool.removePeer(id)
        case pmsg := <- pool.sendMsgReqChan:
            peer, ok := pool.peers[pmsg.peerID]
            if ok {
                peer.SendMsg(pmsg.msg)
            }else{
                logger().Printf("Can not find peer with ID %i", pmsg.peerID)
            }
        }
    }
}

func (pool *Pool) addPeer(p *Peer) {
    if op, ok := pool.peers[p.MyID]; ok {
        logger().Printf("Waring, Peer already added %v", op.info)    
    }
    pool.peers[p.MyID] = p
    logger().Debugf("+Peer count: %v", len(pool.peers))
}

func (pool *Pool) removePeer(id ID) {
    if pool.peers[id] == nil {
        logger().Printf("Can not find peer with ID %i", id)
        return;
    }
    delete(pool.peers, id)
    logger().Debugf("-Peer count: %v", len(pool.peers))
}
