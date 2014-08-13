// All the messages in the bitcoin network protocol
package btcmsg

// Bitcoin protocol message: "getaddr"
type Message_getaddr struct { 
    //No content
}

var getaddrMsg  *Message_getaddr

func NewGetAddrMsg() Message {
    if getaddrMsg == nil {
        getaddrMsg = new(Message_getaddr)
    }
    return getaddrMsg
}

func (m *Message_getaddr) Command() string {
    return "getaddr"
}

func (m *Message_getaddr) Encode() ([]byte, error) {
    return []byte{}, nil
}

func (m *Message_getaddr) Decode(payload []byte) error {
    // Nothing needs to be done
    return nil
}