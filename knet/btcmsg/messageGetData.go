// This file contains implementation of "getdata"
package btcmsg

// The content of Message_getdata is the same as Message_inv'
type Message_getdata Message_inv

func NewGetDataMsg() Message {
    msg := NewInvMsg()
    return (*Message_getdata)(msg.(*Message_inv))
}

func (m *Message_getdata) Command() string {
    return "getdata"
}

func (m *Message_getdata) Encode() ([]byte, error) {
    return ((*Message_inv)(m)).Encode()
}

func (m *Message_getdata) Decode(payload []byte) error {
    return ((*Message_inv)(m)).Decode(payload)
}
