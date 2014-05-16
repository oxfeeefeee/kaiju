// KLIB contains utility functions that are missing from standard library

package klib

import (
    "crypto/sha256"
    )

func Sha256Sha256(p []byte) *Hash256 {
    h := new(Hash256)
    sha := sha256.New()
    sha.Write(p)
    copy(h[:], sha.Sum(nil)[:])
    sha.Reset()
    sha.Write(h[:])
    copy(h[:], sha.Sum(nil)[:])
    return h
}