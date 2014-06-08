package blockchain

import (
    "github.com/oxfeeefeee/kaiju/log"
    "github.com/oxfeeefeee/kaiju/catma"
)

type TxBlock struct {
    Height      int
    Txs         []*catma.Tx
}

func (b *TxBlock) Header() *catma.Header {
    return nil
}

// Handy function
func logger() *log.Logger {
    return log.BlockchainLogger
}