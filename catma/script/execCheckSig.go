// Utility functions for script evaluation
package script

import (
    "bytes"
    "errors"
    "crypto/ecdsa"
    "github.com/oxfeeefeee/kaiju/klib"
    )

func execCheckSig(_ *execContext, _ Opcode, _ []byte) error {
    return nil
}

func execCheckMultiSig(_ *execContext, _ Opcode, _ []byte) error {
    return nil
}

// OP_CHECKSIG and OP_CHECKMULTISIG
// See https://en.bitcoin.it/wiki/OP_CHECKSIG
func verifySig(c scriptContext, pk []byte, sig []byte, subScript Script, currentI int) error {
    // STEP1: 
    // - Remove signature from subScript if present
    // - Remove OP_CODESEPARATOR's from subScript
    for current := 0; current < len(subScript); {
        op, operand, next, err := subScript.getOpcode(current)
        if err != nil {
            return err
        }
        if bytes.Equal(operand, sig) || op == OP_CODESEPARATOR{ 
            // if operand == sig:           Remove [OP_PUSHDATAX-len-Sig]
            // if op == OP_CODESEPARATOR:   Remove OP_CODESEPARATOR
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

