package blockchain

import (
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/log"
    "github.com/oxfeeefeee/kaiju/catma"
)

// Stripped down version of catma.Tx, used to record Txs that are already mined
type TxRecord struct {
    Inputs []catma.OutPoint
    Outputs []catma.TxOut
    Used []uint32
}

func NewTxRecord() *TxRecord {
    return &TxRecord {
        make([]catma.OutPoint, 0),
        make([]catma.TxOut, 0),
        make([]uint32, 0),
    }
}

// The Tx pool containing all the Txs that are mined, but not yet put into KDB
type MinedTxs map[klib.Hash256]*TxRecord

// Add Txs into to pool
func (t *MinedTxs) AddTxs(txs []*catma.Tx) {
    
}

// Handy function
func logger() *log.Logger {
    return log.BlockchainLogger
}