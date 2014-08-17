package peer

import (
    "time"
    "sync"
    "errors"
    "math/rand"
    "github.com/oxfeeefeee/kaiju/knet/btcmsg"
)

type Manager interface {
    // Returns Handle for given IP
    GetHandle(ip btcmsg.PeerIP) Handle
    // Returns peer count
    Count() int
    // Blocks until the number of connnected peers reached "count"
    Wait(count int) <-chan struct{}
    // Returns no more than "count" number of Handles
    Peers(count int) []Handle
}

type peerManager struct {
    handles     map[btcmsg.PeerIP]Handle 
    peers       map[Handle]*Peer
    nextHandle  Handle
    mutex       sync.RWMutex
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

func (m *peerManager) Peers(count int) []Handle {
    l := make([]Handle, 0, count)
    add := func(id Handle) bool {
        if len(l) == count {
            return true
        }else {
            l = append(l, id)
        }
        return false
    }

    if m.Count() == 0 {
        return l
    }
    n := rand.Intn(m.Count())
    i := n
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    // Start with n'th peer
    for id, _ := range m.peers {
        if i > 0 {
            i -= 1
        } else if add(id) {
            return l
        }
    }
    // Start from the beginning again if the list is not full
    for id, _ := range m.peers {
        if add(id) {
            return l
        } else if (i >= n) {
            return l
        } 
        i += 1
    }
    return l
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
        // PRINT ERROR
    }
    delete(m.peers, h) 
}