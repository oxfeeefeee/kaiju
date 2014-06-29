package script

import (
    "github.com/oxfeeefeee/kaiju/klib"
    )

// OP_1ADD
// OP_1SUB
// OP_NEGATE
// OP_ABS
// OP_NOT
// OP_0NOTEQUAL
func execNumeric1(ctx *execContext, op Opcode, _ []byte) error {
    if ctx.stack.empty() {
        return errStackItemMissing
    }
    top := ctx.stack.pop()
    if klib.ScriptIntOverflow(top) {
        return errScriptIntOverflow
    }
    n := klib.ToScriptInt(top)
    // Don't need to worry about ScriptInt(64bit) overflow, becase n is 32bit
    switch op {
    case OP_1ADD:
        n += 1
    case OP_1SUB:
        n -= 1
    case OP_NEGATE:
        n = -n
    case OP_ABS:
        if n < 0 { 
            n = -n 
        }
    case OP_NOT:
        if n == 0 {
            n = 1
        } else {
            n = 0
        }
    case OP_0NOTEQUAL:
        if n != 0 {
            n = 1
        }
    }
    ctx.stack.push(n.Bytes())
    return nil
}

// OP_ADD
// OP_SUB
// OP_BOOLAND
// OP_BOOLOR
// OP_NUMEQUAL
// OP_NUMEQUALVERIFY
// OP_NUMNOTEQUAL
// OP_LESSTHAN
// OP_GREATERTHAN
// OP_LESSTHANOREQUAL
// OP_GREATERTHANOREQUAL
// OP_MIN
// OP_MAX
func execNumeric2(ctx *execContext, op Opcode, _ []byte) error {
    if ctx.stack.height() < 2 {
        return errStackItemMissing
    }
    top1, top2 := ctx.stack.pop(), ctx.stack.pop()
    if klib.ScriptIntOverflow(top1) || klib.ScriptIntOverflow(top2) {
        return errScriptIntOverflow
    }
    n1, n2 := klib.ToScriptInt(top2), klib.ToScriptInt(top1)

    if op == OP_NUMEQUALVERIFY {
        if n1 == n2 {
            return nil
        } else {
            return errVerifyFailed
        }
    }
    var n klib.ScriptInt
    // Don't need to worry about ScriptInt(64bit) overflow, becase n1/n2 is 32bit
    switch op {
    case OP_ADD:
        n = n1 + n2
    case OP_SUB:
        n = n1 - n2
    case OP_BOOLAND:
        n = 0
        if n1 != 0 && n2 != 0 {
            n = 1
        }
    case OP_BOOLOR:
        n = 0
        if n1 != 0 || n2 != 0 {
            n = 1
        }
    case OP_NUMEQUAL:
        n = 0
        if n1 == n2 {
            n = 1
        }
    case OP_NUMNOTEQUAL:
        n = 0
        if n1 != n2 {
            n = 1
        }      
    case OP_LESSTHAN:
        n = 0
        if n1 < n2 {
            n = 1
        }
    case OP_GREATERTHAN:
        n = 0
        if n1 > n2 {
            n = 1
        }
    case OP_LESSTHANOREQUAL:
        n = 0
        if n1 <= n2 {
            n = 1
        } 
    case OP_GREATERTHANOREQUAL:
        n = 0
        if n1 >= n2 {
            n = 1
        }
    case OP_MIN:
        n = n1
        if n2 < n1 {
            n = n2
        }
    case OP_MAX:
        n = n1
        if n2 > n1 {
            n = n2
        }  
    }
    ctx.stack.push(n.Bytes())
    return nil
}

func execWithin(ctx *execContext, _ Opcode, _ []byte) error {
    if ctx.stack.height() < 3 {
        return errStackItemMissing
    }
    top1, top2, top3 := ctx.stack.pop(), ctx.stack.pop(), ctx.stack.pop()
    if (klib.ScriptIntOverflow(top1) ||
        klib.ScriptIntOverflow(top2) ||
        klib.ScriptIntOverflow(top3)){
        return errScriptIntOverflow
    }
    n1 := klib.ToScriptInt(top3)
    n2, n3 := klib.ToScriptInt(top2), klib.ToScriptInt(top1)
    n := klib.ScriptInt(0)
    if n1 >= n2 && n1 < n3 {
        n = 1
    }
    ctx.stack.push(n.Bytes())
    return nil
}
