package peer

import (
    "time"
    "sync"
    "errors"
    "github.com/oxfeeefeee/kaiju/log"
    "github.com/oxfeeefeee/kaiju/knet/btcmsg"
)

type Manager interface {
    // Returns Handle for given IP
    GetHandle(ip btcmsg.PeerIP) Handle
    // Returns peer count
    Count() int
    // Blocks until the number of connnected peers reached "count"
    Wait(count int) <-chan struct{}
    // Exclusively get a handle of a peer
    Borrow() Handle
    // Return a borrowed handle
    Return(h Handle)
}

type peerManager struct {
    handles     map[btcmsg.PeerIP]Handle 
    nextHandle  Handle
    peers       map[Handle]*Peer
    borrowed    map[Handle]bool
    mutex       sync.RWMutex
    bmutex      sync.RWMutex
}

var peerMgr *peerManager

// Perfer explicite initialization.
// So that you have better control over what happens when
func Init() (Manager, error) {
    if peerMgr != nil {
        return nil, errors.New("peer.Init called before")
    }
    peerMgr = &peerManager {
        handles: make(map[btcmsg.PeerIP]Handle),
        peers: make(map[Handle]*Peer),
        borrowed: make(map[Handle]bool),
        nextHandle: 0,
    }
    return peerMgr, nil
}

func (m *peerManager) GetHandle(ip btcmsg.PeerIP) Handle {
    m.mutex.RLock()
    h, ok := m.handles[ip]
    m.mutex.RUnlock()
    if !ok {
        h = m.nextHandle
        m.mutex.Lock()
        defer m.mutex.Unlock()
        m.nextHandle++;
        m.handles[ip] = h
    }
    return h
}

func (m *peerManager) Count() int {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    l := len(m.peers)
    return l
}

func (m *peerManager) Wait(count int) <-chan struct{} {
    ch := make(chan struct{}, 1)
    go func() {
        t := time.NewTicker(time.Second)
        for _ = range t.C {
            if m.Count() >= count {
                t.Stop()
                ch <- struct{}{}
                return
            }
        } 
    }()
    return ch
}

// Exclusively get a handle of a peer
func (m *peerManager) Borrow() Handle {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    m.bmutex.Lock()
    defer m.bmutex.Unlock()
    for h, _ := range m.peers {
        if !m.borrowed[h] {
            return h
        }
    }
    return InvalidHandle
}

// Return a borrowed handle
func (m *peerManager) Return(h Handle) {
    m.bmutex.Lock()
    defer m.bmutex.Unlock()
    delete(m.borrowed, h)
}

func (m *peerManager) getPeer(h Handle) *Peer {
    if h == InvalidHandle {
        return nil
    }
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    return m.peers[h]
}

func (m *peerManager) addPeer(p *Peer) (Handle, error) {
    h := m.GetHandle(p.info.IP)
    m.mutex.Lock()
    defer m.mutex.Unlock()
    if p := m.peers[h]; p != nil {
        return InvalidHandle, errors.New("peerManager.addPeer: Peer with the same IP exists")
    }
    p.handle = h
    m.peers[h] = p
    return h, nil
}

func (m *peerManager) peerDie(h Handle) {
    p := m.peers[h]
    if p == nil {
        log.Errorln("peerManager.peerDie: invalid handle", h)
    }
    delete(m.peers, h)
    log.Debugln("-peer.Peer count: ", len(m.peers), h)
}