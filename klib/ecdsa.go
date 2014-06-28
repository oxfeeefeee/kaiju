package klib

import (
    "math/big"
    "crypto/ecdsa"
    "github.com/conformal/btcec"
    )

//https://bitcointalk.org/index.php?topic=162805
type PubKey []byte

// Returns public key in golang format
func (k PubKey) GoPubKey() (*ecdsa.PublicKey, error) {
    pk, err := btcec.ParsePubKey(k, btcec.S256())
    return (*ecdsa.PublicKey)(pk), err
}

func (k PubKey) IsCanonical() bool {
    return true
}

type Sig []byte

// Returns signature in golang format
func (s Sig) GoSig() (*big.Int, *big.Int, error) {
    sig, err := btcec.ParseSignature(s, btcec.S256())
    return sig.R, sig.S, err 
}

// Returns HashType encoded
func (s Sig) HashType() byte {
    return s[len(s) - 1]
}

func (s Sig) IsCanonical() bool {
    return true
}   