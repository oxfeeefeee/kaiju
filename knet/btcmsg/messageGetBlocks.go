// This file contains implementation of "getblocks"
package btcmsg

// The content of Message_getblocks is the same as Message_getblocks'
type Message_getblocks Message_getheaders

func NewGetBlocksMsg() Message {
    var msg Message = NewGetHeadersMsg()
    return msg.(*Message_getblocks)
}

func (m *Message_getblocks) Command() string {
    return "getblocks"
}

func (m *Message_getblocks) Encode() ([]byte, error) {
    var msg Message = m
    return msg.(*Message_getheaders).Encode()
}

func (m *Message_getblocks) Decode(payload []byte) error {
    var msg Message = m
    return msg.(*Message_getheaders).Decode(payload)
}