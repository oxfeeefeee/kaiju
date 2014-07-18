package script

type EvalFlag uint32

// The value here is different from SCRIPT_VERIFY_XXXX in Satoshi client
// they put SCRIPT_VERIFY_NOCACHE here, which is confusing because
// it has nothing to do with bitcoin protocal itself.
const (
    // EvalFlagP2SH: evaluate P2SH (BIP16) subscripts
    EvalFlagNone, EvalFlagP2SH EvalFlag = 0, 1 << iota
    // enforce strict conformance to DER and SEC2 for signatures and pubkeys
    EvalFlagStrictEnc = 1 << iota        
    // enforce low S values (<n/2) in signatures (depends on STRICTENC)
    // TODO: not implemented yet
    EvalFlagLowS             
    // verify dummy stack item consumed by CHECKMULTISIG is of zero-length
    EvalFlagNullDummy       
)

func RunSigScript(sigScript Script) (error, [][]byte) {
    sstack := stack{}
    err := sstack.eval(sigScript, nil, EvalFlagNone)
    if err != nil {
        return err, nil
    } else {
        return nil, (interface{}(sstack)).([][]byte)
    }
}

// Runs pure scripts(without CheckSig), returns evaluation result or error
func RunScript(pkScript Script, sigScript Script) error {
    return VerifyScript(pkScript, sigScript, nil, EvalFlagP2SH)
}

func VerifyScript(pkScript Script, sigScript Script, sctx scriptContext, flags EvalFlag) error {
    sstack := stack{}
    var stackCopy stack 
    // First eval sigScript
    err := sstack.eval(sigScript, sctx, flags)
    if err != nil {
        return err
    }
    // Make a copy of the stack for P2SH Tx
    p2sh := ((flags & EvalFlagP2SH) != 0)
    if p2sh {
        stackCopy = make([]stackItem, len(sstack))
        copy(stackCopy, sstack)
    }
    // Eval pkScript
    err = sstack.eval(pkScript, sctx, flags)
    if err != nil {
        return err
    }
    // Stack top needs to be "true"
    if sstack.empty() || !sstack.top(-1).toBool() {
        return errEvalNotTrue
    }

    // Extra verification for P2SH
    if p2sh && pkScript.IsTypeScriptHash() {
        if !sigScript.IsPushOnly() {
            return errP2SHSigNotPushOnly
        }
        pkScript2 := Script(stackCopy.pop())
        err := stackCopy.eval(pkScript2, sctx, flags)
        if err != nil {
            return err
        }
        if stackCopy.empty() || !stackCopy.top(-1).toBool() {
            return errEvalNotTrue
        }
    }
    return nil
}