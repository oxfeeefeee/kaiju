package catma

import (
    "bytes"
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju/klib"
)

const (
    SIGHASH_ALL             = 1
    SIGHASH_NONE            = 2
    SIGHASH_SINGLE          = 3
    SIGHASH_ANYONECANPAY    = 0x80 // This is not a type but a flag
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

func (t *Tx) Hash() *klib.Hash256 {
    return klib.Sha256Sha256(t.Bytes())
}

func (t *Tx) Bytes() []byte {
    p := new(bytes.Buffer)
    binary.Write(p, binary.LittleEndian, t.Version)

    p.Write(klib.VarUint(len(t.TxIns)).Bytes())
    for _, txin := range t.TxIns {
        binary.Write(p, binary.LittleEndian, txin.PreviousOutput)
        p.Write(((*klib.VarString)(&txin.SigScript)).Bytes())
        binary.Write(p, binary.LittleEndian, txin.Sequence)
    }
    p.Write(klib.VarUint(len(t.TxOuts)).Bytes())
    for _, txout := range t.TxOuts {
        binary.Write(p, binary.LittleEndian, txout.Value)
        p.Write(((*klib.VarString)(&txout.PKScript)).Bytes())
    }
    binary.Write(p, binary.LittleEndian, t.LockTime)
    return p.Bytes()
}

// Returns the string to sign for an input of a TX, invalid index causes panic
func (t *Tx) HashToSign(subScript []byte, ii int, hashType uint32) *klib.Hash256 {
    if ii >= len(t.TxIns) {
        panic("Tx.StringToSign invalid index")
    }
    anyoneCanPay := (hashType & SIGHASH_ANYONECANPAY) != 0
    htype := hashType & 0x1f
    p := new(bytes.Buffer)

    // STEP0: version 
    binary.Write(p, binary.LittleEndian, t.Version)

    // STEP1: inputs 
    if (anyoneCanPay) != 0 {
        // If SIGHASH_ANYONECANPAY is set, only current input is written, 
        // and subScipt is used as SigScript 
        p.Write(klib.VarUint(1).Bytes())                                // inputs count
        binary.Write(p, binary.LittleEndian, t.TxIns[ii].PreviousOutput)// PreviousOutput
        p.Write(((*klib.VarString)(&subScript)).Bytes())                // subScript
        binary.Write(p, binary.LittleEndian, txin.Sequence)             // Sequence
    } else {
        // Else write all the inputs with modifications
        p.Write(klib.VarUint(len(t.TxIns)).Bytes())
        binary.Write(p, binary.LittleEndian, txin.PreviousOutput)
        for i, txin := range t.TxIns {
            if i == ii { // If this is current input, write subScript
                p.Write(((*klib.VarString)(&subScript)).Bytes())
            } else { // Else write an empty VarString
                p.Write((klib.VarString{0}).Bytes())
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
        p.Write((klib.VarString{0}).Bytes())
    case SIGHASH_SINGLE:
        if ii >= len(t.TxOuts) {
            // This is actually allowed due to a bug in Satoshi client, should do this:
            // panic("Tx.StringToSign invalid index with type SIGHASH_SINGLE")
            return new(klib.Hash256).SetUint64(1)
        }
        p.Write(klib.VarUint(ii + 1).Bytes()) // output count
        for i:=0; i < ii; i++ { // All outputs except the last one are written as blank
            binary.Write(p, binary.LittleEndian, int64(-1)) // value
            p.Write((klib.VarString{0}).Bytes()) // script
        }
        txout := t.TxOuts[ii]
        binary.Write(p, binary.LittleEndian, txout.Value)
        p.Write(((*klib.VarString)(&txout.PKScript)).Bytes())
    case SIGHASH_ALL:
        p.Write(klib.VarUint(len(t.TxOuts)).Bytes())
        for _, txout := range t.TxOuts {
            binary.Write(p, binary.LittleEndian, txout.Value)
            p.Write(((*klib.VarString)(&txout.PKScript)).Bytes())
        }
    default:
        panic("Tx.StringToSign invalid hash type")
    }

    // STEP4: LockTime and HashType
    binary.Write(p, binary.LittleEndian, t.LockTime)
    binary.Write(p, binary.LittleEndian, hashType)
    return klib.Sha256Sha256(p.Bytes())
}








