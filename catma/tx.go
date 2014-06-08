package catma

import (
    "github.com/oxfeeefeee/kaiju/klib"
)

type TxOut struct {
    // Transaction Value
    Value           int64
    // pk_script, usually contains the public key as a Bitcoin script setting up conditions to claim this output.
    PKScript        []byte
}

type OutPoint struct {
    // The hash of the referenced transaction.
    Hash            klib.Hash256
    // The index of the specific output in the transaction. The first output is 0, etc.
    Index           uint32
}

type TxIn struct {
    // The previous output transaction reference.
    PreviousOutput  OutPoint
    // Script for confirming transaction authorization.
    SigScript       []byte
    // http://bitcoin.stackexchange.com/questions/2025/what-is-txins-sequence
    Sequence        uint32 
}

type Tx struct {
    Version         uint32
    TxIns           []*TxIn
    TxOuts          []*TxOut
    LockTime        uint32
}