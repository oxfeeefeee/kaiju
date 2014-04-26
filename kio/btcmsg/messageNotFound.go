// This file contains implementation of "notfound"
package btcmsg

// The content of Message_notfound is the same as Message_inv'
type Message_notfound Message_inv

func NewNotFoundMsg() Message {
    var msg Message = NewInvMsg()
    return msg.(*Message_notfound)
}

func (m *Message_notfound) Command() string {
    return "notfound"
}

func (m *Message_notfound) Encode() ([]byte, error) {
    var msg Message = m
    return msg.(*Message_inv).Encode()
}

func (m *Message_notfound) Decode(payload []byte) error {
    var msg Message = m
    return msg.(*Message_inv).Decode(payload)
}
