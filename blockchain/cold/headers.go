package cold

import (
    "io"
    "os"
    "fmt"
    "sync"
    "errors"
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju/log"
    "github.com/oxfeeefeee/kaiju/catma"
    "github.com/oxfeeefeee/kaiju/klib"
)

type headers struct {
    data    []*catma.Header
    mutex   sync.RWMutex
    file    *os.File
}

func newHeaders(f *os.File) *headers {
    return &headers{
        data :[]*catma.Header{genesisHeader()},
        file : f,
    }
}

func (h *headers) Len() int {
    return len(h.data)
}

// Get the hash of block with height "height"
func (h *headers) Get(height int) *catma.Header {
    if height < 0 || height >= len(h.data) {
        return nil
    }
    return h.data[height]
}

// Append downloaded headers.
// TODO: more robust way of getting old blocks
func (h *headers) Append(hs []*catma.Header) error {
    if len(hs) == 0 {
        return nil
    }
    oldL := len(h.data) - 1 // excluding genesis
    for _, header := range hs {
        err := h.appendHeader(header)
        if err != nil {
            return err
        }
    } 
    // Write to file as well
    f := h.file
    // Caclulate the offset
    offset := int64(binary.Size(hs[0]) * oldL)
    _, err := f.Seek(offset, 0)
    if err != nil {
        return err
    }
    for _, header := range hs {
        err := binary.Write(f, binary.LittleEndian, header)
        if err != nil {
            return err
        }
    }
    log.Infof("Headers total: %v", len(h.data))
    return f.Sync()
}

// Locator is a list of hashes of currently downloaded headers,
// Used to show other peers what we have and what are missing. 
func (h *headers) GetLocator() []*klib.Hash256 {
    h.mutex.RLock()
    defer h.mutex.RUnlock()
    d := h.data
    ind := locatorIndices(len(d) - 1)
    ltor := make([]*klib.Hash256, 0, len(ind))
    for _, v := range ind {
        ltor = append(ltor, d[v].Hash())
    }    
    return ltor
}

// Load block headers saved in file.
// Errors are not returned to caller, simply print a log
func (h *headers) loadHeaders() {
    r := h.file
    for {
        ch := new(catma.Header)
        if err := binary.Read(r, binary.LittleEndian, ch); err == nil {
            if err = h.appendHeader(ch); err != nil {
                log.Infof("Error loading block header: %s", err)
                break;
            }
        } else {
            if err != io.EOF {
                log.Infof("Error reading blcok header file: %s", err)
            }
            break;
        }
    }
    log.Infof("Loaded header count: %v", len(h.data))
}

// Append new block header, the chain always at least has genesis in it.
func (h *headers) appendHeader(ch *catma.Header) error {
    h.mutex.Lock()
    defer h.mutex.Unlock()
    d := h.data
    // Search back if the latest one doesn't match
    for i := len(d) - 1; i >=0; i-- {
        if *(d[len(d)-1].Hash()) == ch.PrevBlock {
            h.data = append(d[:i+1], ch)
            return nil
        }
    }
    return errors.New(fmt.Sprintf(
            "appendHeader: PrevBlock value doesn't match any exsiting header: %s", &(ch.PrevBlock)))
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