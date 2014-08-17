package peer

import (
    "sync"
    "errors"
    "github.com/oxfeeefeee/kaiju/knet/btcmsg"
)

// Monitors the activities of each peer
type Monitor interface  {
    OnPeerUp(p *Peer)
    OnPeerDown(p *Peer)
    OnPeerMsg(handle Handle, msg btcmsg.Message)
    ListenTypes() []string
}

type monitors struct {
    monitorSlice    []Monitor
    msgReceivers    map[string]Monitor
    mutex           sync.RWMutex
}

func (ms *monitors) addMonitors(monitors []Monitor) error {
    ms.mutex.Lock()
    defer ms.mutex.Unlock()

    for _, m := range monitors {
        ms.monitorSlice = append(ms.monitorSlice, m)
    }
    ms.rebuildReceiverMap()
    return nil
}

func (ms *monitors) removeMonitor(monitor Monitor) error {
    ms.mutex.Lock()
    defer ms.mutex.Unlock()

    s := ms.monitorSlice
    for i, m := range s {
        if m == monitor {
            s = append(s[:i], s[i+1:]...)
            ms.rebuildReceiverMap()
            return nil
        }
    }
    return errors.New("Did not find monitor to remove")
}

func (ms *monitors) onPeerUp(p *Peer) {
    ms.mutex.RLock()
    defer ms.mutex.RUnlock()

    for _, m := range ms.monitorSlice {
        m.OnPeerUp(p)
    }
}

func (ms *monitors) onPeerDown(p *Peer) {
    ms.mutex.RLock()
    defer ms.mutex.RUnlock()

    for _, m := range ms.monitorSlice {
        m.OnPeerDown(p)
    }
} 

func (ms *monitors) onPeerMsg(handle Handle, msg btcmsg.Message) {
    ms.mutex.RLock()
    defer ms.mutex.RUnlock()

    m, ok := ms.msgReceivers[msg.Command()]
    if ok {
        m.OnPeerMsg(handle, msg)
    }
} 

func (ms *monitors) rebuildReceiverMap() {
    //Clear msgReceivers
    ms.msgReceivers = make(map[string]Monitor)
    s := ms.monitorSlice
    for _, m := range s {
        types := m.ListenTypes()
        for _, t := range types {
            // TODO: check duplicates and log error
            ms.msgReceivers[t] = m
        }
    }
}

