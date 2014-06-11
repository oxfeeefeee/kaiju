// This file contains implementation of "getheaders"
package btcmsg

import (
    "bytes"
    "errors"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/catma/cst"
)

type Message_getheaders struct {
    Version         uint32
    BlockLocators   []*klib.Hash256
    HashStop        *klib.Hash256
} 

func NewGetHeadersMsg() Message {
    return &Message_getheaders{
        cst.ProtocolVersion,
        nil,
        nil,
    }
}

func (m *Message_getheaders) Command() string {
    return "getheaders"
}

func (m *Message_getheaders) Encode() ([]byte, error) {
    buf := new(bytes.Buffer)
    var err error;

    err = writeData(buf, m.Version, err)
    listSize := klib.VarUint(len(m.BlockLocators))
    err = writeData(buf, &listSize, err)
    for _, l := range m.BlockLocators {
        err = writeData(buf, l, err)
    }
    err = writeData(buf, m.HashStop, err)
    if err != nil {
        return nil, err;
    }
    return buf.Bytes(), nil
}

func (m *Message_getheaders) Decode(payload []byte) error {
    buf := bytes.NewBuffer(payload)
    var err error;
    var listSize klib.VarUint;

    err = readData(buf, &m.Version, err)
    err = readData(buf, &listSize, err)
    if err != nil {
        return err
    } else if listSize > klib.VarUint(cst.MaxInvListSize) {
        return errors.New("Message_getheaders/Message_geblocks list too long")
    }

    inv := make([]*klib.Hash256, listSize)
    for i := uint64(0); i < uint64(listSize); i++ {
        inv[i] = new(klib.Hash256)
        err = readData(buf, inv[i], err)
    } 
    m.BlockLocators = inv
    m.HashStop = new(klib.Hash256)
    err = readData(buf, m.HashStop, err)
    return err
}
