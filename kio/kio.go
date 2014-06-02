// KIO stands for Kaiju IO. 
//
// KIO is responsible for interacting with other bitcoin nodes,
// send and receive bitcoin protocol messages.
// KIO can be considered as an communication interface for the "ledger" and "catchUp" module
// to communicate with the rest of the network
package kio

import (
    "time"
    "github.com/oxfeeefeee/kaiju/config"
    "github.com/oxfeeefeee/kaiju/log"
    "github.com/oxfeeefeee/kaiju/cst"
    "github.com/oxfeeefeee/kaiju/kio/btcmsg"
)

const defalutSendMsgTimeout = time.Second * 10

type BroadcastInclude func (*Peer) bool

// Returns (isTheMessageSwallowed, shouldWeStopExpectingMessage)
type MsgFilter func(btcmsg.Message) (bool, bool)

type KIO struct {
    pool    *Pool
    cc      *CC
    idmgr   *idManager
}

var instance *KIO

// Start KIO module, should be called before any other calls in kio
func Start(count int) <-chan struct{} {
    if instance != nil {
        panic("Start should only be called once")
    }
    idm := newIDManager()
    p := newPool(idm)
    cc := newCC(idm)
    instance = &KIO{p, cc, idm}
    seeds := config.GetConfig().SeedPeers
    for _, ip := range seeds {
        instance.cc.addSeedAddr(ip, cst.ListenPort)
    }
    instance.cc.start([]Monitor{instance.pool, instance.cc})
    return p.waitPeers(count)
}

// Send a message and expect another message in response. e.g. [getheaders -> headers].
// This is a non-blocking call.
func MsgForMsg(id ID, m btcmsg.Message, f MsgFilter) <-chan struct{btcmsg.Message; Error error} {
    p := PeerPool()
    p.SendMsg(id, m, 0)
    return p.ExpectMsg(id, f, 0)
}

// Send a message and expect another message in response. e.g. [getheaders -> headers].
// This is a blocking call.
func MsgForMsgBlock(id ID, m btcmsg.Message, f MsgFilter) struct{btcmsg.Message; Error error} {
    p := PeerPool()
    err := <- p.SendMsg(id, m, 0)
    if err != nil {
        return struct{btcmsg.Message; Error error}{nil, err}
    }
    msg := <- p.ExpectMsg(id, f, 0)
    return msg
}

// A faster version of MsgForMsgBlock
// Similar to http://blog.golang.org/go-concurrency-patterns-timing-out-and
// try more and get the fastest one 
func ParalMsgForMsg(mg btcmsg.Message, f MsgFilter, paral int) btcmsg.Message {
    p := PeerPool()
    ids := p.AnyPeers(paral, nil)
    ch := make(chan struct{btcmsg.Message; Error error}, paral)
    for _, id := range ids {
        id := id
        go func() {
            select {
            case ch <- MsgForMsgBlock(id, mg, f):
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

// Other modules can interact with kio via peer.Pool
func PeerPool() *Pool {
    return instance.pool
}

// Handy function
func logger() *log.Logger {
    return log.KioLogger
}


