// Header of blocks, this is where all the bitcoin magic happens:
// - Holds the merkel tree root of the block
// - Proof of work hash happens with headers only
// - All the headers are chained together not the blocks
//
// Headers are with constant size, so the total size is predictable and stored in memory
package blockchain

import (
    "bytes"
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
    // A random number with which to compute different hashes when mining.
    Nonce           uint32
}

func (h *Header) Hash() *klib.Hash256 {
    w := new(bytes.Buffer)
    binary.Write(w, binary.LittleEndian, h)
    return klib.Sha256Sha256(w.Bytes())
}

type HChain []*Header

var chain HChain

func Chain() HChain {
    if chain == nil {
        chain = []*Header{genesisHeader()}
    }
    return chain
}

// Locator is a list of hashes of currently downloaded headers,
// Used to show other peers what we have and what are missing. 
func (c HChain) GetLocator() []*klib.Hash256 {
    ind := locatorIndices(len(c) - 1)
    ltor := make([]*klib.Hash256, 0, len(ind))
    for _, v := range ind {
        ltor = append(ltor, c[v].Hash())
    }    
    return ltor
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

