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
    "github.com/oxfeeefeee/kaiju/knet/btcmsg"
)

const defalutSendMsgTimeout = time.Second * 20

type BroadcastInclude func (*Peer) bool

// Returns (isTheMessageSwallowed, shouldWeStopExpectingMessage)
type MsgFilter func(btcmsg.Message) (accept bool, stop bool)

// Returns isTheMessageAccepted, used by  MsgForMsgs
type MsgHandler func(btcmsg.Message) bool

type KNet struct {
    pool    *Pool
    cc      *CC
    idmgr   *idManager
}

var instance *KNet

// Start KNet module, should be called before any other calls in knet
func Start(count int) <-chan struct{} {
    if instance != nil {
        panic("Start should only be called once")
    }
    idm := newIDManager()
    p := newPool(idm)
    cc := newCC(idm)
    instance = &KNet{p, cc, idm}
    seeds := kaiju.GetConfig().SeedPeers
    for _, ip := range seeds {
        instance.cc.addSeedAddr(ip, kaiju.ListenPort)
    }
    instance.cc.start([]Monitor{instance.pool, instance.cc})
    return p.waitPeers(count)
}

// Send a message and expect another message in response. e.g. [getheaders -> headers].
// This is a non-blocking call.
func MsgForMsg(id ID, m btcmsg.Message, f MsgFilter) <-chan struct{btcmsg.Message; Error error} {
    ch := make(chan struct{btcmsg.Message; Error error})
    go func() { ch <- MsgForMsgBlock(id, m, f) } ()
    return ch
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

// Send a message and expect more than one messages in return
// i.e. getting blocks or txs
func MsgForMsgs(m btcmsg.Message, handler MsgHandler, count int) error {
    p := PeerPool()
    ids := p.AnyPeers(1, nil)
    if len(ids) == 0 {
        return errors.New("Failed to find any remote peers.")
    }
    id := ids[0]
    p.SendMsg(id, m, 0)
    ch := p.ExpectMsg(id,
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
func ParalMsgForMsg(m btcmsg.Message, f MsgFilter, paral int) btcmsg.Message {
    p := PeerPool()
    ids := p.AnyPeers(paral, nil)
    ch := make(chan struct{btcmsg.Message; Error error}, paral)
    for _, id := range ids {
        id := id
        go func() {
            select {
            case ch <- MsgForMsgBlock(id, m, f):
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

// Other modules can interact with knet via peer.Pool
func PeerPool() *Pool {
    return instance.pool
}

// Handy function
func logger() *kaiju.Logger {
    return kaiju.MainLogger()
}


