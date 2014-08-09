package catma

import (
    "bytes"
    //"errors"
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju/klib"
)

const (
    SIGHASH_ALL             byte  = 1
    SIGHASH_NONE            byte  = 2
    SIGHASH_SINGLE          byte  = 3
    SIGHASH_ANYONECANPAY    byte  = 0x80 // This is not a type but a flag
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

func (to *TxOut) IsDust() bool {
    return false // TODO_1
}

func (to *TxOut) Bytes() []byte {
    var p bytes.Buffer
    binary.Write(&p, binary.LittleEndian, to.Value)
    p.Write(to.PKScript)
    return p.Bytes()
}

func (to *TxOut) FromBytes(p []byte)  {
    b := bytes.NewBuffer(p)
    binary.Read(b, binary.LittleEndian, &to.Value)
    to.PKScript = b.Bytes()
}

func (op OutPoint) Equals(op2 OutPoint) bool {
    return bytes.Equal(op.Hash[:], op2.Hash[:]) && op.Index == op2.Index
}

func (op *OutPoint) SetNull() {
    op.Hash.SetZero()
    op.Index = 0xffffffff
}

func (op *OutPoint) IsNull() bool {
    return op.Hash.IsZero() && op.Index == 0xffffffff
}

// Returns the sha256^2 of serialized Tx
func (t *Tx) Hash() *klib.Hash256 {
    return klib.Sha256Sha256(t.Bytes())
}

// Returns the data size of serialized Tx
func (t *Tx) ByteSize() int {
    opLen := 32/*OutPoint.Hash*/ + 4/*OutPoint.Index*/
    totalLen := 4 // Version
    totalLen += klib.VarUint(len(t.TxIns)).ByteSize()
    for _, txin := range t.TxIns {
        totalLen += opLen
        totalLen += klib.VarString(txin.SigScript).ByteSize()
        totalLen += 4 // Sequence
    }
    totalLen += klib.VarUint(len(t.TxOuts)).ByteSize()
    for _, txout := range t.TxOuts {
        totalLen += 8 // Value
        totalLen += klib.VarString(txout.PKScript).ByteSize()
    }
    totalLen += 4 // LockTime
    return totalLen
}

// Returns the serialized byte of the Tx
func (t *Tx) Bytes() []byte {
    p := new(bytes.Buffer)
    binary.Write(p, binary.LittleEndian, t.Version)

    p.Write(klib.VarUint(len(t.TxIns)).Bytes())
    for _, txin := range t.TxIns {
        binary.Write(p, binary.LittleEndian, txin.PreviousOutput)
        p.Write(klib.VarString(txin.SigScript).Bytes())
        binary.Write(p, binary.LittleEndian, txin.Sequence)
    }
    p.Write(klib.VarUint(len(t.TxOuts)).Bytes())
    for _, txout := range t.TxOuts {
        binary.Write(p, binary.LittleEndian, txout.Value)
        p.Write(klib.VarString(txout.PKScript).Bytes())
    }
    binary.Write(p, binary.LittleEndian, t.LockTime)
    return p.Bytes()
}

func (t *Tx) IsCoinBase() bool {
    return len(t.TxIns) == 1 && t.TxIns[0].PreviousOutput.IsNull()
}