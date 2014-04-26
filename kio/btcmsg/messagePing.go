package btcmsg

import (
    "bytes"
    )

// Bitcoin protocol message: "ping"
type Message_ping struct {
    Nonce   uint64
}

var pingMsg   *Message_ping

func NewPingMsg() Message {
    if pingMsg == nil {
        pingMsg = new(Message_ping)
    }
    return pingMsg
}

func (m *Message_ping) Command() string {
    return "ping"
}

func (m *Message_ping) Encode() ([]byte, error) {
    buf := new(bytes.Buffer)
    var err error
    err = writeData(buf, &m.Nonce, err)
    return buf.Bytes(), err
}

func (m *Message_ping) Decode(payload []byte) error {
    buf := bytes.NewBuffer(payload)
    var err error
    err = readData(buf, &m.Nonce, err)
    return err
}