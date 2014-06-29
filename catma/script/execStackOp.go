package script

import (
    "github.com/oxfeeefeee/kaiju/klib"
    )

func execStackOp(ctx *execContext, op Opcode, _ []byte) error {
    switch op {
    case OP_TOALTSTACK:
        if ctx.stack.empty() {
            return errStackItemMissing
        }
        ctx.altStack.push(ctx.stack.pop())
    case OP_FROMALTSTACK:
        if ctx.altStack.empty() {
            return errStackItemMissing
        }
        ctx.stack.push(ctx.altStack.pop())
    case OP_2DROP:
        // (x1 x2 -- )
        if ctx.stack.height() < 2 {
            return errStackItemMissing
        }
        ctx.stack.pop()
        ctx.stack.pop()
    case OP_2DUP:
        // (x1 x2 -- x1 x2 x1 x2)
        if ctx.stack.height() < 2 {
            return errStackItemMissing
        }
        si1, si2 := ctx.stack.top(-2), ctx.stack.top(-1)
        ctx.stack.push(si1)
        ctx.stack.push(si2)
    case OP_3DUP:
        // (x1 x2 x3 -- x1 x2 x3 x1 x2 x3)
        if ctx.stack.height() < 3 {
            return errStackItemMissing
        }
        si1, si2, si3 := ctx.stack.top(-3), ctx.stack.top(-2), ctx.stack.top(-1)
        ctx.stack.push(si1)
        ctx.stack.push(si2)
        ctx.stack.push(si3)
    case OP_2OVER:
        // (x1 x2 x3 x4 -- x1 x2 x3 x4 x1 x2)
        if ctx.stack.height() < 4 {
            return errStackItemMissing
        }
        si1, si2 := ctx.stack.top(-4), ctx.stack.top(-3)
        ctx.stack.push(si1)
        ctx.stack.push(si2)
    case OP_2ROT:
        // (x1 x2 x3 x4 x5 x6 -- x3 x4 x5 x6 x1 x2)
        if ctx.stack.height() < 6 {
            return errStackItemMissing
        }
        s := *(ctx.stack)
        l := len(s)
        *(ctx.stack) = append(append(append(s[:l-6], s[l-4:]...), s[l-6]), s[l-5])
    case OP_2SWAP:
        // (x1 x2 x3 x4 -- x3 x4 x1 x2)
        if ctx.stack.height() < 4 {
            return errStackItemMissing
        }
        s := *(ctx.stack)
        l := len(s)
        *(ctx.stack) = append(append(s[:l-4], s[l-2:]...), s[l-4:l-2]...)
    case OP_IFDUP:
        // Duplicate if true
        if ctx.stack.empty() {
            return errStackItemMissing
        }
        si := ctx.stack.top(-1)
        if si.toBool() {
            ctx.stack.push(si)
        }
    case OP_DEPTH:
        // Push stack height on to stack
        ctx.stack.push(intToStackItem(ctx.stack.height()))
    case OP_DROP:
        if ctx.stack.empty() {
            return errStackItemMissing
        }
        ctx.stack.pop()
    case OP_DUP:
        // Duplicate stack top
        if ctx.stack.empty() {
            return errStackItemMissing
        }
        ctx.stack.push(ctx.stack.top(-1))
    case OP_NIP:
        // (x1 x2 -- x2)
        if ctx.stack.height() < 2 {
            return errStackItemMissing
        }
        s := *(ctx.stack)
        l := len(s)
        *ctx.stack = append(s[:l-2], s[l-1])
    case OP_OVER:
        // (x1 x2 -- x1 x2 x1)
        if ctx.stack.height() < 2 {
            return errStackItemMissing
        }
        s := *(ctx.stack)
        ctx.stack.push(s[len(s)-2])
    case OP_PICK, OP_ROLL:
        // (xn ... x2 x1 x0 n - xn ... x2 x1 x0 xn)
        // (xn ... x2 x1 x0 n - ... x2 x1 x0 xn)
        if ctx.stack.height() < 2 {
            return errStackItemMissing
        }
        top := ctx.stack.pop()
        if klib.ScriptIntOverflow(top) {
            return errScriptIntOverflow
        }
        n := int(klib.ToScriptInt(top))
        if n < 0 || n >= ctx.stack.height() {
            return errIndexOutOfRange
        }
        s := *(ctx.stack)
        l := len(s)
        si := s[l-n-1]
        if op == OP_ROLL {
            s = append(s[:l-n-1], s[l-n:]...)
        }
        *ctx.stack = append(s, si)
    case OP_ROT:
        // (x1 x2 x3 -- x2 x3 x1)
        if ctx.stack.height() < 3 {
            return errStackItemMissing
        }
        s := *(ctx.stack)
        l := len(s)
        *(ctx.stack) = append(append(append(s[:l-3], s[l-2]), s[l-1]), s[l-3])
    case OP_SWAP:
        // (x1 x2 -- x2 x1)
        if ctx.stack.height() < 2 {
            return errStackItemMissing
        }
        s := *(ctx.stack)
        l := len(s)
        *(ctx.stack) = append(append(s[:l-2], s[l-1]), s[l-2])
    case OP_TUCK:
        // (x1 x2 -- x2 x1 x2)
        if ctx.stack.height() < 2 {
            return errStackItemMissing
        }
        s := *(ctx.stack)
        l := len(s)
        *(ctx.stack) = append(append(append(s[:l-2], s[l-1]), s[l-2]), s[l-1])
    }
    return nil
}
