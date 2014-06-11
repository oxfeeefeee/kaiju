package btcmsg

import (
    "bytes"
    "errors"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/catma"
    "github.com/oxfeeefeee/kaiju/catma/cst"
)

// Bitcoin protocol message: "block"
type Message_block struct {
    Header      *catma.Header
    Txs         []*Tx        
}

func NewBlockMsg() Message {
    return &Message_block{
        Header: new(catma.Header),
    }
}

func (m *Message_block) Command() string {
    return "block"
}

func (m *Message_block) Encode() ([]byte, error) {
    buf := new(bytes.Buffer)
    var err error;

    err = writeData(buf, m.Header, err)
    listSize := klib.VarUint(len(m.Txs))
    err = writeData(buf, &listSize, err)
    for _, t := range m.Txs {
        err = writeData(buf, t, err)
    }
    if err != nil {
        return nil, err;
    }
    return buf.Bytes(), nil
}

func (m *Message_block) Decode(payload []byte) error {
    buf := bytes.NewBuffer(payload)
    var err error;
    var listSize klib.VarUint;

    err = readData(buf, m.Header, err)
    err = readData(buf, &listSize, err)
    if err != nil {
        return err
    } else if listSize > klib.VarUint(cst.MaxInvListSize) {
        return errors.New("Message_block list too long")
    }

    txs := make([]*Tx, listSize)
    for i := uint64(0); i < uint64(listSize); i++ {
        txs[i] = new(Tx)
        err = readData(buf, txs[i], err)
    } 
    m.Txs = txs
    return err
}