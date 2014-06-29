// KLIB contains utility functions that are missing from standard library

package klib

import (
    "crypto/sha256"
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju"
    )

func Sha256Sha256(p []byte) *Hash256 {
    h := sha256.Sum256(p)
    hash := Hash256(sha256.Sum256(h[:]))
    return &hash
}

func UInt16ToBytes(i uint16) []byte {
    p := make([]byte, 2)
    binary.LittleEndian.PutUint16(p, i)
    return p
}

func UInt32ToBytes(i uint32) []byte {
    p := make([]byte, 4)
    binary.LittleEndian.PutUint32(p, i)
    return p
}

// Handy function
func logger() *kaiju.Logger {
    return kaiju.KlibLogger
}
