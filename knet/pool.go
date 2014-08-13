// Pool manages all remote bitcoin nodes, represented by knet.Peer's
package knet

import (
    "time"
    "sync"
    "errors"
    "math/rand"
    "github.com/oxfeeefeee/kaiju/knet/btcmsg"
    )

type Pool struct {
    // All Peers
    peers           map[ID]*Peer
    // For peers access sync
    mutex           sync.RWMutex
    // Channel for receive bitcoin message
    receiveMsgChan  chan *receivedMsg
    // embed the global idManager for convenience  
    *idManager
}

type receivedMsg struct {
    peerID      ID
    msg         btcmsg.Message
} 

func newPool(im *idManager) *Pool {
    return &Pool {
        peers: make(map[ID]*Peer),
        receiveMsgChan: make(chan *receivedMsg, 64),
        idManager: im,
    }
}

// Send a btc message to a specific peer
// Pass 0 as timeout for default timeout length
func (pool *Pool) SendMsg(p ID, m btcmsg.Message, timeout time.Duration) <-chan error {
    if timeout <= 0 {
        timeout = defalutSendMsgTimeout
    }
    pmsg := &msgSent{m, timeout, make(chan error, 1)}
    
    pool.mutex.RLock()
    peer, ok := pool.peers[p]
    pool.mutex.RUnlock()
    if ok {
        peer.sendMsg(pmsg)
    }else{
        pmsg.errChan <- errors.New("Peer no longer in pool.")
    }
    return pmsg.errChan
}

// Broadcast a btc message to all connected peers
func (pool *Pool) BroadcastMsg(m btcmsg.Message, incFunc BroadcastInclude, timeout time.Duration) {
    if timeout <= 0 {
        timeout = defalutSendMsgTimeout
    }
    pool.mutex.RLock()
    defer pool.mutex.RUnlock()
    for _, p := range pool.peers {
        if incFunc(p) {
            pmsg := &msgSent{m, timeout, nil}
            p.sendMsg(pmsg)  
        }
    }
}

// Expect a btc message to be sent from peer "p" that matched the filter "f"
func (pool *Pool) ExpectMsg(p ID, f MsgFilter, timeout time.Duration) <-chan struct{btcmsg.Message; Error error} {
    if timeout <= 0 {
        timeout = defalutSendMsgTimeout
    }
    emsg := &msgExpector{f, timeout, make(chan struct{btcmsg.Message; Error error}, 1)}

    pool.mutex.RLock()
    peer, ok := pool.peers[p]
    pool.mutex.RUnlock()
    if ok {
        peer.expectMsg(emsg)
    }else{
        if emsg.retChan != nil {
            emsg.retChan <- struct{btcmsg.Message; Error error}{
                nil, errors.New("Peer no longer in pool.")}
        }
    }
    return emsg.retChan
}

// Returns no more than "count" number of IDs,
// TODO: optimize the speed maybe?
func (pool *Pool) AnyPeers(count int, exclude []ID) []ID {
    excl := make(map[ID]bool)
    if exclude != nil {
        for _, id := range exclude {
            excl[id] = true
        }
    }
    l := make([]ID, 0, count)
    add := func(id ID) bool {
        if len(l) == count {
            return true
        }else {
            if _, ok := excl[id]; !ok {
                l = append(l, id)
            }
        }
        return false
    }

    pc := len(pool.peers)
    if pc == 0 {
        return l
    }
    n := rand.Intn(pc)
    i := n
    pool.mutex.RLock()
    defer pool.mutex.RUnlock()
    // Start with n'th peer
    for id, _ := range pool.peers {
        if i > 0 {
            i -= 1
        } else if add(id) {
            return l
        }
    }
    // Start from the beginning again if the list is not full
    for id, _ := range pool.peers {
        if add(id) {
            return l
        } else if (i >= n) {
            return l
        } 
        i += 1
    }
    return l
}

func (pool *Pool) peerCount() int {
    pool.mutex.RLock()
    l := len(pool.peers)
    pool.mutex.RUnlock()
    return l
}

// Blocks until the number of connnected peers reached "count"
func (pool *Pool) waitPeers(count int) <-chan struct{} {
    ch := make(chan struct{}, 1)
    go func() {
        t := time.NewTicker(time.Second)
        for _ = range t.C {
            if pool.peerCount() >= count {
                t.Stop()
                ch <- struct{}{}
                return
            }
        } 
    }()
    return ch
}

// Member of Monitor interface
func (pool *Pool) listenTypes() []string {
    return []string{"inv", "headers", "block", "tx"}
}

// Member of Monitor interface
func (pool *Pool) onPeerUp(p *Peer) {
    pool.mutex.Lock()
    defer pool.mutex.Unlock()

    if op, ok := pool.peers[p.myID]; ok {
        logger().Printf("Waring, Peer already added %v", op.info)    
    }
    pool.peers[p.myID] = p
    //logger().Debugf("+Peer count: %v", len(pool.peers))
}

// Member of Monitor interface
func (pool *Pool) onPeerDown(p *Peer) {
    pool.mutex.Lock()
    defer pool.mutex.Unlock()

    id := p.myID
    if pool.peers[id] == nil {
        logger().Printf("Can not find peer with ID %i", id)
        return;
    }
    delete(pool.peers, id)
    logger().Debugf("-Peer count: %v", len(pool.peers))
}

// Member of Monitor interface
func (pool *Pool) onPeerMsg(id ID, m btcmsg.Message) {
    //logger().Debugf("Got Message, type: %s", m.Command())
}

