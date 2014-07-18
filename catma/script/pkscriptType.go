package script

import (
    "github.com/oxfeeefeee/kaiju/catma/numbers"
    )

type PKScriptType byte

const (
    PKS_NonStandard PKScriptType = iota
    PKS_PubKey
    PKS_PubKeyHash
    PKS_ScriptHash
    PKS_MultiSig
    PKS_NullData
)

func (t PKScriptType) String() string {
    switch t {
    case PKS_NonStandard:   return "PKS_NonStandard"
    case PKS_PubKey:        return "PKS_PubKey"
    case PKS_PubKeyHash:    return "PKS_PubKeyHash"
    case PKS_ScriptHash:    return "PKS_ScriptHash"
    case PKS_MultiSig:      return "PKS_MultiSig"
    case PKS_NullData:      return "PKS_NullData"
    default:                return "PKS_Invalid"
    }
}

func (s Script) PKScriptType() PKScriptType {
    switch {
    case s.IsTypePubKey():      return PKS_PubKey
    case s.IsTypePubKeyHash():  return PKS_PubKeyHash
    case s.IsTypeScriptHash():  return PKS_ScriptHash
    case s.IsTypeNullData():    return PKS_NullData
    case s.IsTypeMultiSig():    return PKS_MultiSig
    default:                    return PKS_NonStandard
    }
}

// Returns if PKScipt is of type PKS_PubKey
func (s Script) IsTypePubKey() bool {
    op, operand, next, err := s.getOpcode(0)
    l := len(operand)
    if err != nil || l < numbers.MinPubKeyLen || l > numbers.MaxPubKeyLen {
        return false
    }
    op, _, next, err = s.getOpcode(next)
    if err != nil || op != OP_CHECKSIG {
        return false
    }
    return next == len(s)
}

// Returns if PKScipt is of type PKS_PubKeyHash
func (s Script) IsTypePubKeyHash() bool {
    op, _, next, err := s.getOpcode(0)
    if err != nil || op != OP_DUP {
        return false
    }
    op, _, next, err = s.getOpcode(next)
    if err != nil || op != OP_HASH160 {
        return false
    }
    op, operand, next, err := s.getOpcode(next)
    if err != nil || len(operand) != numbers.PubKeyHashLen {
        return false
    }
    op, _, next, err = s.getOpcode(next)
    if err != nil || op != OP_EQUALVERIFY {
        return false
    }
    op, _, next, err = s.getOpcode(next)
    if err != nil || op != OP_CHECKSIG {
        return false
    }
    return next == len(s)
}

// Returns if PKScipt is of type PKS_ScriptHash
func (s Script) IsTypeScriptHash() bool {
    return len(s) == 23 &&
        Opcode(s[0])  == OP_HASH160 &&
        Opcode(s[1])  == OP_PUSHDATA14 &&
        Opcode(s[22]) == OP_EQUAL
}

// Returns if PKScipt is of type PKS_MultiSig
func (s Script) IsTypeMultiSig() bool {
    if len(s) < 1 {
        return false
    }

    next := 0
    m := Opcode(s[next]).number()
    if m < 0 {
        return false
    }

    next += 1
    count := 0
    n := 0
    for {
        op, operand, np, err := s.getOpcode(next)
        next = np
        l := len(operand)
        if err == nil {
            if n = op.number(); n >= 0 { // It's OP_N
                break
            } else if l >= numbers.MinPubKeyLen && l <= numbers.MaxPubKeyLen {
                count += 1
            } else {
                return false
            }
        } else {
            return false
        }
    }
    if n < 0 {
        return false
    }

    op, _, next, err := s.getOpcode(next)
    if err != nil || op != OP_CHECKMULTISIG {
        return false
    }

    return m >=1 && n >= 1 && m <= n && count == n && next == len(s)
}

// Returns if PKScipt is of type PKS_NullData
func (s Script) IsTypeNullData() bool {
    if len(s) < 1 || Opcode(s[0]) != OP_RETURN {
        return false
    }
    if len(s) == 1 {
        return true
    } else {
        _, operand, next, err := s.getOpcode(1)
        return err != nil && 
            len(operand) <= numbers.MaxOpReturnRelay &&  
            next == len(s)
    }
}

