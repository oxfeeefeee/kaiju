// All the messages in the bitcoin network protocol
package btcmsg

// Bitcoin protocol message: "verack"
type Message_verack struct {
    //No content
}

var verackMsg   *Message_verack

func NewVerAckMsg() Message {
    if verackMsg == nil {
        verackMsg = new(Message_verack)
    }
    return verackMsg
}

func (m *Message_verack) Command() string {
    return "verack"
}

func (m *Message_verack) Encode() ([]byte, error) {
    return []byte{}, nil
}

func (m *Message_verack) Decode(payload []byte) error {
    // Nothing needs to be done
    return nil
}