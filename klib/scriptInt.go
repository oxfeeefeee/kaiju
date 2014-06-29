package klib

import (
    )

const scriptIntMaxSize = 4

type ScriptInt int64

func ToScriptInt(p []byte) ScriptInt {
    var i ScriptInt
    i.SetBytes(p)
    return i
}

func ScriptIntOverflow(p []byte) bool {
    return len(p) > scriptIntMaxSize
}

func (i ScriptInt) Bytes() []byte {
    if i == 0 {
        return []byte{} 
    } 
    
    p := make([]byte, 0)
    neg := i < 0
    abs := i
    if neg {
        abs = -i
    }
    for abs != 0 {
        p = append(p, byte(abs))
        abs >>= 8
    }
    // Let (most_significant_byte & 0x80) = 0 when i is positive 
    if (p[len(p)-1] & 0x80) != 0 {
        if neg {
            p = append(p, 0x80)    
        } else {
            p = append(p, 0x00) 
        }
    } else if neg {
        p[len(p)-1] |= 0x80
    }
    return p
}

// Check overflow before call this, otherwise it could panic
func (i *ScriptInt) SetBytes(p []byte) {
    if len(p) == 0 {
        *i = 0
        return
    } 
    if len(p) > scriptIntMaxSize {
        panic("ScriptInt.SetBytes overflow")
    }
    var val int64
    for j := 0; j < len(p); j++ {
        val |= int64(p[j]) << uint(j * 8)
    }
    lb := p[len(p)-1]
    if (lb & 0x80) != 0 {
        val = -(val & ^(0x80 << uint((len(p) - 1) * 8)))
    }
    *i = ScriptInt(val)
}
