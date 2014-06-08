// Dogma of Bitcoin.
package catma

import (
    "fmt"
    "time"
    "bytes"
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju/klib"
)

// Header of blocks, this is where all the bitcoin magic happens:
// - Holds the merkel tree root of the block
// - Proof of work hash happens with headers only
// - All the headers are chained together not the blocks
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

func (h *Header) String() string {
    return fmt.Sprintf("<Block Header> Hash: %s, Time: %s", h.Hash(), h.Time())
}
