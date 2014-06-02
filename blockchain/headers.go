// Header of blocks, this is where all the bitcoin magic happens:
// - Holds the merkel tree root of the block
// - Proof of work hash happens with headers only
// - All the headers are chained together not the blocks
//
// Headers are with constant size, so the total size is predictable and stored in memory
package blockchain

import (
    "io"
    "os"
    "fmt"
    "time"
    "sync"
    "bytes"
    "errors"
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju/klib"
)

type Header struct {
    // The software version which created this block
    Version         uint32
    // The hash of the previous block
    PrevBlock       klib.Hash256
    // The hash(fingerprint) of all txs in this block
    MerkleRoot      klib.Hash256
    // When this block was created
    Timestamp       uint32
    // The difficulty target being used for this block
    Bits            uint32
    // A random number with fileswhich to compute different hashes when mining.
    Nonce           uint32
}

func (h *Header) Hash() *klib.Hash256 {
    w := new(bytes.Buffer)  
    binary.Write(w, binary.LittleEndian, h)
    return klib.Sha256Sha256(w.Bytes())
}

func (h *Header) Time() time.Time {
    return time.Unix(int64(h.Timestamp), 0)
}

type HChain struct {
    headers     []*Header
    mutex       sync.RWMutex
    file        *os.File
} 

var chainIns *HChain

func Chain() *HChain {
    if chainIns == nil {
        chainIns = &HChain{
            headers :[]*Header{genesisHeader()},
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

// Append newly downloaded headers.
// TODO: This is somewhat broken, malicious remote peers could break this process.
func (hc *HChain) AppendHeaders(hs []*Header) error {
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
        h := new(Header)
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
func (hc *HChain) appendHeader(h *Header) error {
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

func genesisHeader() *Header {
    h := new(Header)
    h.Version = 1
    h.MerkleRoot.SetString("4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b")
    h.Timestamp = 1231006505
    h.Bits = 0x1d00ffff
    h.Nonce = 2083236893
    return h
}

