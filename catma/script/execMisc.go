package script

import (
    "bytes"
    "github.com/oxfeeefeee/kaiju/klib"
    )

func execInvalid(_ *execContext, _ Opcode, _ []byte) error {
    return errInvalidOp
}

func execNop(_ *execContext, _ Opcode, _ []byte) error {
    return nil
}

func execSize(ctx *execContext, _ Opcode, _ []byte) error {
    if ctx.stack.empty() {
        return errStackItemMissing
    }
    si := klib.ScriptInt(len(ctx.stack.top(-1))).Bytes()
    ctx.stack.push(si)
    return nil
}

// OP_EQUAL
// OP_EQUALVERIFY
func execEqual(ctx *execContext, op Opcode, _ []byte) error {
    if ctx.stack.height() < 2 {
        return errStackItemMissing
    }
    si1, si2 := ctx.stack.pop(), ctx.stack.pop()
    equal := bytes.Equal(si1, si2)
    if op == OP_EQUALVERIFY {
        if !equal {
            return errEqualVerifyFailed
        }
    } else {
        ctx.stack.push(boolToStackItem(equal))
    }
    return nil
}

func execSeparator(ctx *execContext, _ Opcode, _ []byte) error {
    ctx.separator = ctx.pc
    return nil
}