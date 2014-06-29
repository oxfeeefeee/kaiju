package script

import (
    )

func execControl(ctx *execContext, op Opcode, _ []byte) error {
    switch op {
    case OP_VERIFY:
        if ctx.stack.empty() {
            return errStackItemMissing
        }
        val := ctx.stack.pop().toBool()
        if !val {
            return errVerifyFailed
        }
    case OP_RETURN:
        return errReturned
    }
    return nil
}

func execBranching(ctx *execContext, op Opcode, _ []byte) error {
    switch op {
    case OP_IF, OP_NOTIF:
        val := false
        if alive := ctx.bStack.alive(); alive {
            if ctx.stack.empty() {
                return errStackItemMissing
            }
            val = ctx.stack.pop().toBool()
            if op == OP_NOTIF {
                val = !val
            }
        }
        ctx.bStack.push(val)
    case OP_ELSE:
        if ctx.bStack.empty() {
            return errIfElseMismatch
        }
        ctx.bStack.push(!ctx.bStack.pop())
    case OP_ENDIF:
        if ctx.bStack.empty() {
            return errIfElseMismatch
        }
        ctx.bStack.pop()        
    }
    return nil
} 