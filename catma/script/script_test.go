package script

import (
    "testing"
    "fmt"
    "strings"
    "strconv"
    "errors"
    "encoding/hex"
)

var opcodeNameMap map[string]Opcode

func nameMap() map[string]Opcode {
    if opcodeNameMap != nil {
        return opcodeNameMap
    }

    m := map[string]Opcode{"RESERVED": OP_RESERVED,}
    for op := OP_NOP; op <= OP_NOP10; op++ {
        name := op.String()
        if name == "OP_UNKNOWN" {
            continue
        }
        m[name] = op
        m[name[3:]] = op
    }
    opcodeNameMap = m
    return opcodeNameMap
}

func allDigit(s string) bool {
    for _, c := range s {
        if c < '0' || c > '9' {
            return false
        }
    }
    return true
}

func parseScript(s string) (Script, error) {
    m := nameMap()
    script := NewScript()
    words := strings.FieldsFunc(s, func (r rune) bool {
        return r == ' ' || r == '\n' || r == '\t'
    })
    for _, w := range words {
        if allDigit(w) || (w[0] == '-' && allDigit(w[1:])) {
            n, err := strconv.Atoi(w)
            if err != nil {
                return nil, errors.New(fmt.Sprintf("Error parsing integer: %s", w))
            }
            script.AppendPushInt(int64(n))
        } else if len(w) > 2 && w[:2] == "0x" {
            data, err := hex.DecodeString(w[2:])
            if err != nil {
                return nil, errors.New(fmt.Sprintf("Error parsing hex: %s", w))
            }
            script.AppendData(data)
        } else if len(w) >= 2 && w[0] == '\'' && w[len(w)-1] == '\'' {
            script.AppendPushData([]byte(w[1:len(w)-1]))
        } else if op, ok := m[w]; ok {
            script.AppendOp(op)
        } else {
            return nil, errors.New(fmt.Sprintf("Error parsing unknown word: %s", w))
        }
    }
    return *script, nil
}

func TestPKScriptType(t *testing.T) {
    for str, stype := range testScripts() {
        script, err := parseScript(str)
        if err != nil {
            t.Errorf("TestPKScriptType error: %s", err)
        } else if st := script.Type(); st != stype {
            t.Errorf("Expect type %v, got %v", stype, st)
        }
    }
}

func TestOpcodeName(t *testing.T) {
    if "OP_PUSHDATA14" != OP_PUSHDATA14.String() {
        t.Errorf("Bad name for Opcode %s", OP_PUSHDATA14)
    }
    if "OP_PUSHDATA1" != OP_PUSHDATA1.String() {
        t.Errorf("Bad name for Opcode %s", OP_PUSHDATA1)
    }
    if "OP_CODESEPARATOR" != OP_CODESEPARATOR.String() {
        t.Errorf("Bad name for Opcode %s", OP_CODESEPARATOR)
    }
    if "OP_INVALIDOPCODE" != OP_INVALIDOPCODE.String() {
        t.Errorf("Bad name for Opcode %s", OP_INVALIDOPCODE)
    }

    for i := 0; i < int(OP_NOP10); i++ {
        //fmt.Printf("value: %x, Opcode: %s\n", i, Opcode(i))
    }
}
