// Data types used in Bitcoin messages
package btcmsg

import (
    "io"
    "errors"
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/catma"
)

// The interface of all the message types that are used in bitcoin network protocol
type Message interface {
    Command() string
    Encode() ([]byte, error)
    Decode(payload []byte) error
}

type Writeable interface {
    Serialize(w io.Writer) error
}

type Readable interface {
    Deserialize(r io.Reader) error
}

type blockHeader catma.Header

func (bh *blockHeader) Serialize(w io.Writer) error {
    lastError := writeData(w, (*catma.Header)(bh), nil)
    txCount := byte(0)
    lastError = writeData(w, &txCount, lastError)
    return lastError
}

func (bh *blockHeader) Deserialize(r io.Reader) error {
    lastError := readData(r, (*catma.Header)(bh), nil)
    var txc byte // Read the unused tx_count
    lastError = readData(r, &txc, lastError)
    return lastError
}

type Tx catma.Tx

func (tx *Tx) Serialize(w io.Writer) error {
    t := (*catma.Tx)(tx)
    _, err := w.Write(t.Bytes())
    return err
}

func (tx *Tx) Deserialize(r io.Reader) error {
    lastError := readData(r, &tx.Version, nil)
    var listSize klib.VarUint
    lastError = readData(r, &listSize, lastError)
    if lastError != nil {
        return lastError
    } else if listSize > klib.VarUint(kaiju.MaxInvListSize) {
        return errors.New("TxIn list too long")
    }
    tx.TxIns = make([]*catma.TxIn, listSize)
    txins := tx.TxIns
    for i := uint64(0); i < uint64(listSize); i++ {
        txins[i] = new(catma.TxIn)
        lastError = readData(r, &txins[i].PreviousOutput, lastError)
        lastError = readData(r, (*klib.VarString)(&txins[i].SigScript), lastError)
        lastError = readData(r, &txins[i].Sequence, lastError)
    }
    lastError = readData(r, &listSize, lastError)
    if lastError != nil {
        return lastError
    } else if listSize > klib.VarUint(kaiju.MaxInvListSize) {
        return errors.New("TxOut list too long")
    }
    tx.TxOuts = make([]*catma.TxOut, listSize)
    txouts := tx.TxOuts
    for i := uint64(0); i < uint64(listSize); i++ {
        txouts[i] = new(catma.TxOut)
        lastError = readData(r, &txouts[i].Value, lastError)
        lastError = readData(r, (*klib.VarString)(&txouts[i].PKScript), lastError)
    }
    lastError = readData(r, &tx.LockTime, lastError)
    return lastError
} 

func writeData(w io.Writer, data interface{}, lastError error) error {
    if lastError != nil {
        return lastError
    }
    if md, ok := data.(Writeable); ok {
        lastError = md.Serialize(w)
    } else {
        lastError = binary.Write(w, binary.LittleEndian, data)
    }
    return lastError
}

func readData(r io.Reader, data interface{}, lastError error) error {
    if lastError != nil {
        return lastError
    }
    if md, ok := data.(Readable); ok {
        lastError = md.Deserialize(r)
    } else {
        lastError = binary.Read(r, binary.LittleEndian, data)
    }
    return lastError
}

// Handy function
func logger() *kaiju.Logger {
    return kaiju.KioMsgLogger
}
