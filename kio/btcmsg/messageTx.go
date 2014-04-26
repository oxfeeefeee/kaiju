package btcmsg

import (
    "bytes"
)

// Bitcoin protocol message: "tx"
type Message_tx struct {
    Content     Tx       
}

func NewTxMsg() Message {
    return &Message_tx{
    }
}

func (m *Message_tx) Command() string {
    return "tx"
}

func (m *Message_tx) Encode() ([]byte, error) {
    buf := new(bytes.Buffer)
    var err error;
    err = writeTx(buf, &m.Content, err)
    if err != nil {
        return nil, err;
    }
    return buf.Bytes(), nil
}

func (m *Message_tx) Decode(payload []byte) error {
    buf := bytes.NewBuffer(payload)
    var err error;
    return readTx(buf, &m.Content, err)
}
