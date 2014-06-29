// The script evaluation function.
package script

import (
    "github.com/oxfeeefeee/kaiju/catma/numbers"
    )

var fnTable []execFunc

type evalFlag byte

// The value here is different from SCRIPT_VERIFY_XXXX in Satoshi client
// they put SCRIPT_VERIFY_NOCACHE here, which is confusing because
// it has nothing to do with bitcoin protocal itself.
const (
    // evaluate P2SH (BIP16) subscripts
    evalFlag_P2SH evalFlag = 1 << iota
    // enforce strict conformance to DER and SEC2 for signatures and pubkeys
    evalFlag_STRICTENC          
    // enforce low S values (<n/2) in signatures (depends on STRICTENC)
    evalFlag_LOW_S              
    // verify dummy stack item consumed by CHECKMULTISIG is of zero-length
    evalFlag_NULLDUMMY          
)

// Context used by execXXXX functions
type execContext struct {
    stack       *stack          // Script running main stack
    altStack    stack           // Alt stack
    bStack      boolStack       // Branching stack
    separator   int             // Hash starts after the code separator
    pc          int             // Next pc
    opCount     int             // Opcode count
    script      Script
    sctx        scriptContext
    flags       evalFlag
}

type execFunc func(ctx *execContext, op Opcode, operand []byte) error

func (s *stack) eval(script Script, c scriptContext, flags evalFlag) error {
    pc := 0
    ctx := &execContext{s, make([]stackItem, 0), make([]bool, 0),
        0, 0, 0, script, c, flags}
    for pc < len(script) {
        op, operand, next, err := script.getOpcode(pc)
        pc = next
        ctx.pc = next
        if err != nil {
            return err
        }

        if op >= OP_NOP {
            ctx.opCount++
            if ctx.opCount > numbers.MaxOpcodeCount {
                return errOpcodeCount
            }
        }

        if int(op) >= len(fnTable) {
            return errInvalidOp
        }

        alive := ctx.bStack.alive()   
        if !alive && (op >= OP_IF && op <= OP_ENDIF) {
            // Skip the code if we are in non-execute branch and the op is not
            // OP_IF / OP_NOTIF / OP_ELSE / OP_ENDIF
            continue
        }

        fn := fnTable[op]
        err = fn(ctx, op, operand)
        if err != nil {
            return err
        }
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


