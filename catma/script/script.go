package script

import (
    "encoding/binary"
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
        s.AppendData(klib.Uint16ToBytes(uint16(len(p))))
    default:
        s.AppendOp(OP_PUSHDATA4)
        s.AppendData(klib.Uint32ToBytes(uint32(len(p))))
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

// Returns if all opcode are data push
func (s Script) IsPushOnly() bool {
    next := 0
    for next < len(s){
        op, _, np, err := s.getOpcode(next)
        next = np
        if err != nil || op > OP_16 {
            return false
        }
    }
    return true
}

// Returns if all data-push-opcodes are canonical
func (s Script) PushesCanonical() bool {
    next := 0
    for next < len(s){
        op, operand, np, err := s.getOpcode(next)
        next = np
        if err != nil {
            return false
        }
        switch {
        case op > OP_PUSHDATA00 && op < OP_PUSHDATA1:
            // Could have used an OP_N, rather than a 1-byte push.
            return !(len(operand) == 1 && operand[0] <= 16)
        case op == OP_PUSHDATA1:
            // Could have used an OP_PUSHDATAXX, rather than OP_PUSHDATA1.
            return len(operand) >= int(OP_PUSHDATA1)
        case op == OP_PUSHDATA2:
            // Could have used an OP_PUSHDATA1.
            return len(operand) > 0xFF
        case op == OP_PUSHDATA4:
            return len(operand) > 0xFFFF
        }
    }
    return true
}

// Returns (True, how-many-items-on-stack) after SigScript is run 
// for valid standard PKScript, (False, 0) otherwise

func (s Script) SigArgsExpected(t PKScriptType) (bool, int) {
    switch t {
    case PKS_NonStandard, PKS_NullData:
        return false, 0
    case PKS_PubKey:
        return true, 1  // Expect: <sig>
    case PKS_PubKeyHash:
        return true, 2  // Expect: <sig> <pubKey>
    case PKS_ScriptHash:
        return true, 1  // Expect: <sig> <script> but <script> doesn't count
    case PKS_MultiSig:
        m := Opcode(s[0]).number()
        return true, m + 1    // Expect: m * <sig> + Satoshi_Bug
    }
    return false, 0
}

// Returns opcode at p and related data
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

