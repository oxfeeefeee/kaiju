// Sliding window download -- Fast parallel block downloading
package catchUp 

import (
    "sync"
    "time"
    "github.com/oxfeeefeee/kaiju/log"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/catma"
    "github.com/oxfeeefeee/kaiju/blockchain"
    "github.com/oxfeeefeee/kaiju/blockchain/cold"
    "github.com/oxfeeefeee/kaiju/knet"
    "github.com/oxfeeefeee/kaiju/knet/btcmsg"
)

type swdl struct {
    begin       int
    end         int
    paral       int
    load        int
    cursor      int
    size        int
    window      []interface{}
    chout       chan map[int]*blockchain.InvElement
    chin        chan map[int]interface{}
    chblock     chan struct{btcmsg.Message; I int}
    done        chan struct{}
    ca          *swdlca
    wg          sync.WaitGroup
}

func newSwdl(begin int, end int, paral int, load int) *swdl {
    // Open a window that wider than paral * load
    maxSlots := (end - begin) / load
    slots := paral * 4
    if slots > maxSlots {
        slots = maxSlots
    }
    s := slots * load
    if slots == 0 {
        s = end - begin
    }
    log.Infof("newSwdl begin %d end %d winsize %d", begin, end, s)
    return &swdl{
        begin: begin,
        end: end,
        paral: paral,
        load: load,
        cursor: begin,
        size: s,
        window: make([]interface{}, 0),
        chout: make(chan map[int]*blockchain.InvElement),
        chin: make(chan map[int]interface{}),
        chblock: make(chan struct{btcmsg.Message; I int}),
        done: make(chan struct{}),
        ca: newSwdlca(),
    }
}

func (sw *swdl) start() {
    sw.wg.Add(1) // For doSaveBlocks
    for i := 0; i < sw.paral; i++ {
        go sw.doDownload()
    }
    go sw.doSaveBlock()
    go sw.doSchedule()

    sw.chin <- nil // Trigger downloading
    sw.wg.Wait()

    db := cold.Get().OutputDB()
    if err := db.Commit(uint32(sw.end-1),true); err != nil {
        log.Panicf("db commit error: %s", err)
    }
    log.Infof("Finished downloading from %d to %d", sw.begin, sw.end)
}

func (sw *swdl) doSchedule() {
    running := true
    for running {
        select {
        case msgs := <- sw.chin:
            sw.schedule(msgs)
        case <- sw.done:
            running = false
        }   
    }
    close(sw.chblock) // To end doSaveBlock
}

func (sw *swdl) doDownload() {
    running := true
    for running {
        select {
        case req := <- sw.chout:
            if req != nil {
                sw.chin <- download(req)
            } else {
                time.Sleep(30 * time.Second)
                sw.chin <- nil
            }
        case <- sw.done:
            running = false
        }     
    }
    log.Infoln("doDownload exit")
}

func (sw *swdl) doSaveBlock() {
    defer sw.wg.Done()
    for bm := range sw.chblock {
        saveBlock(bm.Message, bm.I, false)
    }
    log.Infoln("doSaveBlock exit")
}

func (sw *swdl) schedule(msgs map[int]interface{}) {
    // 1. Fill blanks with downloaded blocks
    got := 0
    for k, v := range msgs {
        i := k - sw.cursor
        if _, ok := v.(*btcmsg.Message_block); ok {
            got++
        }
        sw.window[i] = v
    }
    // 2. Congestion control
    if len(msgs) > 0 && sw.ca.update(got, len(msgs)) {
        // Send nil for a "NOP" download
        sw.sendWork(nil)
        return
    }
    // 3. Slide window and process blocks
    dist := len(sw.window) // Slide distance
    for i, elem := range sw.window {
        if bm, ok := elem.(*btcmsg.Message_block); ok {
            //log.Infoln("save block", i + sw.cursor, i, sw.cursor, bm.Header.Hash())
            sw.chblock <- struct{btcmsg.Message; I int}{bm, i + sw.cursor}
        } else {
            dist = i
            break
        }
    }
    if dist > 0 {
        log.Infof("swdl: window slided %d", dist)
    }
    sw.window = sw.window[dist:]
    sw.cursor += dist
    l := len(sw.window)
    for i := l; i < sw.size; i++ {
        p := sw.cursor + i
        if p >= sw.end {
            break
        }
        ie := blockchain.GetInvElem(p)
        sw.window = append(sw.window, ie)
    }
    if len(sw.window) == 0 {
        close(sw.done)// All blocks downloaded
    }
    // 4. Handle unfinished work
    unfinished := make(map[int]*blockchain.InvElement)
    for i, elem := range sw.window {
        if ie, ok := elem.(*blockchain.InvElement); ok {
            unfinished[sw.cursor+i] = ie
        }
    }
    l = len(unfinished)
    if len(msgs) > 0 {
        log.Debugf("winsize %d left %d got %d cursor %d health %f", len(sw.window), l, got, sw.cursor, sw.ca.health)
    }
    if l == 0 { // No blocks left, do nothing
    } else if l <= sw.load {
        sw.sendWork(unfinished)
    } else {
        work := make(map[int]*blockchain.InvElement)
        for k, v := range unfinished {
            work[k] = v
            if len(work) >= sw.load {
                if !sw.sendWork(work) { // channel is full
                    break
                }
                work = make(map[int]*blockchain.InvElement)
            }
        }
    }
}

func (sw *swdl) sendWork(w map[int]*blockchain.InvElement) bool {
    select {
    case sw.chout <- w:
        for k, _ := range w {
            sw.window[k-sw.cursor] = nil // nil in window means "in process"
        }
        return true
    default:
        return false
    }
}

func download(req map[int]*blockchain.InvElement) map[int]interface{} {
    inv := make([]*blockchain.InvElement, 0, len(req))
    records := make(map[klib.Hash256]interface{})
    for _, v := range req {
        inv = append(inv, v)
        records[v.Hash] = v
    }
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
    msg := btcmsg.NewGetDataMsg().(*btcmsg.Message_getdata)
    msg.Inventory = inv
    knet.MsgForMsgs(msg, f, len(inv))
    ret := make(map[int]interface{})
    for k, v := range req {
        ret[k] = records[v.Hash] // Either btcmsg.Message or *blockchain.InvElement
    }
    return ret
}

func saveBlock(m btcmsg.Message, i int, verify bool) {
    db := cold.Get().OutputDB()
    bm, _ := m.(*btcmsg.Message_block)
    for _, tx := range bm.Txs {
        ctx := (*catma.Tx)(tx)
        err := catma.VerifyTx(ctx, db, true, false, !verify)
        if err != nil {
            log.Panicf("Process tx %s error: %s", ctx.Hash(), err)
        }
    }
    if err := db.Commit(uint32(i),false); err != nil {
        log.Panicf("db commit error: %s", err)
    }
}

