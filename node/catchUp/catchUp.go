// Here is how stuff gets downloaded:
// 0. [getblocks ] "getblocks" which will be responded with "inv"
// 1. [inv] "inv" contains the information of what remote peer has
// 2. [getdata] If we find interesting stuff in "inv", we send "getdata" to request them
// 3. [tx/block] Remote peer sends "tx"/"block" in response to "getdata"
//
// "getheaders" are responded with "headers"
package catchUp 

import (
    "github.com/oxfeeefeee/kaiju/log"
    "github.com/oxfeeefeee/kaiju/blockchain/cold"
)

func CatchUp() {
    headersCatchUp()

    blocksCatchUp()
}

func headersCatchUp() {
    for !headerUpToDate() {
        moreHeaders()
    }
}

func blocksCatchUp() {
    total := cold.Get().Headers().Len()
    db := cold.Get().OutputDB()
    for {
        tag, err := db.Tag()
        if err != nil {
            log.Panicf("Error reading OutputDB tag: %s", err)
        }
        begin := int(tag) + 1
        if begin >= total {
            break
        }
        end, paral, load := swdlParam(begin, total)
        dl := newSwdl(begin, end, paral, load)
        dl.start()      
    }
    log.Infoln("Block downloading done.")
}

func swdlParam(begin int, total int) (int, int, int) {
    switch {
    case begin < 120001:
        return 120001, 200, 100
    case begin < 180001:
        return 180001, 200, 10
    default:
        return total, 200, 1
    }
}
