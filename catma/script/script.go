package script

import (
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju"
    "github.com/oxfeeefeee/kaiju/klib"
    )

type Script []byte

func NewScript() *Script {
    return &Script{}
}

func (s *Script) AppendOp(op Opcode) {
    *s = append(*s, byte(op))
}

func (s *Script) AppendData(p []byte) {
    *s = append(*s, p...)
}

// Append a PushData operation.
// notice the length of data is NOT encoded as a "stackInt"(stackItem representing a integer)
func (s *Script) AppendPushData(p []byte) {
    switch  {
    case len(p) < int(OP_PUSHDATA1):
        s.AppendOp(Opcode(len(p)))
    case len(p) <= 0xff:
        s.AppendOp(OP_PUSHDATA1)
        s.AppendData([]byte{byte(len(p)),})
    case len(p) <= 0xffff:
        s.AppendOp(OP_PUSHDATA2)
        s.AppendData(klib.UInt16ToBytes(uint16(len(p))))
    default:
        s.AppendOp(OP_PUSHDATA4)
        s.AppendData(klib.UInt32ToBytes(uint32(len(p))))
    } 
    s.AppendData(p)
}

func (s *Script) AppendPushInt(v int64) {
    if v == int64(OP_1NEGATE) || (v >= 1 && v <= 16) {
        s.AppendOp(Opcode(v + int64(OP_1) - 1))
    } else {
        i := klib.ScriptInt(v)
        s.AppendPushData(i.Bytes())
    }
}

func (s Script) getOpcode(p int) (op Opcode, operand []byte, next int, err error) {
    if p >= len(s) {
        err = errEOS
        return
    }
    op = Opcode(s[p])
    next = p + 1
    if op <= OP_PUSHDATA4 {
        size := 0
        switch {
        case op < OP_PUSHDATA1:
            size = int(op)
        case op == OP_PUSHDATA1:
            if next >= len(s) {
                err = errDataNotFoundToPush
                return
            }
            size = int(s[next])
            next += 1
        case op == OP_PUSHDATA2:
            if next >= len(s) - 1 {
                err = errDataNotFoundToPush
                return
            }
            size = int(binary.LittleEndian.Uint16(s[next:next+2]))
            next += 2
        case op == OP_PUSHDATA4:
            if next >= len(s) - 3 {
                err = errDataNotFoundToPush
                return
            }
            size = int(binary.LittleEndian.Uint32(s[next:next+4]))
            next += 4  
        }
        if next > len(s) - size {
            err = errDataNotFoundToPush
            return
        }
        operand = s[next:next+size]
        next += size
    }
    return
}

// Handy function
func logger() *kaiju.Logger {
    return kaiju.CatmaScriptLogger
}
