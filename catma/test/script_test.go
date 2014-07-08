package test

import (
    "testing"
    "fmt"
    "strings"
    "strconv"
    "errors"
    "encoding/hex"
    "encoding/json"
    "github.com/oxfeeefeee/kaiju/catma/script"
)

var opcodeNameMap map[string]script.Opcode

func nameMap() map[string]script.Opcode {
    if opcodeNameMap != nil {
        return opcodeNameMap
    }

    m := map[string]script.Opcode{"RESERVED": script.OP_RESERVED,}
    for op := script.OP_NOP; op <= script.OP_NOP10; op++ {
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

func parseScript(s string) (script.Script, error) {
    m := nameMap()
    scr := script.NewScript()
    words := strings.FieldsFunc(s, func (r rune) bool {
        return r == ' ' || r == '\n' || r == '\t'
    })
    for _, w := range words {
        if allDigit(w) || (w[0] == '-' && allDigit(w[1:])) {
            n, err := strconv.Atoi(w)
            if err != nil {
                return nil, errors.New(fmt.Sprintf("Error parsing integer: %s", w))
            }
            scr.AppendPushInt(int64(n))
        } else if len(w) > 2 && w[:2] == "0x" {
            data, err := hex.DecodeString(w[2:])
            if err != nil {
                return nil, errors.New(fmt.Sprintf("Error parsing hex: %s", w))
            }
            scr.AppendData(data)
        } else if len(w) >= 2 && w[0] == '\'' && w[len(w)-1] == '\'' {
            scr.AppendPushData([]byte(w[1:len(w)-1]))
        } else if op, ok := m[w]; ok {
            scr.AppendOp(op)
        } else {
            return nil, errors.New(fmt.Sprintf("Error parsing unknown word: %s", w))
        }
    }
    return *scr, nil
}

func testScriptList(t *testing.T, list string, valid bool) {
    var f interface{}
    err := json.Unmarshal([]byte(list), &f)
    if err != nil {
        t.Errorf("json.Unmarshal error %s", err)
    }
    scrStrList := f.([]interface{})
    for _, scrStr := range scrStrList {
        singleCase := scrStr.([]interface{})
        if len(singleCase) < 2 {
            continue
        }
        sigS, errSig := parseScript(singleCase[0].(string))
        pkS, errPK := parseScript(singleCase[1].(string))
        if errSig != nil || errPK != nil {
            t.Errorf("parse script error %s; %s", errSig, errPK)
        }
        //fmt.Printf("Running: %s--%s\n", singleCase[0].(string), singleCase[1].(string))
        err = script.RunScript(pkS, sigS)
        if valid && err != nil {
            t.Errorf("Run valid script error %s \n CONTENT:%s--%s", err, singleCase[0].(string), singleCase[1].(string))
        } else if !valid && err == nil {
            t.Errorf("Run invalid script passed, CONTENT:%s--%s", singleCase[0].(string), singleCase[1].(string))
        } else {
            //fmt.Printf("Script OK, test: %b\n", valid)
        }
    }
}

func TestScript(t *testing.T) {
    testScriptList(t, validScripts, true)
    testScriptList(t, invalidScripts, false)
}

func TestPKScriptType(t *testing.T) {
    for str, stype := range testScripts() {
        scr, err := parseScript(str)
        if err != nil {
            t.Errorf("TestPKScriptType error: %s", err)
        } else if st := scr.PKSType(); st != stype {
            t.Errorf("Expect type %v, got %v", stype, st)
        }
    }
}

func TestOpcodeName(t *testing.T) {
    if "OP_PUSHDATA14" != script.OP_PUSHDATA14.String() {
        t.Errorf("Bad name for Opcode %s", script.OP_PUSHDATA14)
    }
    if "OP_PUSHDATA1" != script.OP_PUSHDATA1.String() {
        t.Errorf("Bad name for Opcode %s", script.OP_PUSHDATA1)
    }
    if "OP_CODESEPARATOR" != script.OP_CODESEPARATOR.String() {
        t.Errorf("Bad name for Opcode %s", script.OP_CODESEPARATOR)
    }
    if "OP_INVALIDOPCODE" != script.OP_INVALIDOPCODE.String() {
        t.Errorf("Bad name for Opcode %s", script.OP_INVALIDOPCODE)
    }

    for i := 0; i < int(script.OP_NOP10); i++ {
        //fmt.Printf("value: %x, Opcode: %s\n", i, Opcode(i))
    }
}
