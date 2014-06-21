package klib

import (
    "io"
    "errors"
    "github.com/oxfeeefeee/kaiju"
)

// Encoded as a VarUint representing the length of the string, followed by the content of the string
type VarString []byte

func (s VarString) Bytes() []byte {
    data := VarUint(len(s)).Bytes()
    return append(data, []byte(s)...)
}

func (s VarString) Serialize(w io.Writer) error {
    _, err := w.Write(s.Bytes())
    return err
}

func (s *VarString) Deserialize(r io.Reader) error {
    var strLen VarUint
    err := strLen.Deserialize(r)
    if err != nil {
        return err
    } else if strLen > VarUint(kaiju.MaxStrSize) {
        return errors.New("String too long")
    }

    strBuf := make([]byte, strLen)
    _, err = io.ReadFull(r, strBuf)
    if err != nil {
        return err
    }
    *s = strBuf
    return nil
}