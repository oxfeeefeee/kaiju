package script

import (
    //"fmt"
    "errors"
    )

// Script erros are listed here to avoid errors.New calls and to 
// make it easier to manage

var errEOS = errors.New("Script.getOpcode: End of script")

var errDataNotFoundToPush = errors.New("Script.getOpcode: Data size not found after OP_PUSHDATAX")

var errInvalidOp =  errors.New("eval: Invalid opcode")

var errOpcodeCount = errors.New("eval: Opcode count exceeds limit")

var errStackItemMissing = errors.New("eval: Stack item count less than expected")

var errIfElseMismatch = errors.New("eval: OP_IF / OP_ELSE / OP_ENDIF mismatch")

var errVerifyFailed = errors.New("eval: OP_VERIFY failed")

var errReturned = errors.New("eval: OP_RETURN")