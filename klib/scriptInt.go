package klib

import (
    )

const scriptIntMaxSize = 4

type ScriptInt int64

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

func (i *ScriptInt) SetBytes(p []byte) error {
    if len(p) == 0 {
        *i = 0
        return nil
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
    return nil
}
