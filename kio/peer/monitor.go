package peer

import (
    "sync"
    "github.com/oxfeeefeee/kaiju/kio/btcmsg"
)

// Monitors the activities of each peer
type Monitor interface  {
	OnPeerUp(p *Peer)
	OnPeerDown(p *Peer)
	OnPeerRecevieMsg(id ID, msg btcmsg.Message)
	InterestedMsgTypes() []string
}

type monitors struct {
	monitorSlice 	[]Monitor
	msgReceivers	map[string]Monitor
	mutex 			sync.Mutex
}

func (ms *monitors) OnPeerUp(p *Peer) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	for _, m := range ms.monitorSlice {
		m.OnPeerUp(p)
	}
}

func (ms *monitors) OnPeerDown(p *Peer) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	for _, m := range ms.monitorSlice {
		m.OnPeerDown(p)
	}
} 

func (ms *monitors) OnPeerRecevieMsg(id ID, msg btcmsg.Message) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	m, ok := ms.msgReceivers[msg.Command()]
	if ok {
		m.OnPeerRecevieMsg(id, msg)
	}
} 

func (ms *monitors) AddMonitors(monitors []Monitor) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	for _, m := range monitors {
		ms.monitorSlice = append(ms.monitorSlice, m)
	}
	ms.rebuildReceiverMap()
}

func (ms *monitors) AddMonitor(m Monitor) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	ms.monitorSlice = append(ms.monitorSlice, m)
	ms.rebuildReceiverMap()
} 

func (ms *monitors) RemoveMonitor(monitor Monitor) {
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
		types := m.InterestedMsgTypes()
		for _, t := range types {
			// TODO: check duplicates and log error
			ms.msgReceivers[t] = m
		}
	}
}

