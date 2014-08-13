// Dogma of Bitcoin, be very careful!
package catma

import (
    "errors"
    "github.com/oxfeeefeee/kaiju"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/catma/script"
)

type UtxoSet interface {
    Get(h *klib.Hash256, i uint32) ([]byte, error)
    Use(h *klib.Hash256, i uint32, val []byte) error
    Add(h *klib.Hash256, i uint32, val []byte) error
}

func VerifyTx(tx *Tx, utxo UtxoSet, preBip16 bool, standard bool) error {
    err := tx.FormatCheck()
    if err != nil {
        return err
    }
    // TODO more checks...

    if !tx.IsCoinBase() {
        //logger().Debugf("yesyesyes %s", tx.Hash())
        for i, txi := range tx.TxIns {
            op := &(txi.PreviousOutput)
            opBytes, err := utxo.Get(&op.Hash, op.Index)
            if err != nil {
                return err
            } else if opBytes == nil {
                return errors.New("VerifyTx: Cannot find input.")
            }
            var txo TxOut
            txo.FromBytes(opBytes)
            err = VerifyInput(txo.PKScript, tx, i, preBip16, standard)
            if err != nil {
                return err
            }
        }
        for _, txi := range tx.TxIns {
            op := &(txi.PreviousOutput)
            err := utxo.Use(&op.Hash, op.Index, nil)
            if err != nil {
                return err
            }
        }
    }
    hash := tx.Hash()
    for i, txo := range tx.TxOuts {
        //logger().Debugf("added OTX %s %d", hash, i)
        err := utxo.Add(hash, uint32(i), txo.Bytes())
        if err != nil {
            return err
        }
    }
    return nil
}

func VerifyInput(pkScript []byte, tx *Tx, idx int, preBip16 bool, standard bool) error {
    var evalFlags script.EvalFlag
    if preBip16 {
        evalFlags = script.EvalFlagNone
    } else if standard {
        evalFlags = 
            script.EvalFlagP2SH | 
            script.EvalFlagStrictEnc | 
            script.EvalFlagNullDummy
    } else {
        evalFlags = script.EvalFlagP2SH
    }
    return VerifyInputWithFlags(pkScript, tx, idx, evalFlags)
}

func VerifyInputWithFlags(pkScript []byte, tx *Tx, idx int, flags script.EvalFlag) error {
    if idx >= len(tx.TxIns) {
        return errors.New("VerifyInput: Input index out of range")
    }
    sig := tx.TxIns[idx].SigScript
    ie := &InputEntry{tx, idx}
    return script.VerifyScript(pkScript, sig, ie, flags)
}

// Handy function
func logger() *kaiju.Logger {
    return kaiju.MainLogger()
}