package script

import (
    "github.com/oxfeeefeee/kaiju/catma/numbers"
    )

type PKScriptType byte

const (
    PKS_NONSTANDARD PKScriptType = iota
    PKS_PUBKEY
    PKS_PUBKEYHASH
    PKS_SCRIPTHASH
    PKS_MULTISIG
    PKS_NULLDATA
)

func (t PKScriptType) String() string {
    switch t {
    case PKS_NONSTANDARD:   return "PKS_NONSTANDARD"
    case PKS_PUBKEY:        return "PKS_PUBKEY"
    case PKS_PUBKEYHASH:    return "PKS_PUBKEYHASH"
    case PKS_SCRIPTHASH:    return "PKS_SCRIPTHASH"
    case PKS_MULTISIG:      return "PKS_MULTISIG"
    case PKS_NULLDATA:      return "PKS_NULLDATA"
    default:                return "PKS_INVALID"
    }
}

func (s Script) PKScriptType() PKScriptType {
    switch {
    case s.IsTypePubKey():      return PKS_PUBKEY
    case s.IsTypePubKeyHash():  return PKS_PUBKEYHASH
    case s.IsTypeScriptHash():  return PKS_SCRIPTHASH
    case s.IsTypeNullData():    return PKS_NULLDATA
    case s.IsTypeMultiSig():    return PKS_MULTISIG
    default:                    return PKS_NONSTANDARD
    }
}

// Returns if PKScipt is of type PKS_PUBKEY
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

// Returns if PKScipt is of type PKS_PUBKEYHASH
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

// Returns if PKScipt is of type PKS_SCRIPTHASH
func (s Script) IsTypeScriptHash() bool {
    return len(s) == 23 &&
        Opcode(s[0])  == OP_HASH160 &&
        Opcode(s[1])  == OP_PUSHDATA14 &&
        Opcode(s[22]) == OP_EQUAL
}

// Returns if PKScipt is of type PKS_MULTISIG
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

// Returns if PKScipt is of type PKS_NULLDATA
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

