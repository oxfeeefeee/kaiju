// Here is how stuff gets downloaded:
// 0. [getblocks ] "getblocks" which will be responded with "inv"
// 1. [inv] "inv" contains the information of what remote peer has
// 2. [getdata] If we find interesting stuff in "inv", we send "getdata" to request them
// 3. [tx/block] Remote peer sends "tx"/"block" in response to "getdata"
//
// "getheaders" are responded with "headers"
package node 

import (
    "time"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/kio"
    "github.com/oxfeeefeee/kaiju/kio/btcmsg"
    "github.com/oxfeeefeee/kaiju/catma"
    "github.com/oxfeeefeee/kaiju/blockchain"
    "github.com/oxfeeefeee/kaiju/blockchain/cold"
)

func catchUp() {
    for !headerUpToDate() {
        moreHeaders()
    }

    moreBlocks()
}

// Returns if we should stop catching up
func headerUpToDate() bool {
    headers := cold.TheHeaders()
    h := headers.Get(headers.Len() - 1)
    return h.Time().Add(time.Hour * 2).After(time.Now())
}

func moreHeaders() {
    headers := cold.TheHeaders()
    l := headers.GetLocator()
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
        err := headers.Append(h.Headers)
        if err != nil {
            logger().Printf("Error appending headers: %s", err)
        }
    }
}

func moreBlocks() bool {
    idx := make([]int, 0)
    for i := 1; i <= 20000; i++ {
        idx = append(idx, i)
    }
    blocks, err := getBlocks(idx)
    if err != nil {
        logger().Debugf("getBlocks error %s", err)
    }
    logger().Debugf("Got %d blocks\n", len(blocks))
    processBlocks(blocks)
    return true
}

func processBlocks(bmsgs []btcmsg.Message) {
    db := cold.TheOutputDB()
    for _, m := range bmsgs {
        bm, _ := m.(*btcmsg.Message_block)
        for _, tx := range bm.Txs {
            err := catma.VerifyTx((*catma.Tx)(tx), db, true, false)
            if err != nil {
                logger().Printf("Process block error: %s", err)
            }
        }
    }
}

func getBlocks(idx []int) ([]btcmsg.Message, error) {
    inv := blockchain.GetInv(idx)
    msg := btcmsg.NewGetDataMsg()
    m := msg.(*btcmsg.Message_getdata)
    m.Inventory = inv
    records := make(map[klib.Hash256]interface{})
    for _, elem := range inv {
        records[elem.Hash] = elem
    }
    count := len(inv)
    for {
        f := func(m btcmsg.Message) bool {
            bmsg, ok := m.(*btcmsg.Message_block)
            if ok {
                hash := bmsg.Header.Hash()
                v, ok := records[*hash]
                if ok {
                    _, ok := v.(*blockchain.InvElement)
                    records[*hash] = m
                    return ok
                }
            }
            return false
        }
        err := kio.MsgForMsgs(m, f, count)
        if err == nil {
            break
        } else {
            invLeft := make([]*blockchain.InvElement, 0, count)
            for _, v := range records {
                elem, ok := v.(*blockchain.InvElement)
                if ok {
                    invLeft = append(invLeft, elem)
                }
            }
            m.Inventory = invLeft
            count = len(invLeft)
            logger().Debugf("COUNT: %d", count)
        }
    }
    // Assemble the result
    var ret []btcmsg.Message
    for _, elem := range inv {
        m := records[elem.Hash].(btcmsg.Message)
        ret = append(ret, m)
    }
    return ret, nil
}

