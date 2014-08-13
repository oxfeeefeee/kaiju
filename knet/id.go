// Manages ID for all the peers
package knet

import (
    "sync"
    "github.com/oxfeeefeee/kaiju/knet/btcmsg"
)

type ID uint64

var maxNormalID uint64

type idManager struct {
    record      map[btcmsg.PeerIP]ID
    nextID      uint64
    mutex       sync.Mutex
}

func newIDManager() *idManager {
    return &idManager {
        record: make(map[btcmsg.PeerIP]ID),
        nextID: 0,
    }
}

func (m *idManager) getID(ip btcmsg.PeerIP) ID {
    m.mutex.Lock()
    defer m.mutex.Unlock()
    
    id, ok := m.record[ip]
    if !ok {
        id = ID(m.nextID)
        m.nextID++;
        m.record[ip] = id
    }
    return id
}