package script

import (
    //"fmt"
    //"errors"
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
        if ctx.stack.height() < 2 {
            return errStackItemMissing
        }
        ctx.stack.pop()
        ctx.stack.pop()
    default:
        panic("Unknown opcode!") // Should not happen
    }
    return nil
}