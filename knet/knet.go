// KNet stands for Kaiju Networking. 
//
// KNet is responsible for interacting with other bitcoin nodes,
// send and receive bitcoin protocol messages.
// KNet can be considered as an communication interface for the "ledger" and "catchUp" module
// to communicate with the rest of the network
package knet

import (
    "errors"
    "github.com/oxfeeefeee/kaiju"
    "github.com/oxfeeefeee/kaiju/knet/peer"
    "github.com/oxfeeefeee/kaiju/knet/btcmsg"
)

// Returns isTheMessageAccepted, used by  MsgForMsgs
type MsgHandler func(btcmsg.Message) bool

type KNet struct {
    cc      *CC
    pm      peer.Manager
}

var instance *KNet

// Start KNet module, should be called before any other calls in knet
func Start(count int) (<-chan struct{}, error) {
    if instance != nil {
        return nil, errors.New("KNet.Start should only be called once")
    }
    pm, err := peer.Init()
    if err != nil {
        return nil, err
    }
    cc := newCC()
    instance = &KNet{cc, pm}
    seeds := kaiju.GetConfig().SeedPeers
    for _, ip := range seeds {
        instance.cc.addSeedAddr(ip, kaiju.ListenPort)
    }
    instance.cc.start([]peer.Monitor{instance.cc})
    return pm.Wait(count), nil
}

// Send a message and expect more than one messages in return
// i.e. getting blocks or txs
func MsgForMsgs(m btcmsg.Message, handler MsgHandler, count int) error {
    handles := instance.pm.Peers(1)
    if len(handles) == 0 {
        return errors.New("Failed to find any remote peers.")
    }
    h := handles[0]
    h.SendMsg(m, 0)
    ch := h.ExpectMsg(
        func(m btcmsg.Message) (bool, bool) {
            return handler(m), count == 0
        }, 0)
    for count > 0 {
        ret := <- ch
        if ret.Error != nil {
            return ret.Error
        }
        count--
    }
    return nil
}

// A faster version of MsgForMsgBlock
// Similar to http://blog.golang.org/go-concurrency-patterns-timing-out-and
// try more and get the fastest one 
func ParalMsgForMsg(m btcmsg.Message, f peer.MsgFilter, paral int) btcmsg.Message {
    handles := instance.pm.Peers(paral)
    ch := make(chan struct{btcmsg.Message; Error error}, paral)
    for _, h := range handles {
        h := h
        go func() {
            select {
            case ch <- h.MsgForMsg(m, f):
            default:
            }
        }()
    }
    for i := 0; i < paral; i ++ {
        me := <-ch
        if me.Error == nil {
            return me.Message
        } else {
            logger().Debugf("ParalMsgForMsg error : %s", me.Error)
        }
    }
    return nil
}

func getHandle(ip btcmsg.PeerIP) peer.Handle {
    return instance.pm.GetHandle(ip)
}

// Handy function
func logger() *kaiju.Logger {
    return kaiju.MainLogger()
}


