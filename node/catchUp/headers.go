package catchUp 

import (
    "time"
    "github.com/oxfeeefeee/kaiju/log"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/knet"
    "github.com/oxfeeefeee/kaiju/knet/btcmsg"
    "github.com/oxfeeefeee/kaiju/blockchain/cold"
)

// Returns if we should stop catching up
func headerUpToDate() bool {
    headers := cold.Get().Headers()
    h := headers.Get(headers.Len() - 1)
    return h.Time().Add(time.Hour * 2).After(time.Now())
}

func moreHeaders() {
    headers := cold.Get().Headers()
    l := headers.GetLocator()
    msg := btcmsg.NewGetHeadersMsg()
    mg := msg.(*btcmsg.Message_getheaders)
    mg.BlockLocators = l
    mg.HashStop = new(klib.Hash256)

    f := func(m btcmsg.Message) (bool, bool) {
        _, ok := m.(*btcmsg.Message_headers)
        return ok, true
    }

    mh := knet.ParalMsgForMsg(mg, f, 3)
    if mh != nil {
        h, _ := mh.(*btcmsg.Message_headers)
        err := headers.Append(h.Headers)
        if err != nil {
            log.Infof("Error appending headers: %s", err)
        }
    }
}
