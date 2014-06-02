// Here is how stuff gets downloaded:
// 0. [getblocks ] "getblocks" which will be responded with "inv"
// 1. [inv] "inv" contains the information of what remote peer has
// 2. [getdata] If we find interesting stuff in "inv", we send "getdata" to request them
// 3. [tx/block] Remote peer sends "tx"/"block" in response to "getdata"
//
// "getheaders" are responded with "headers"
package brain 

import (
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/kio"
    "github.com/oxfeeefeee/kaiju/kio/btcmsg"
    "github.com/oxfeeefeee/kaiju/blockchain"
)

func catchUp() {
    for moreHeaders() {}
    //for moreBlocks() {}
}

func moreHeaders() bool {
    c := blockchain.Chain()
    l := c.GetLocator()
    msg := btcmsg.NewGetHeadersMsg()
    mg := msg.(*btcmsg.Message_getheaders)
    mg.BlockLocators = l
    mg.HashStop = new(klib.Hash256)

    f := func(m btcmsg.Message) (bool, bool) {
        _, ok := m.(*btcmsg.Message_headers)
        return ok, true
    }

    mh := kio.ParalMsgForMsg(mg, f, 3)
    if mh != nil {
        h, _ := mh.(*btcmsg.Message_headers)
        err := c.AppendHeaders(h.Headers)
        if err != nil {
            logger().Printf("Error appending headers: %s", err)
        }
    }
    return !c.UpToDate()
}

func moreBlocks() bool {
    return true
}