// Headers are with constant size, so the total size is predictable and stored in memory
package blockchain

import (
    "io"
    "os"
    "fmt"
    "time"
    "sync"
    "errors"
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju/catma"
    "github.com/oxfeeefeee/kaiju/klib"
)

const (
    InvTypeError = 0
    InvTypeTx = 1
    InvTypeBlock = 2
)

type InvElement struct {
    InvType     uint32
    Hash        klib.Hash256
}

type HChain struct {
    headers     []*catma.Header
    mutex       sync.RWMutex
    file        *os.File
} 

var chainIns *HChain

func Chain() *HChain {
    if chainIns == nil {
        chainIns = &HChain{
            headers :[]*catma.Header{genesisHeader()},
            file : fileHeaders()}
        chainIns.loadHeaders()
    }
    return chainIns
}

// Used to tell if we should stop catching up
func (hc *HChain) UpToDate() bool {
    h := hc.headers[len(hc.headers)-1]
    return h.Time().Add(time.Hour * 2).After(time.Now())
}

// Locator is a list of hashes of currently downloaded headers,
// Used to show other peers what we have and what are missing. 
func (hc *HChain) GetLocator() []*klib.Hash256 {
    hc.mutex.RLock()
    defer hc.mutex.RUnlock()
    c := hc.headers
    ind := locatorIndices(len(c) - 1)
    ltor := make([]*klib.Hash256, 0, len(ind))
    for _, v := range ind {
        ltor = append(ltor, c[v].Hash())
    }    
    return ltor
}

// Get an array of InvElement to make a "getdata" message
func (hc *HChain) GetInv(heights []int) []*InvElement {
    inv := make([]*InvElement, 0)
    for _, h := range heights {
        ele := &InvElement{InvTypeBlock, *(hc.headers[h].Hash())}
        inv = append(inv, ele)
    }
    return inv
}

// Append newly downloaded headers.
// TODO: This is somewhat broken, malicious remote peers could break this process.
func (hc *HChain) AppendHeaders(hs []*catma.Header) error {
    logger().Debugf("Append header count: %v", len(hs))
    if len(hs) == 0 {
        return nil
    }
    oldL := len(hc.headers) - 1 // excluding genesis
    for _, h := range hs {
        err := hc.appendHeader(h)
        if err != nil {
            return err
        }
    } 
    // Write to file as well
    f := hc.file
    // Caclulate the offset
    offset := int64(binary.Size(hs[0]) * oldL)
    _, err := f.Seek(offset, 0)
    if err != nil {
        return err
    }
    for _, h := range hs {
        err := binary.Write(f, binary.LittleEndian, h)
        if err != nil {
            return err
        }
    }
    logger().Printf("Headers total: %v", len(hc.headers))
    return f.Sync()
}

// Load block headers saved in file.
// Errors are not returned to caller, simply print a log
func (hc *HChain) loadHeaders() {
    r := hc.file
    for {
        h := new(catma.Header)
        if err := binary.Read(r, binary.LittleEndian, h); err == nil {
            if err = hc.appendHeader(h); err != nil {
                logger().Printf("Error loading block header: %s", err)
                break;
            }
        } else {
            if err != io.EOF {
                logger().Printf("Error reading blcok header file: %s", err)
            }
            break;
        }
    }
    logger().Printf("Loaded header count: %v", len(hc.headers))
}   

// Append new block header, the chain always has at least genesis in it.
func (hc *HChain) appendHeader(h *catma.Header) error {
    hc.mutex.Lock()
    defer hc.mutex.Unlock()
    c := hc.headers
    for i := len(c) - 1; i >=0; i-- {
        if *(c[len(c)-1].Hash()) == h.PrevBlock {
            hc.headers = append(c[:i+1], h)
            return nil
        }
    }
    return errors.New(fmt.Sprintf(
            "appendHeader: PrevBlock value doesn't match any exsiting header: %s", &(h.PrevBlock)))
}

func locatorIndices(h int) []int {
    l := make([]int, 0, 32)
    step := 1
    for ; h > 0; h -= step {
        if len(l) >= 10 {
            step *= 2
        }
        l = append(l, h)
    }
    // Add genesis block
    l = append(l, 0)
    return l
}

func genesisHeader() *catma.Header {
    h := new(catma.Header)
    h.Version = 1
    h.MerkleRoot.SetString("4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b")
    h.Timestamp = 1231006505
    h.Bits = 0x1d00ffff
    h.Nonce = 2083236893
    return h
}

