// Here is how stuff gets downloaded:
// 0. [getblocks ] "getblocks" which will be responded with "inv"
// 1. [inv] "inv" contains the information of what remote peer has
// 2. [getdata] If we find interesting stuff in "inv", we send "getdata" to request them
// 3. [tx/block] Remote peer sends "tx"/"block" in response to "getdata"
//
// "getheaders" are responded with "headers"
package node 

import (
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/kio"
    "github.com/oxfeeefeee/kaiju/kio/btcmsg"
    "github.com/oxfeeefeee/kaiju/blockchain"
)

func catchUp() {
    for moreHeaders() {}
    for moreBlocks() {}
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
    idx := make([]int, 0)
    for i := 1; i <= 500; i++ {
        idx = append(idx, i)
    }
    logger().Debugf("aaaaaaaa")
    blocks := getBlocks(idx)
    for _, b := range blocks {
        logger().Debugf("%s \n", b.(*btcmsg.Message_block).Header)
    }
    return true
}

func getBlocks(idx []int) []btcmsg.Message {
    c := blockchain.Chain()
    inv := c.GetInv(idx)
    msg := btcmsg.NewGetDataMsg()
    m := msg.(*btcmsg.Message_getdata)
    m.Inventory = inv

    // Make a map of hash->bool, to be used to check if a incomming block is expected
    record := make(map[klib.Hash256]bool)
    for _, e := range inv {
        record[e.Hash]= false
    }
    count := len(idx)
    f := func(m btcmsg.Message) (accept bool, stop bool) {
        accept, stop = false, false
        bmsg, ok := m.(*btcmsg.Message_block)
        if ok {
            hash := bmsg.Header.Hash()
            b, ok := record[*hash]
            if ok { // Is expected block
                accept = true
                if !b { // We didn't get it before
                    record[*hash] = true
                    count -= 1
                    if count <= 0 {
                        stop = true
                    }
                }
            }
        }
        return
    }
    return kio.MsgForMsgs(m, f, len(inv))
}







