package knet

import (
    "sync"
    "github.com/oxfeeefeee/kaiju/knet/btcmsg"
)

// Monitors the activities of each peer
type Monitor interface  {
    onPeerUp(p *Peer)
    onPeerDown(p *Peer)
    onPeerMsg(id ID, msg btcmsg.Message)
    listenTypes() []string
}

type monitors struct {
    monitorSlice    []Monitor
    msgReceivers    map[string]Monitor
    mutex           sync.RWMutex
}

func (ms *monitors) onPeerUp(p *Peer) {
    ms.mutex.RLock()
    defer ms.mutex.RUnlock()

    for _, m := range ms.monitorSlice {
        m.onPeerUp(p)
    }
}

func (ms *monitors) onPeerDown(p *Peer) {
    ms.mutex.RLock()
    defer ms.mutex.RUnlock()

    for _, m := range ms.monitorSlice {
        m.onPeerDown(p)
    }
} 

func (ms *monitors) onPeerMsg(id ID, msg btcmsg.Message) {
    ms.mutex.RLock()
    defer ms.mutex.RUnlock()

    m, ok := ms.msgReceivers[msg.Command()]
    if ok {
        m.onPeerMsg(id, msg)
    }
} 

func (ms *monitors) addMonitors(monitors []Monitor) {
    ms.mutex.Lock()
    defer ms.mutex.Unlock()

    for _, m := range monitors {
        ms.monitorSlice = append(ms.monitorSlice, m)
    }
    ms.rebuildReceiverMap()
}

func (ms *monitors) addMonitor(m Monitor) {
    ms.mutex.Lock()
    defer ms.mutex.Unlock()

    ms.monitorSlice = append(ms.monitorSlice, m)
    ms.rebuildReceiverMap()
} 

func (ms *monitors) removeMonitor(monitor Monitor) {
    ms.mutex.Lock()
    defer ms.mutex.Unlock()

    s := ms.monitorSlice
    for i, m := range s {
        if m == monitor {
            s = append(s[:i], s[i+1:]...)
            ms.rebuildReceiverMap()
            return
        }
    }
}

func (ms *monitors) rebuildReceiverMap() {
    //Clear msgReceivers
    ms.msgReceivers = make(map[string]Monitor)
    s := ms.monitorSlice
    for _, m := range s {
        types := m.listenTypes()
        for _, t := range types {
            // TODO: check duplicates and log error
            ms.msgReceivers[t] = m
        }
    }
}

