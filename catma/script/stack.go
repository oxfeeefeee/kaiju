package script

import (
    "github.com/oxfeeefeee/kaiju/klib"
    )

type scriptNum klib.ScriptInt 

func (sn *scriptNum) toStackItem() stackItem {
    return (*klib.ScriptInt)(sn).Bytes()
}

type stackItem []byte

func (si stackItem) toBool() bool {
    for i, v := range si {
        if v != 0 {
            if v != 0x80 {
                return true
            } else {
                return i == (len(si) - 1) // negative zero
            }
        }
    }
    return false
}

type stack []stackItem

func (s *stack) push(i stackItem) {
    *s = append(*s, i)
}

func (s *stack) height() int {
    return len(*s)
}

func (s *stack) empty() bool {
    return len(*s) == 0
}

func (s stack) top() stackItem {
    return s[len(s) - 1]
}

func (s *stack) pop() stackItem {
    stk := *s
    v := stk[len(stk)-1]
    *s = stk[:len(stk)-1]
    return v
}

// True/Fase stack used to help handle OP_IF, OP_ELSE ...
// it records the conditions used by OP_IF, OP_ELSE ...
// the height of the stack means how many level of IFs we're in.
type boolStack []bool

// Returns false if not all true on stack, othewise returns true
// If not alive, we are at the non-execution side of branching
// but we still need to go through the code.
func (bs *boolStack) alive() bool {
    for _, v := range *bs {
        if !v {
            return false
        }
    }
    return true
}

func (bs *boolStack) push(v bool) {
    *bs = append(*bs, v)
}

// Check not empty before call this
func (bs *boolStack) pop() bool {
    stk := *bs
    v := stk[len(stk)-1]
    *bs = stk[:len(stk)-1]
    return v
}

func (bs *boolStack) empty() bool {
    return len(*bs) == 0
}

func (bs *boolStack) height() int {
    return len(*bs)
}

