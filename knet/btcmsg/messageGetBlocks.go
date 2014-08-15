// This file contains implementation of "getblocks"
package btcmsg

// The content of Message_getblocks is the same as Message_getblocks'
type Message_getblocks Message_getheaders

func NewGetBlocksMsg() Message {
    var msg Message = NewGetHeadersMsg()
    return (*Message_getblocks)(msg.(*Message_getheaders))
}

func (m *Message_getblocks) Command() string {
    return "getblocks"
}

func (m *Message_getblocks) Encode() ([]byte, error) {
    msg := (*Message_getheaders)(m)
    return msg.Encode()
}

func (m *Message_getblocks) Decode(payload []byte) error {
    msg := (*Message_getheaders)(m)
    return msg.Decode(payload)
}