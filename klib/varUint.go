package klib

import (
    "io"
    "errors"
    "encoding/binary"
)

// Value               StorageLength   Format
// <  0xfd             1               uint8_t
// <= 0xffff           3               0xfd followed by the length as uint16_t
// <= 0xffffffff       5               0xfe followed by the length as uint32_t
// >  0xffffffff       9               0xff followed by the length as uint64_t
type VarUint uint64

func (v VarUint) Bytes() []byte {
    var buffer [9]byte
    dataLen := 0
    switch {
    case v < 0xfd:
        buffer[0] = byte(v)
        dataLen = 1
    case v <= 0xffff:
        buffer[0] = 0xfd
        binary.LittleEndian.PutUint16(buffer[1:], uint16(v))
        dataLen = 3
    case v <= 0xffffffff:
        buffer[0] = 0xfe
        binary.LittleEndian.PutUint32(buffer[1:], uint32(v))
        dataLen = 5
    case v > 0xffffffff:
        buffer[0] = 0xff
        binary.LittleEndian.PutUint64(buffer[1:], uint64(v))
        dataLen = 9
    }
    return buffer[:dataLen]
}

func (v VarUint) Serialize(w io.Writer) error {
    _, err := w.Write(v.Bytes())
    return err
}

func (v *VarUint) Deserialize(r io.Reader) error {
    oneByteBuf := make([]byte, 1)
    _, err := io.ReadFull(r, oneByteBuf)
    b := oneByteBuf[0]
    switch {
    case b < 0xfd:
        *v = VarUint(b)
        return nil
    case b == 0xfd:
        twoBytesBuf := make([]byte, 2)
        _, err = io.ReadFull(r, twoBytesBuf)
        if err != nil {
            return err
        }
        *v = VarUint(binary.LittleEndian.Uint16(twoBytesBuf))
        return nil
    case b == 0xfe:
        fourBytesBuf := make([]byte, 4)
        _, err = io.ReadFull(r, fourBytesBuf)
        if err != nil {
            return err
        }
        *v = VarUint(binary.LittleEndian.Uint32(fourBytesBuf))
        return nil
    case b == 0xff:
        eightBytesBuf := make([]byte, 8)
        _, err = io.ReadFull(r, eightBytesBuf)
        if err != nil {
            return err
        }
        *v = VarUint(binary.LittleEndian.Uint64(eightBytesBuf))
        return nil
    }  
    return errors.New("VarUint.Deserialize internal error")
}
