// KLIB contains utility functions that are missing from standard library

package klib

import (
    "errors"
    "encoding/hex"
    )

// A 256 bit hash, e.g. result of sha256
type Hash256 [32]byte

// MSB first string
func (h *Hash256) SetString(s string) error {
    data, err := hex.DecodeString(s)
    if err != nil {
        return err
    } else if (len(data) != 32) {
        return errors.New("Hash256.SetString invalid length.") 
    }
    for i := 0; i<32; i++ {
        h[i] = data[31-i]
    }
    return nil
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
