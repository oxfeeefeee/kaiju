// The script evaluation function.
package script

import (
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/catma/numbers"
    )

type scriptContext interface {
    HashToSign(subScript []byte, hashType byte) (*klib.Hash256, error)
}

var fnTable []execFunc

// Context used by execXXXX functions
type execContext struct {
    stack       *stack          // Script running main stack
    altStack    *stack          // Alt stack
    bStack      boolStack       // Branching stack
    separator   int             // Hash starts after the code separator
    pc          int             // Next pc
    opCount     int             // Opcode count
    script      Script
    sctx        scriptContext
    flags       EvalFlag
}

type execFunc func(ctx *execContext, op Opcode, operand []byte) error

func (s *stack) eval(script Script, c scriptContext, flags EvalFlag) error {
    if len(script) > numbers.MaxScriptSize {
        return errScriptSizeLimit
    }
    pc := 0
    ctx := &execContext{s, &stack{}, make([]bool, 0),
        0, 0, 0, script, c, flags}
    for pc < len(script) {
        op, operand, next, err := script.getOpcode(pc)
        //logger().Debugf("op: %s %v\n", op, operand)
        pc = next
        ctx.pc = next
        if err != nil {
            return err
        }

        if len(operand) > numbers.MaxScriptElementSize {
            return errOperandSizeLimit
        }

        if op >= OP_NOP {
            ctx.opCount++
            if ctx.opCount > numbers.MaxOpcodeCount {
                return errOpcodeCount
            }
        }

        // Another Satoshi Bug: any other junk data can be included in script as long as not
        // getting executed, but Disabled Opcodes make the script invalid no matter what.
        if int(op) >= 0 && int(op) < len(fnTable) && fnTable[op] == nil {
            return errDisabledOp
        }

        alive := ctx.bStack.alive()  
        if !alive && !(op >= OP_IF && op <= OP_ENDIF) {
            // Skip the code if we are in non-execute branch and the op is not
            // OP_IF / OP_NOTIF / OP_ELSE / OP_ENDIF
            continue
        }

        if op < 0 || int(op) >= len(fnTable) {
            return errInvalidOp
        }

        fn := fnTable[op]
        err = fn(ctx, op, operand)
        if err != nil {
            return err
        }

        if ctx.stack.height() + ctx.altStack.height() > numbers.MaxScriptEvalStackSize {
            return errStackSizeLimit
        }
    }
    if !ctx.bStack.empty() {
        return errIfElseMismatch
    }
    return nil
}

// Init function table
func init() {
    fnTable = make([]execFunc, 0, byte(OP_NOP10) + 1)
    for op := OP_PUSHDATA00; op <= OP_NOP10; op++ {
        _, fn := op.attr()
        fnTable = append(fnTable, fn)
    }
}


