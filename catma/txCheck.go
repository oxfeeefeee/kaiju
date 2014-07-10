package catma

import (
    "errors"
    "github.com/oxfeeefeee/kaiju/catma/script"
    "github.com/oxfeeefeee/kaiju/catma/numbers"
)

var (
    // Tx.FormatCheck ----------------------------------------------------------------
    errEmptyInput = errors.New("Tx.FormatCheck: empty input")

    errEmptyOutput = errors.New("Tx.FormatCheck: empty output")

    errTxSizeLimit = errors.New("Tx.FormatCheck: size limit exceeded")

    errNegativeVOut = errors.New("Tx.FormatCheck: negative value out")

    errTooLargeVout = errors.New("Tx.FormatCheck: value out larger than SatoshiInTotal")

    errDuplicateInputs = errors.New("Tx.FormatCheck: duplicate inputs")

    errBadCoinBaseScriptSize = errors.New("Tx.FormatCheck: bad coin base script size")

    errPrevOutIsNull = errors.New("Tx.FormatCheck: previous output is null")

    // Tx.IsStandard -----------------------------------------------------------------
    errBadVersion = errors.New("Tx.IsStandard: bad version number")

    errNotFinal = errors.New("Tx.IsStandard: not final")

    errStdSizeLimit = errors.New("Tx.IsStandard: size limit exceeded")

    errSigScriptSizeLimit = errors.New("Tx.IsStandard: SigScript size limit exceeded")

    errSigScriptNotPushOnly = errors.New("Tx.IsStandard: SigScript not push only")

    errNonCanonicalPush = errors.New("TxIsStandard: contains non-canonical push")

    errNonStandardPKScript = errors.New("TxIsStandard: non-standard PKScript")

    errDustTxOut = errors.New("TxIsStandard: dust output")

    errMoreThanOneReturn = errors.New("TxIsStandard: more than one OP_RETURN")
    )

// CheckTransaction in Satoshi client
func (t *Tx) FormatCheck() error {
    if len(t.TxIns) == 0 {
        return errEmptyInput
    }
    if len(t.TxOuts) == 0 {
        return errEmptyOutput
    }
    if t.ByteSize() > numbers.MaxBlockSize {
        return errTxSizeLimit
    }
    // Check for negative or too big ouput values
    valueOut := int64(0)
    for _, txout := range t.TxOuts {
        if txout.Value < 0 {
            return errNegativeVOut
        }
        valueOut += txout.Value
        if valueOut > numbers.SatoshiInTotal {
            return errTooLargeVout
        }
    }
    // Check for dupliacted inputs
    set := make(map[OutPoint]bool)
    for _, txin := range t.TxIns {
        if set[txin.PreviousOutput] {
            return errDuplicateInputs
        }
        set[txin.PreviousOutput] = true
    }
    if t.IsCoinBase() {
        sl := len(t.TxIns[0].SigScript)
        if sl < numbers.MinCoinBaseSigScriptSize || sl > numbers.MaxCoinBaseSigScriptSize {
            return errBadCoinBaseScriptSize
        }
    } else {
        for _, txin := range t.TxIns {
            if txin.PreviousOutput.IsNull() {
                return errPrevOutIsNull
            }
        }
    }
    return nil
}

// Check if Tx is standard.
func (t *Tx) IsStandard(blockHeight uint32, blockTime uint32) error {
    if t.Version > numbers.TxCurrentVersion || t.Version < 1 {
        return errBadVersion
    }
    if ! t.IsFinal(blockHeight, blockTime) {
        return errNotFinal
    }
    // Uses ">=" to follow the Satoshi client
    if t.ByteSize() >= numbers.MaxStandardTxSize {
        return errStdSizeLimit
    }
    for _, txin := range t.TxIns {
        s := script.Script(txin.SigScript)
        if len(s) > numbers.MaxSigScriptSize {
            return errSigScriptSizeLimit
        }
        if !s.IsPushOnly() {
            return errSigScriptNotPushOnly
        }
        if !s.PushesCanonical() {
            return errNonCanonicalPush
        }
    }
    dataOut := 0
    for _, txout := range t.TxOuts {
        sType := script.Script(txout.PKScript).PKScriptType()
        if sType == script.PKS_NONSTANDARD {
            return errNonStandardPKScript
        } else if sType == script.PKS_NULLDATA {
            dataOut++
        } else if txout.IsDust() {
            return errDustTxOut
        }
    }
    if dataOut > 1 {
        return errMoreThanOneReturn
    }
    return nil
}

// Returns if Tx is final
func (t *Tx) IsFinal(blockHeight uint32, blockTime uint32) bool {
    if t.LockTime == 0 {
        return true
    }
    if t.LockTime < numbers.LockTimeThreshold {
        return t.LockTime < blockHeight
    } else {
        return t.LockTime < blockTime
    }
}
