// Data types used in Bitcoin messages
package btcmsg

import (
    "io"
    "errors"
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju/log"
    "github.com/oxfeeefeee/kaiju/catma"
    "github.com/oxfeeefeee/kaiju/catma/cst"
)

// The interface of all the message types that are used in bitcoin network protocol
type Message interface {
    Command() string
    Encode() ([]byte, error)
    Decode(payload []byte) error
}

// Value               StorageLength   Format
// <  0xfd             1               uint8_t
// <= 0xffff           3               0xfd followed by the length as uint16_t
// <= 0xffffffff       5               0xfe followed by the length as uint32_t
// >  0xffffffff       9               0xff followed by the length as uint64_t
type VarUint uint64

// Encoded as a VarUint representing the length of the string, followed by the content of the string
type VarString []byte

type Tx catma.Tx

func writeVarUint(w io.Writer, vuint *VarUint, lastError error) error {
    if lastError != nil {
        return nil
    }
    var buffer [9]byte
    dataLen := 0
    v := *vuint
    switch {
    case v < 0xfd:
        buffer[0] = byte(v)
        dataLen = 1
    case v <= 0xffff:
        buffer[0] = 0xfd
        binary.LittleEndian.PutUint16(buffer[1:], uint16(v))
        dataLen = 3
    case v <= 0xffffffff:
        buffer[0] = 0xfe
        binary.LittleEndian.PutUint32(buffer[1:], uint32(v))
        dataLen = 5
    case v > 0xffffffff:
        buffer[0] = 0xff
        binary.LittleEndian.PutUint64(buffer[1:], uint64(v))
        dataLen = 9
    }
    _, lastError = w.Write(buffer[:dataLen])
    return lastError
}

func readVarUint(r io.Reader, v *VarUint, lastError error) error {
    if lastError != nil {
        return lastError
    }

    oneByteBuf := make([]byte, 1)
    _, lastError = io.ReadFull(r, oneByteBuf)
    if lastError != nil {
        return lastError
    }

    b := oneByteBuf[0]
    switch {
    case b < 0xdf:
        *v = VarUint(b)
    case b == 0xfd:
        twoBytesBuf := make([]byte, 2)
        _, lastError = io.ReadFull(r, twoBytesBuf)
        if lastError != nil {
            return lastError
        }
        *v = VarUint(binary.LittleEndian.Uint16(twoBytesBuf))
    case b == 0xfe:
        fourBytesBuf := make([]byte, 4)
        _, lastError = io.ReadFull(r, fourBytesBuf)
        if lastError != nil {
            return lastError
        }
        *v = VarUint(binary.LittleEndian.Uint32(fourBytesBuf))
    case b == 0xff:
        eightBytesBuf := make([]byte, 8)
        _, lastError = io.ReadFull(r, eightBytesBuf)
        if lastError != nil {
            return lastError
        }
        *v = VarUint(binary.LittleEndian.Uint64(eightBytesBuf))
    }  
    return nil
}

func writeVarString(w io.Writer, p *VarString, lastError error) error {
    if lastError != nil {
        return lastError
    }

    strLen := VarUint(len(*p))
    lastError = writeVarUint(w, &strLen, nil)
    if lastError != nil {
        return lastError
    }

    _, lastError = w.Write(*p)
    return lastError
}

func readVarString(r io.Reader, p *VarString, lastError error) error {
    if lastError != nil {
        return lastError
    }

    var strLen VarUint
    lastError = readVarUint(r, &strLen, nil)
    if lastError != nil {
        return lastError
    } else if strLen > VarUint(cst.MaxStrSize) {
        return errors.New("String too long")
    }

    strBuf := make([]byte, strLen)
    _, lastError = io.ReadFull(r, strBuf)
    if lastError != nil {
        return lastError
    }
    *p = strBuf
    return nil
}

func writeBlockHeader(w io.Writer, bh *catma.Header, lastError error) error {
    if lastError == nil {
        lastError = writeData(w, &bh.Version, lastError)
        lastError = writeData(w, &bh.PrevBlock, lastError)
        lastError = writeData(w, &bh.MerkleRoot, lastError)
        lastError = writeData(w, &bh.Timestamp, lastError)
        lastError = writeData(w, &bh.Bits, lastError)
        lastError = writeData(w, &bh.Nonce, lastError)
        txCount := byte(0)
        lastError = writeData(w, &txCount, lastError)
    }
    return lastError
}

func readBlockHeader(r io.Reader, bh *catma.Header, lastError error) error {
    if lastError == nil {
        lastError = readData(r, &bh.Version, lastError)
        lastError = readData(r, &bh.PrevBlock, lastError)
        lastError = readData(r, &bh.MerkleRoot, lastError)
        lastError = readData(r, &bh.Timestamp, lastError)
        lastError = readData(r, &bh.Bits, lastError)
        lastError = readData(r, &bh.Nonce, lastError)
        var txc byte // Read the unused tx_count
        lastError = readData(r, &txc, lastError)
    }
    return lastError
}

func writeTx(w io.Writer, tx *Tx, lastError error) error {
    if lastError != nil {
        return lastError
    }
    lastError = writeData(w, &tx.Version, lastError)
    var listSize VarUint = VarUint(len(tx.TxIns))
    lastError = writeVarUint(w, &listSize, lastError)
    for _, txin := range tx.TxIns {
        lastError = writeData(w, &txin.PreviousOutput, lastError)
        lastError = writeVarString(w, (*VarString)(&txin.SigScript), lastError)
        lastError = writeData(w, &txin.Sequence, lastError)
    }
    listSize = VarUint(len(tx.TxOuts))
    lastError = writeVarUint(w, &listSize, lastError)
    for _, txout := range tx.TxOuts {
        lastError = writeData(w, &txout.Value, lastError)
        lastError = writeVarString(w, (*VarString)(&txout.PKScript), lastError)
    }
    lastError = writeData(w, &tx.LockTime, lastError)
    return lastError
}

func readTx(r io.Reader, tx *Tx, lastError error) error {
    if lastError != nil {
        return lastError
    }
    lastError = readData(r, &tx.Version, lastError)
    var listSize VarUint
    lastError = readVarUint(r, &listSize, lastError)
    if lastError != nil {
        return lastError
    } else if listSize > VarUint(cst.MaxInvListSize) {
        return errors.New("TxIn list too long")
    }
    tx.TxIns = make([]*catma.TxIn, listSize)
    txins := tx.TxIns
    for i := uint64(0); i < uint64(listSize); i++ {
        txins[i] = new(catma.TxIn)
        lastError = readData(r, &txins[i].PreviousOutput, lastError)
        lastError = readVarString(r, (*VarString)(&txins[i].SigScript), lastError)
        lastError = readData(r, &txins[i].Sequence, lastError)
    }
    lastError = readVarUint(r, &listSize, lastError)
    if lastError != nil {
        return lastError
    } else if listSize > VarUint(cst.MaxInvListSize) {
        return errors.New("TxOut list too long")
    }
    tx.TxOuts = make([]*catma.TxOut, listSize)
    txouts := tx.TxOuts
    for i := uint64(0); i < uint64(listSize); i++ {
        txouts[i] = new(catma.TxOut)
        lastError = readData(r, &txouts[i].Value, lastError)
        lastError = readVarString(r, (*VarString)(&txouts[i].PKScript), lastError)
    }
    lastError = readData(r, &tx.LockTime, lastError)
    return lastError
} 

func writeData(w io.Writer, data interface{}, lastError error) error {
    if lastError == nil {
        lastError = binary.Write(w, binary.LittleEndian, data)
    }
    return lastError
}

func readData(r io.Reader, data interface{}, lastError error) error {
    if lastError == nil {
        lastError = binary.Read(r, binary.LittleEndian, data)
    }
    return lastError
}

// Handy function
func logger() *log.Logger {
    return log.KioMsgLogger
}
