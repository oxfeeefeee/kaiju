// This file contains implementation of "notfound"
package btcmsg

// The content of Message_notfound is the same as Message_inv'
type Message_notfound Message_inv

func NewNotFoundMsg() Message {
    var msg Message = NewInvMsg()
    return (*Message_notfound)(msg.(*Message_inv))
}

func (m *Message_notfound) Command() string {
    return "notfound"
}

func (m *Message_notfound) Encode() ([]byte, error) {
    msg := (*Message_inv)(m)
    return msg.Encode()
}

func (m *Message_notfound) Decode(payload []byte) error {
    msg := (*Message_inv)(m)
    return msg.Decode(payload)
}
