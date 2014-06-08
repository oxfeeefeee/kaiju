package btcmsg

import (
    "bytes"
    "errors"
    "encoding/hex"
    "github.com/oxfeeefeee/kaiju/catma/cst"
)

// Bitcoin protocol message: "alert"
type Message_alert struct {
    Content []byte
}

var alertPubKey []byte

func init() {
    alertPubKey, _ = hex.DecodeString(cst.AlertPublicKey)
} 

func NewAlertMsg() Message {
    return new(Message_alert)
}

func (m *Message_alert) Command() string {
    return "alert"
}

func (m *Message_alert) Encode() ([]byte, error) {
    return []byte{}, nil
}

func (m *Message_alert) Decode(payload []byte) error {
    buf := bytes.NewBuffer(payload)
    var lastError error
    var contentLen, signatureLen VarUint

    lastError = readVarUint(buf, &contentLen, lastError) 
    if lastError != nil {
        return lastError
    } else if contentLen > VarUint(cst.MaxAlertSize) {
        return errors.New("alert content is too long")
    }
    m.Content = make([]byte, contentLen)
    lastError = readData(buf, m.Content, lastError) 
    if lastError != nil {
        return lastError
    }

    lastError = readVarUint(buf, &signatureLen, lastError) 
    if lastError != nil {
        return lastError
    } else if contentLen > VarUint(cst.MaxAlertSingnatureSize) {
        return errors.New("alert singature is too long")
    }
    sing := make([]byte, signatureLen)
    lastError = readData(buf, sing, lastError) 
    if lastError != nil {
        return lastError
    }

    // Now check the singnature
    // TODO
    return nil
}



