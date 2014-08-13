package btcmsg

import (
    "bytes"
    )

// Bitcoin protocol message: "pong"
type Message_pong struct {
    Nonce   uint64
}

var pongMsg   *Message_pong

func NewPongMsg() Message {
    if pongMsg == nil {
        pongMsg = new(Message_pong)
    }
    return pongMsg
}

func (m *Message_pong) Command() string {
    return "pong"
}

func (m *Message_pong) Encode() ([]byte, error) {
    buf := new(bytes.Buffer)
    var err error
    err = writeData(buf, &m.Nonce, err)
    return buf.Bytes(), err
}

func (m *Message_pong) Decode(payload []byte) error {
    buf := bytes.NewBuffer(payload)
    var err error
    err = readData(buf, &m.Nonce, err)
    return err
}