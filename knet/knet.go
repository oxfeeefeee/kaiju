// KNet stands for Kaiju Networking. 
//
// KNet is responsible for interacting with other bitcoin nodes,
// send and receive bitcoin protocol messages.
// KNet can be considered as an communication interface for the "ledger" and "catchUp" module
// to communicate with the rest of the network
package knet

import (
    "time"
    "errors"
    "github.com/oxfeeefeee/kaiju"
    "github.com/oxfeeefeee/kaiju/log"
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

func Peers() peer.Manager {
    return instance.pm
}

// Send a message and expect more than one messages in return
// i.e. getting blocks or txs
func MsgForMsgs(m btcmsg.Message, handler MsgHandler, count int) error {
    h := Peers().Borrow()
    defer Peers().Return(h)
    h.SendMsg(m, 0)
    ch := h.ExpectMsg(
        func(m btcmsg.Message) (bool, bool) {
            if acc := handler(m); !acc {
                return false, false
            } else {
                return true, count <= 1
            }
        }, time.Second * 60)
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
    ch := make(chan struct{btcmsg.Message; Error error}, paral)
    for i := 0; i < paral; i++ {
        h := Peers().Borrow()
        go func() {
            select {
            case ch <- h.MsgForMsg(m, f):
                Peers().Return(h)
            default:
            }
        }()
    }
    for i := 0; i < paral; i ++ {
        me := <-ch
        if me.Error == nil {
            return me.Message
        } else {
            log.Debugf("ParalMsgForMsg error : %s", me.Error)
        }
    }
    return nil
}

func getHandle(ip btcmsg.PeerIP) peer.Handle {
    return instance.pm.GetHandle(ip)
}

