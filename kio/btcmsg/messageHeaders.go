package btcmsg

import (
    "bytes"
    "errors"
    "github.com/oxfeeefeee/kaiju/catma"
    "github.com/oxfeeefeee/kaiju/catma/cst"
)

// Bitcoin protocol message: "headers"
type Message_headers struct {
    Headers       []*catma.Header
}

func NewHeadersMsg() Message {
    return &Message_headers{
        make([]*catma.Header, 0),
    }
}

func (m *Message_headers) Command() string {
    return "headers"
}

func (m *Message_headers) Encode() ([]byte, error) {
    buf := new(bytes.Buffer)
    var err error;

    listSize := VarUint(len(m.Headers))
    err = writeVarUint(buf, &listSize, err)
    for _, h := range m.Headers {
        err = writeBlockHeader(buf, h, err)
    }
    if err != nil {
        return nil, err;
    }
    return buf.Bytes(), nil
}

func (m *Message_headers) Decode(payload []byte) error {
    buf := bytes.NewBuffer(payload)
    var err error;
    var listSize VarUint;

    err = readVarUint(buf, &listSize, err)
    if err != nil {
        return err
    } else if listSize > VarUint(cst.MaxInvListSize) {
        return errors.New("Message_headers list too long")
    }

    bhs := make([]*catma.Header, listSize)
    for i := uint64(0); i < uint64(listSize); i++ {
        bhs[i] = new(catma.Header)
        err = readBlockHeader(buf, bhs[i], err)
    } 
    m.Headers = bhs
    return err
}