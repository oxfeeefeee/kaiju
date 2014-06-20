package blockchain

import (
    "errors"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/log"
    "github.com/oxfeeefeee/kaiju/catma"
)

// Stripped down version of catma.Tx, used to record Txs that are already mined
type TxRecord struct {
    Inputs  []*catma.OutPoint
    Outputs []*catma.TxOut
    Used    []uint32
}

func NewTxRecord() *TxRecord {
    return &TxRecord {
        make([]*catma.OutPoint, 0),
        make([]*catma.TxOut, 0),
        make([]uint32, 0),
    }
}

// The Tx pool containing all the Txs that are mined
type Ledger struct {
    // All the Txs that are not yet put into KDB
    newTxs  map[klib.Hash256]*TxRecord
    // Txs that are already in KDB
    db      *utxoDB
}

func NewLedger() *Ledger {
    return &Ledger{
        make(map[klib.Hash256]*TxRecord),
        newUpspentDB(),
    }
}

// Add Txs into to pool
func (l *Ledger) AddTxs(txs []*catma.Tx) {
    for _, tx := range txs {
        h := tx.Hash()
        r := NewTxRecord()
        for _, txi := range tx.TxIns {
            r.Inputs = append(r.Inputs, &txi.PreviousOutput)
        }
        r.Outputs = tx.TxOuts
        l.newTxs[*h] = r
    }
}

func (l *Ledger) validateInput(txi *catma.TxIn) (int64, error) {
    if tx, ok := l.newTxs[txi.PreviousOutput.Hash]; ok {
        if int(txi.PreviousOutput.Index) >= len(tx.Outputs) {
            return 0, errors.New("Ledger.validateInput invalid output index")
        } else {
            txo := tx.Outputs[txi.PreviousOutput.Index]
            if catma.VerifySig(txo.PKScript, txi.SigScript) {
                return txo.Value, nil
            } else {
                return 0, errors.New("Ledger.validateInput invalid signature")
            }
            
        }
    } else {
        return 0, errors.New("Ledger.validateInput kdb not implemented")
    }
}


// Handy function
func logger() *log.Logger {
    return log.BlockchainLogger
}