package peer

import (
    "time"
    "errors"
    "github.com/oxfeeefeee/kaiju/knet/btcmsg"
)

const InvalidHandle Handle = 0

type Handle uint64

var ErrBadHandle = errors.New("Invalid peer handle")

// Send a btc message to a specific peer
// Pass 0 as timeout for default timeout length
func (h Handle) SendMsg(m btcmsg.Message, timeout time.Duration) <-chan error {
    ch := make(chan error, 1)
    p := peerMgr.getPeer(h)
    if p != nil {
        p.sendMsg(m, timeout, ch)
    }else{
        ch <- ErrBadHandle
    }
    return ch
}

// Expect a btc message to be sent from peer "p" that matched the filter "f"
func (h Handle) ExpectMsg(f MsgFilter, timeout time.Duration) <-chan struct{btcmsg.Message; Error error} {
    ch := make(chan struct{btcmsg.Message; Error error}, 1)
    p := peerMgr.getPeer(h)
    if p != nil {
        p.expectMsg(f, timeout, ch)
    }else{
        ch <- struct{btcmsg.Message; Error error}{
            nil, ErrBadHandle}
    }
    return ch
}

// Send a message and expect another message in response. e.g. [getheaders -> headers].
// This is a blocking call.
func (h Handle) MsgForMsg(m btcmsg.Message, f MsgFilter) struct{btcmsg.Message; Error error} {
    err := <- h.SendMsg(m, 0)
    if err != nil {
        return struct{btcmsg.Message; Error error}{nil, err}
    }
    msg := <- h.ExpectMsg(f, 0)
    return msg
}

func (h Handle) Start() error{
    p := peerMgr.getPeer(h)
    if p == nil {
        return ErrBadHandle
    }
    return p.start()
}

func (h Handle) Kill() {
    p := peerMgr.getPeer(h)
    if p != nil {
        p.kill()
    }
}

func (h Handle) AddMonitors(monitors []Monitor) error {
    p := peerMgr.getPeer(h)
    if p == nil {
        return ErrBadHandle
    }
    return p.addMonitors(monitors)
}

func (h Handle) RemoveMonitor(monitor Monitor) error {
    p := peerMgr.getPeer(h)
    if p == nil {
        return ErrBadHandle
    }
    return p.removeMonitor(monitor)
}

