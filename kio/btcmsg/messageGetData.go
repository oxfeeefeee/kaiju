// This file contains implementation of "getdata"
package btcmsg

// The content of Message_getdata is the same as Message_inv'
type Message_getdata Message_inv

func NewGetDataMsg() Message {
    var msg Message = NewInvMsg()
    return msg.(*Message_getdata)
}

func (m *Message_getdata) Command() string {
    return "getdata"
}

func (m *Message_getdata) Encode() ([]byte, error) {
    var msg Message = m
    return msg.(*Message_inv).Encode()
}

func (m *Message_getdata) Decode(payload []byte) error {
    var msg Message = m
    return msg.(Message).(*Message_inv).Decode(payload)
}
