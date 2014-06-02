package kio

import (
    "sync"
    "container/list"
    "github.com/oxfeeefeee/kaiju/kio/btcmsg"
)

const maxTriesToConnectPeer = 3

type addrStatus struct {
    // How many times we tried to connect to this peer but failed.
    timesFailed int32
    // The time of last failed try
    lastTryTime int64
    // This is exactly the same as btcmsg.PeerInfo.Time
    // We keep this redundant data for faster comparision, pls refer to betterOrEqualAddr
    lastAvailableTime int64
    // ID of PeerInfo
    id ID
}

type addrPoolEntry struct {
    status  *list.Element
    *btcmsg.PeerInfo
}  

// addrPool keeps all addresses, and periodically pick the best 
// address and try to connect to it.
//
// A linked list and a hash map are used to manage addresses:
//      addrStatusQueue: an ordered list to record and manage the status of addresses
//      addresses: a map containing all addresses for easy look up
type addrPool struct {
    addrStatusQueue     *list.List
    addresses           map[ID]*addrPoolEntry
    mutex               sync.Mutex
    // Embed the global idManager for convenience  
    *idManager
}

func newAddrPool(im *idManager) *addrPool {
    return &addrPool{
        addrStatusQueue: list.New(),
        addresses: make( map[ID]*addrPoolEntry),
        idManager: im,
    }
}

// Returns both *btcmsg.PeerInfo and timesFailed
func (p *addrPool) pickOutBest() (*btcmsg.PeerInfo, int32) {
    q := p.addrStatusQueue
    m := p.addresses
    p.mutex.Lock()
    defer p.mutex.Unlock()

    if q.Len() == 0 {
        return nil, 0
    }
    // Get the addrStatus from the front of the list, which should be the best peer we have
    e := q.Front()
    t := e.Value.(*addrStatus).timesFailed
    if t < maxTriesToConnectPeer {
        q.Remove(e)
        as := e.Value.(*addrStatus)
        tf := as.timesFailed
        entry := m[as.id]
        // Now we put it at the back of the queue, so that addAddr doesn't create duplicates
        // with newly received addresses
        as.timesFailed = maxTriesToConnectPeer + 1
        q.PushBack(as)
        return entry.PeerInfo, tf
    }
    return nil, 0
}

// "replace" dectates if we want to update the existing info about the peer or not
func (p *addrPool) addAddr(replace bool, addr *btcmsg.PeerInfo, timesFailed int32, lastTryTime int64) {
    id := p.getID(addr.IP)
    q := p.addrStatusQueue
    m := p.addresses
    
    p.mutex.Lock()
    defer p.mutex.Unlock()
    // First check if it's already in the pool
    entry, ok := m[id]
    if ok {
        if !replace { // We are done here if we dont want to replace
            return
        }
        // Remove from both the map and the queue
        delete(m, id)
        q.Remove(entry.status)
    }

    var element *list.Element
    as := &addrStatus{timesFailed, lastTryTime, int64(addr.Time), id}
    // Do a linear search for the right spot.
    // TODO: binary search if nessary
    for e := q.Front(); e != nil; e = e.Next() {
        if betterOrEqualAddr(as, e.Value.(*addrStatus)) {
            element = q.InsertBefore(as, e)
            break
        }
    }   
    // If addrInfo no better than any addresses we have (including the case of l.Len()==0), 
    // then put it at the end.
    if element == nil {
        element = q.PushBack(as)    
    }
    // Finally put address in p.addresses
    m[id] = &addrPoolEntry{element, addr} 
}

// Compares the quality(how likely we can connect to them) of two addresses
func betterOrEqualAddr(a0 *addrStatus, a1 *addrStatus) bool {
    if a0.timesFailed == a1.timesFailed {
        if a0.timesFailed == 0 {
            // If never failed, the fresher the better.
            // This should be the most common case.
            return a0.lastAvailableTime >= a1.lastAvailableTime
        } else {
            // If ever failed trying, the longer we waited since the failure the better
            return a0.lastTryTime <= a1.lastTryTime
        }
    } else {
        return a0.timesFailed <= a1.timesFailed
    }
}

func (p *addrPool) dump() {
    stats := make(map[int32]int32)
    for e := p.addrStatusQueue.Front(); e != nil; e = e.Next() {
        t := e.Value.(*addrStatus).timesFailed
        if c, ok := stats[t]; ok {
            stats[t] = c + 1
        } else {
            stats[t] = 1
        }
    }
    logger().Printf("AddrPool Dump %v", stats)
}