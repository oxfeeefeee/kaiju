package script

import (
    "bytes"
    "errors"
    "crypto/ecdsa"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/catma/numbers"
    )

// TODO: IsCanonical

// OP_CHECKSIG
// OP_CHECKSIGVERIFY
func execCheckSig(ctx *execContext, op Opcode, _ []byte) error {
    if ctx.stack.height() < 2 {
        return errStackItemMissing
    }
    pk := ctx.stack.pop()
    sig := ctx.stack.pop()
    subScript := make([]byte, 0, len(ctx.script) - ctx.pc)
    copy(subScript, ctx.script[:ctx.separator])
    subScript, err := removeSig(subScript, sig)
    if err != nil {
        return err
    }
    err = verifySig(ctx.sctx, pk, sig, subScript)
    if op == OP_CHECKSIGVERIFY {
        if err != nil {
            return errSigVerify
        }
    } else { // OP_CHECKSIG
        b := klib.ScriptInt(1)
        if err != nil {
            b = 0
        }
        ctx.stack.push(b.Bytes())
    }
    return nil
}

// OP_CHECKMULTISIG
// OP_CHECKMULTISIGVERIFY
func execCheckMultiSig(ctx *execContext, op Opcode, _ []byte) error {
    i := 1
    if ctx.stack.height() < i {
        return errStackItemMissing
    }
    kCount := int(klib.ToScriptInt(ctx.stack.top(-i)))
    if kCount < 0 || kCount > numbers.MaxMultiSigKeyCount {
        return errKeySigCountOutOfRange
    }
    ctx.opCount += kCount
    if ctx.opCount > numbers.MaxOpcodeCount {
        return errOpcodeCount
    }
    i++
    iKey := i
    i += kCount
    if ctx.stack.height() < i {
        return errStackItemMissing
    }
    sCount := int(klib.ToScriptInt(ctx.stack.top(-i)))
    if sCount < 0 || sCount > kCount {
        return errKeySigCountOutOfRange
    }
    i++
    iSig := i
    i += sCount
    if ctx.stack.height() < i {
        return errStackItemMissing
    }

    subScript := make([]byte, 0, len(ctx.script) - ctx.pc)
    copy(subScript, ctx.script[:ctx.separator])
    for k := 0; k < sCount; k++ {
        var err error
        subScript, err = removeSig(subScript, ctx.stack.top(-(iSig + k)))
        if err != nil {
            return err
        }
    }
    success := true
    for success && (sCount > 0) {
        pk := ctx.stack.top(-iKey)
        sig := ctx.stack.top(-iSig)
        err := verifySig(ctx.sctx, pk, sig, subScript)
        if err == nil {
            iSig++
            sCount--
        }
        iKey++
        kCount--
        // If there are more signatures left than keys left,
        // then too many signatures have failed
        if (sCount > kCount) {
            success = false
        }
    }

    // A old bug causes CHECKMULTISIG to consume an extra item on the stack
    // We first clear all the "real" arguments
    stk := *ctx.stack
    *ctx.stack = stk[:len(stk)-i+1]
    // The dummy item is still requried
    if ctx.stack.empty() {
        return errStackItemMissing    
    }
    // Now we check dummy being null when required
    si := ctx.stack.pop()
    if (ctx.flags & evalFlag_NULLDUMMY) != 0 && len(si) > 0 {
        return errDummyArgNotNull
    }

    if op == OP_CHECKMULTISIGVERIFY {
        if !success {
            return errSigVerify
        } 
    } else { // OP_CHECKMULTISIG
        b := klib.ScriptInt(1)
        if !success {
            b = 0
        }
        ctx.stack.push(b.Bytes())
    }
    return nil
}

func removeSig(subScript Script, sig []byte) (Script, error) {
    for current := 0; current < len(subScript); {
        _, operand, next, err := subScript.getOpcode(current)
        if err != nil {
            return nil, err
        }
        if bytes.Equal(operand, sig) { 
            // Remove [OP_PUSHDATAX-len-Sig]
            subScript = append(subScript[:current], subScript[next:]...)
            // Do not advance the cursor in this case
        } else {
            current = next
        }
    }
    return subScript, nil
}

// See https://en.bitcoin.it/wiki/OP_CHECKSIG
func verifySig(c scriptContext, pk []byte, sig []byte, subScript Script) error {
    // STEP1: Remove OP_CODESEPARATOR's from subScript
    for current := 0; current < len(subScript); {
        op, _, next, err := subScript.getOpcode(current)
        if err != nil {
            return err
        }
        if op == OP_CODESEPARATOR{       
            // Remove OP_CODESEPARATOR
            subScript = append(subScript[:current], subScript[next:]...)
            // Do not advance the cursor in this case
        } else {
            current = next
        }
    }

    // STEP2: Get hash to sign from Context
    hash, err := c.HashToSign(subScript, uint32(pk[len(pk) - 1]))
    if err != nil {
        return err
    }

    // STEP3: Verify the sig
    pubKey, err := klib.PubKey(pk).GoPubKey()
    if err != nil {
        return err
    }
    r, s, err := klib.Sig(sig).GoSig()
    if err != nil {
        return err
    }
    if !ecdsa.Verify(pubKey, hash[:], r, s) {
        return errors.New("verifySig: ecdsa verification failed")
    }
    return nil
}

