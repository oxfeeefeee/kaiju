// Pool manages all remote bitcoin nodes, represented by knet.peer.Peer's
package knet

import (
    "time"
    "sync"
    "errors"
    "math/rand"
    "github.com/oxfeeefeee/kaiju/knet/peer"
    "github.com/oxfeeefeee/kaiju/knet/btcmsg"
    )

type Pool struct {
    // All Peers
    peers           map[peer.ID]*peer.Peer
    // For peers access sync
    mutex           sync.RWMutex
    // Channel for receive bitcoin message
    receiveMsgChan  chan *receivedMsg
    // embed the global idManager for convenience  
    *idManager
}

type receivedMsg struct {
    peerID      peer.ID
    msg         btcmsg.Message
} 

func newPool(im *idManager) *Pool {
    return &Pool {
        peers: make(map[peer.ID]*peer.Peer),
        receiveMsgChan: make(chan *receivedMsg, 64),
        idManager: im,
    }
}

// Send a btc message to a specific peer
// Pass 0 as timeout for default timeout length
func (pool *Pool) SendMsg(id peer.ID, m btcmsg.Message, timeout time.Duration) <-chan error {
    ch := make(chan error, 1)
    pool.mutex.RLock()
    p, ok := pool.peers[id]
    pool.mutex.RUnlock()
    if ok {
        p.SendMsg(m, timeout, ch)
    }else{
        ch <- errors.New("peer.Peer no longer in pool.")
    }
    return ch
}

// Broadcast a btc message to all connected peers
func (pool *Pool) BroadcastMsg(m btcmsg.Message, incFunc BroadcastInclude, timeout time.Duration) {
    pool.mutex.RLock()
    defer pool.mutex.RUnlock()
    for _, p := range pool.peers {
        if incFunc(p) {
            p.SendMsg(m, timeout, nil)
        }
    }
}

// Expect a btc message to be sent from peer "p" that matched the filter "f"
func (pool *Pool) ExpectMsg(p peer.ID, f peer.MsgFilter, timeout time.Duration) <-chan struct{btcmsg.Message; Error error} {
    ch := make(chan struct{btcmsg.Message; Error error}, 1)
    pool.mutex.RLock()
    peer, ok := pool.peers[p]
    pool.mutex.RUnlock()
    if ok {
        peer.ExpectMsg(f, timeout, ch)
    }else{
        ch <- struct{btcmsg.Message; Error error}{
            nil, errors.New("peer.Peer no longer in pool.")}
    }
    return ch
}

// Returns no more than "count" number of IDs,
// TODO: optimize the speed maybe?
func (pool *Pool) AnyPeers(count int, exclude []peer.ID) []peer.ID {
    excl := make(map[peer.ID]bool)
    if exclude != nil {
        for _, id := range exclude {
            excl[id] = true
        }
    }
    l := make([]peer.ID, 0, count)
    add := func(id peer.ID) bool {
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
func (pool *Pool) ListenTypes() []string {
    return []string{"inv", "headers", "block", "tx"}
}

// Member of Monitor interface
func (pool *Pool) OnPeerUp(p *peer.Peer) {
    pool.mutex.Lock()
    defer pool.mutex.Unlock()

    if op, ok := pool.peers[p.ID()]; ok {
        logger().Printf("Warning, peer.Peer already added %v", op.BtcInfo())    
    }
    pool.peers[p.ID()] = p
    //logger().Debugf("+peer.Peer count: %v", len(pool.peers))
}

// Member of Monitor interface
func (pool *Pool) OnPeerDown(p *peer.Peer) {
    pool.mutex.Lock()
    defer pool.mutex.Unlock()

    id := p.ID()
    if pool.peers[id] == nil {
        logger().Printf("Can not find peer with peer.ID %i", id)
        return;
    }
    delete(pool.peers, id)
    logger().Debugf("-peer.Peer count: %v", len(pool.peers))
}

// Member of Monitor interface
func (pool *Pool) OnPeerMsg(id peer.ID, m btcmsg.Message) {
    //logger().Debugf("Got Message, type: %s", m.Command())
}

