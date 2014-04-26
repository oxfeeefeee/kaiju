// All the messages in the bitcoin network protocol
package btcmsg

import (
    "bytes"
    "errors"
    "github.com/oxfeeefeee/kaiju/cst"
)

// Bitcoin protocol message: "inv"
type Message_inv struct {
    Inventory       []*InvElement
}

func NewInvMsg() Message {
    return &Message_inv{
    }
}

func (m *Message_inv) Command() string {
    return "inv"
}

func (m *Message_inv) Encode() ([]byte, error) {
    buf := new(bytes.Buffer)
    var err error;
    
    listSize := VarUint(len(m.Inventory))
    err = writeVarUint(buf, &listSize, err)
    for _, e := range m.Inventory {
        err = writeData(buf, e, err)
    }
    if err != nil {
        return nil, err;
    }
    return buf.Bytes(), nil
}

func (m *Message_inv) Decode(payload []byte) error {
    buf := bytes.NewBuffer(payload)
    var err error;
    var listSize VarUint;

    err = readVarUint(buf, &listSize, err)
    if err != nil {
        return err
    } else if listSize > VarUint(cst.MaxInvListSize) {
        return errors.New("Message_inv list too long")
    }

    inv := make([]*InvElement, listSize)
    for i := uint64(0); i < uint64(listSize); i++ {
        inv[i] = new(InvElement)
        err = readData(buf, inv[i], err)
    } 
    m.Inventory = inv
    return err
}