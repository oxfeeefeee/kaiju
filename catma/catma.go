// Dogma of Bitcoin, be very careful!
package catma

import (
    "errors"
    "github.com/oxfeeefeee/kaiju"
    "github.com/oxfeeefeee/kaiju/catma/script"
)

type InputVerifyType uint32

const (
    InputVerifyPreBIP16 InputVerifyType = iota
    InputVerifyMandatory
    InputVerifyStandard
)

func VerifyInput(pkScript []byte, tx *Tx, idx int, vt InputVerifyType) error {
    var evalFlags script.EvalFlag
    switch vt {
    case InputVerifyPreBIP16:
        evalFlags = script.EvalFlagNone
    case InputVerifyMandatory:
        evalFlags = script.EvalFlagP2SH
    case InputVerifyStandard:
        evalFlags = script.EvalFlagP2SH | script.EvalFlagStrictEnc | script.EvalFlagNullDummy
    default:
        return errors.New("VerifyInput: Invalid verify type")
    }
    return VerifyInputWithFlags(pkScript, tx, idx, evalFlags)
}

func VerifyInputWithFlags(pkScript []byte, tx *Tx, idx int, flags script.EvalFlag) error {
    if idx >= len(tx.TxIns) {
        return errors.New("VerifyInput: Input index out of range")
    }
    sig := tx.TxIns[idx].SigScript
    ie := &InputEntry{tx, idx}
    return script.VerifyScript(pkScript, sig, ie, flags)
}

func VerifySig(pkScript []byte, sigScript []byte) bool {
    return true
}

// Handy function
func logger() *kaiju.Logger {
    return kaiju.CatmaLogger
}