// Manages ID for all the peers
package peer

import (
    "sync"
    "github.com/oxfeeefeee/kaiju/kio/btcmsg"
)

type ID uint64

var maxNormalID uint64

type IDManager struct {
    record      map[btcmsg.PeerIP]ID
    nextID      uint64
    mutex       sync.Mutex
}

func NewIDManager() *IDManager {
    return &IDManager {
        record: make(map[btcmsg.PeerIP]ID),
        nextID: 0,
    }
}

func (m *IDManager) GetID(ip btcmsg.PeerIP) ID {
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