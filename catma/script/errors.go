package script

import (
    "errors"
    )

// Script erros are listed here to avoid errors.New calls and to 
// make it easier to manage

var (
    errScriptIntOverflow = errors.New("ScriptInt.SetBytes: slice length larger than Maximum")

    errEOS = errors.New("Script.getOpcode: End of script")

    errDataNotFoundToPush = errors.New("Script.getOpcode: Data size not found after OP_PUSHDATAX")

    errInvalidOp =  errors.New("eval: Invalid opcode")

    errOpcodeCount = errors.New("eval: Opcode count exceeds limit")

    errStackItemMissing = errors.New("eval: Stack item count less than expected")

    errIfElseMismatch = errors.New("eval: OP_IF / OP_ELSE / OP_ENDIF mismatch")

    errVerifyFailed = errors.New("eval: OP_VERIFY failed")

    errEqualVerifyFailed = errors.New("eval: OP_EQUALVERIFY failed")

    errReturned = errors.New("eval: OP_RETURN")

    errIndexOutOfRange = errors.New("eval: OP_PICK/OP_ROLL index out of range")

    errKeySigCountOutOfRange = errors.New("eval: MultiSig key/sig index out of range")

    errDummyArgNotNull = errors.New("CHECKMULTISIG dummy argument not null")

    errSigVerify = errors.New("Signature verification failed")

    errPKNonCanonical = errors.New("Non-canonical public key")

    errSigNonCanonical = errors.New("Non-canonical signature")

    errEvalNotTrue = errors.New("Eval result is not true")

    errP2SHSigNotPushOnly = errors.New("P2SH sigScript not push-only")

    errEmptySig = errors.New("verifySig: empty signature")

    errStackSizeLimit = errors.New("eval: stack size exceeded limit")

    errOperandSizeLimit = errors.New("eval: size of opcode operand exceeded limit")

    errScriptSizeLimit = errors.New("eval: size of script exceeded limit")

    errDisabledOp = errors.New("eval: Disabled opcode in script")
    )