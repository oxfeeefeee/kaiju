package catma

import (
    "bytes"
    "errors"
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/catma/numbers"
)

type InputEntry struct {
    tx              *Tx
    index           int
}

// Returns the data(Hash of custom serialized tx) to sign for an input of a TX
func (e *InputEntry) HashToSign(subScript []byte, hashType byte) (*klib.Hash256, error) {
    return e.tx.HashToSign(subScript, e.index, hashType)
}

func (t *Tx) HashToSign(subScript []byte, ii int, hashType byte) (*klib.Hash256, error) {
    if ii >= len(t.TxIns) {
        return nil, errors.New("Tx.StringToSign invalid index")
    }
    anyoneCanPay := (hashType & SIGHASH_ANYONECANPAY) != 0
    htype := hashType & numbers.HashTypeMask
    p := new(bytes.Buffer)

    // STEP0: version 
    binary.Write(p, binary.LittleEndian, t.Version)
    // STEP1: inputs 
    if anyoneCanPay {
        // If SIGHASH_ANYONECANPAY is set, only current input is written, 
        // and subScipt is used as SigScript 
        p.Write(klib.VarUint(1).Bytes())                                // inputs count
        binary.Write(p, binary.LittleEndian, t.TxIns[ii].PreviousOutput)// PreviousOutput
        p.Write(((klib.VarString)(subScript)).Bytes())                  // subScript
        binary.Write(p, binary.LittleEndian, t.TxIns[ii].Sequence)      // Sequence
    } else {
        // Else write all the inputs with modifications
        p.Write(klib.VarUint(len(t.TxIns)).Bytes())
        for i, txin := range t.TxIns {
            binary.Write(p, binary.LittleEndian, txin.PreviousOutput)
            if i == ii { // If this is current input, write subScript
                p.Write(((klib.VarString)(subScript)).Bytes())
            } else { // Else write an empty VarString
                p.Write((klib.VarString{}).Bytes())
            }
            sequence := txin.Sequence
            if i != ii && (htype == SIGHASH_NONE || htype == SIGHASH_SINGLE) {
                // If not current input, and of type SIGHASH_NONE || SIGHASH_SINGLE,
                // set sequence to 0
                sequence = 0
            }
            binary.Write(p, binary.LittleEndian, sequence)
        } 
    }
    // STEP3: outputs
    switch htype {
    case SIGHASH_NONE:
        p.Write((klib.VarString{}).Bytes())
    case SIGHASH_SINGLE:
        if ii >= len(t.TxOuts) {
            // This is actually allowed due to a Satoshi Bug, right thing to do:
            // panic("Tx.StringToSign invalid index with type SIGHASH_SINGLE")
            return new(klib.Hash256).SetUint64(1), nil
        }
        p.Write(klib.VarUint(ii + 1).Bytes()) // output count
        for i:=0; i < ii; i++ { // All outputs except the last one are written as blank
            binary.Write(p, binary.LittleEndian, int64(-1)) // value
            p.Write((klib.VarString{}).Bytes()) // script
        }
        txout := t.TxOuts[ii]
        binary.Write(p, binary.LittleEndian, txout.Value)
        p.Write(((klib.VarString)(txout.PKScript)).Bytes())
    default:
        // Another Satoshi Bug: any other hashtype are considered as SIGHASH_ALL
        p.Write(klib.VarUint(len(t.TxOuts)).Bytes())
        for _, txout := range t.TxOuts {
            binary.Write(p, binary.LittleEndian, txout.Value)
            p.Write(((klib.VarString)(txout.PKScript)).Bytes())
        }
    }
    // STEP4: LockTime and HashType
    binary.Write(p, binary.LittleEndian, t.LockTime)
    // Notice hashTypes needs to take 4 bytes
    binary.Write(p, binary.LittleEndian, uint32(hashType))
    return klib.Sha256Sha256(p.Bytes()), nil
}