package script

import (
    "bytes"
    "errors"
    "crypto/ecdsa"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/catma/numbers"
    )

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
    err = checkKeySig(ctx.sctx, pk, sig, subScript, ctx.flags)
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
        err := checkKeySig(ctx.sctx, pk, sig, subScript, ctx.flags)
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

func checkKeySig(c scriptContext, pk []byte, sig []byte, subScript Script, flags evalFlag) error {
    if (flags & evalFlag_STRICTENC) != 0 {
        if (flags & evalFlag_LOW_S) != 0 {
            panic("evalFlag_LOW_S not implemented")
        }
        err := canonicalPK(pk)
        if err != nil {
            return err
        }
        err = canonicalSig(sig)
        if err != nil {
            return err
        }
    }
    return verifySig(c, pk, sig, subScript)
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

func canonicalPK(pk []byte) error {
    l := len(pk)
    if l < 33 {
        return errPKNonCanonical
    }

    switch pk[0] {
    case 0x04:
        if l != 65 {
            return errPKNonCanonical
        }
    case 0x02, 0x03:
        if l != 33 {
            return errPKNonCanonical
        }
    default:
        return errPKNonCanonical
    }
    return nil
}

func canonicalSig(sig []byte) error {
    l := len(sig)
    if l < 9 || l > 73 {
        return errSigNonCanonical
    }
    hashType := sig[l-1] & 0x0f
    if hashType < 1 || hashType > 3 { 
        return errSigNonCanonical   // Unknown hashtype byte
    }
    if sig[0] != 0x30 { 
        return errSigNonCanonical   // Wrong type
    }
    if int(sig[1]) != (l - 3) {
        return errSigNonCanonical   // Wrong length marker
    }
    lenR := sig[3]
    if (5 + int(lenR)) >= l {
        return errSigNonCanonical   // S length misplaced
    }
    lenS := sig[5+lenR]
    if int(lenR + lenS + 7) != l {
        return errSigNonCanonical   // R+S length mismatch
    }

    rBegin := 4
    r := sig[rBegin:]
    if sig[rBegin-2] != 0x02 {
        return errSigNonCanonical   //R value type mismatch
    }
    if lenR == 0 {
        return errSigNonCanonical   //R length is zero
    }
    if (r[0] & 0x80) != 0 {
        return errSigNonCanonical   //R length is negative
    }
    if (lenR > 1) && (r[0] == 0x00) && ((r[1] & 0x80) == 0) {
        return errSigNonCanonical   // R value excessively padded
    }

    sBegin := 6 + lenR
    s := sig[sBegin:]
    if sig[sBegin-2] != 0x02 {
        return errSigNonCanonical   //S value type mismatch
    }
    if lenS == 0 {
        return errSigNonCanonical   //S length is zero
    }
    if (s[0] & 0x80) != 0 {
        return errSigNonCanonical   //S length is negative
    }
    if (lenS > 1) && (s[0] == 0x00) && ((s[1] & 0x80) == 0) {
        return errSigNonCanonical   //S value excessively padded
    }
    return nil
}
