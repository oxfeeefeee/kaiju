// KLIB contains utility functions that are missing from standard library

package klib

import (
    "bytes"
    "errors"
    "encoding/hex"
    )

// A 256 bit hash, e.g. result of sha256
type Hash256 [32]byte

var zeroHash256 Hash256

// Little endian
func (h *Hash256) SetUint64(v uint64) *Hash256 {
    h[0] = byte(v)
    h[1] = byte(v >> 8)
    h[2] = byte(v >> 16)
    h[3] = byte(v >> 24)
    h[4] = byte(v >> 32)
    h[5] = byte(v >> 40)
    h[6] = byte(v >> 48)
    h[7] = byte(v >> 56)
    return h
}

// MSB first string
func (h *Hash256) SetString(s string) (*Hash256, error) {
    data, err := hex.DecodeString(s)
    if err != nil {
        return nil, err
    } else if (len(data) != 32) {
        return nil, errors.New("Hash256.SetString invalid length.") 
    }
    for i := 0; i<32; i++ {
        h[i] = data[31-i]
    }
    return h, nil
}

func (h *Hash256) String() string {
    data := make([]byte, 32, 32)
    for i := 0; i<32; i++ {
        data[i] = h[31-i]
    }
    return hex.EncodeToString(data)
}

func (h *Hash256) SetZero() {
    for i := range h { h[i] = 0 }
}

func (h *Hash256) IsZero() bool {
    return bytes.Equal(h[:], zeroHash256[:])
}