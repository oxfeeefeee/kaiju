package script

import (
    "errors"
    )

// Script erros are listed here to avoid errors.New calls and to 
// make it easier to manage

var errScriptIntOverflow = errors.New("ScriptInt.SetBytes: slice length larger than Maximum")

var errEOS = errors.New("Script.getOpcode: End of script")

var errDataNotFoundToPush = errors.New("Script.getOpcode: Data size not found after OP_PUSHDATAX")

var errInvalidOp =  errors.New("eval: Invalid opcode")

var errOpcodeCount = errors.New("eval: Opcode count exceeds limit")

var errStackItemMissing = errors.New("eval: Stack item count less than expected")

var errIfElseMismatch = errors.New("eval: OP_IF / OP_ELSE / OP_ENDIF mismatch")

var errVerifyFailed = errors.New("eval: OP_VERIFY failed")

var errEqualVerifyFailed = errors.New("eval: OP_EQUALVERIFY failed")

var errReturned = errors.New("eval: OP_RETURN")

var errIndexOutOfRange = errors.New("eval: OP_PICK/OP_ROLL index out of range")

var errKeySigCountOutOfRange = errors.New("eval: MultiSig key/sig index out of range")

var errDummyArgNotNull = errors.New("CHECKMULTISIG dummy argument not null")

var errSigVerify = errors.New("Signature verification failed")

var errPKNonCanonical = errors.New("Non-canonical public key")

var errSigNonCanonical = errors.New("Non-canonical signature")