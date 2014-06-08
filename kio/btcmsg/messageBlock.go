package btcmsg

import (
    "bytes"
    "errors"
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

    err = writeBlockHeader(buf, m.Header, err)
    listSize := VarUint(len(m.Txs))
    err = writeVarUint(buf, &listSize, err)
    for _, t := range m.Txs {
        err = writeTx(buf, t, err)
    }
    if err != nil {
        return nil, err;
    }
    return buf.Bytes(), nil
}

func (m *Message_block) Decode(payload []byte) error {
    buf := bytes.NewBuffer(payload)
    var err error;
    var listSize VarUint;

    err = readBlockHeader(buf, m.Header, err)
    err = readVarUint(buf, &listSize, err)
    if err != nil {
        return err
    } else if listSize > VarUint(cst.MaxInvListSize) {
        return errors.New("Message_block list too long")
    }

    txs := make([]*Tx, listSize)
    for i := uint64(0); i < uint64(listSize); i++ {
        txs[i] = new(Tx)
        err = readTx(buf, txs[i], err)
    } 
    m.Txs = txs
    return err
}