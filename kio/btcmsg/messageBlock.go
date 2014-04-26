package btcmsg

import (
    "bytes"
    "errors"
    "github.com/oxfeeefeee/kaiju/cst"
)

// Bitcoin protocol message: "block"
type Message_block struct {
    Header      *BlockHeader
    Txns        []*Tx        
}

func NewBlockMsg() Message {
    return &Message_block{
    }
}

func (m *Message_block) Command() string {
    return "block"
}

func (m *Message_block) Encode() ([]byte, error) {
    buf := new(bytes.Buffer)
    var err error;

    err = writeBlockHeader(buf, m.Header, err)
    listSize := VarUint(len(m.Txns))
    err = writeVarUint(buf, &listSize, err)
    for _, t := range m.Txns {
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

    txns := make([]*Tx, listSize)
    for i := uint64(0); i < uint64(listSize); i++ {
        txns[i] = new(Tx)
        err = readData(buf, txns[i], err)
    } 
    m.Txns = txns
    return err
}