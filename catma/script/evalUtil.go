// Utility functions for script evaluation
package script

import (
    "github.com/oxfeeefeee/kaiju/klib"
    )

type txContent interface {
    HashToSign(subScript []byte, ii int, hashType uint32) *klib.Hash256
}

// OP_CHECKSIG and OP_CHECKMULTISIG
//
// See https://en.bitcoin.it/wiki/OP_CHECKSIG
// Also, interesting read: 
//   -- https://bitcointalk.org/index.php?topic=260595.0
//   -- http://bitcoin.stackexchange.com/questions/4213/what-is-the-point-of-sighash-none
func verifySig(tx txContent, inputI int, pk []byte, sig []byte, subScript Script, currentI int) bool {
    // STEP1: Remove signature from subScript

    // STEP2: Remove OP_CODESEPARATOR's from subScript

    // STEP3: Get hash to sign from TX
    //hash := tx.HashToSign(subScript, inputI, uint32(pk[len(pk) - 1]))

    // STEP5: Verify the sig

    return true
}