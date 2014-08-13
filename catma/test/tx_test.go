package test

import (
    "bytes"
    "testing"
    "encoding/json"
    "encoding/hex"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/knet/btcmsg"
    "github.com/oxfeeefeee/kaiju/catma"
    "github.com/oxfeeefeee/kaiju/catma/script"
    "fmt"
    "strings"
    //"strconv"
    //"errors"
    //"encoding/hex"
)

type prevOutput struct {
    prevHash klib.Hash256
    prevIndex uint32
    pkScript []byte
}

func (p *prevOutput) String() string {
    return fmt.Sprintf("Phash: %s[%d]\n PKScript: %s\n", p.prevHash, p.prevIndex, p.pkScript)
}

type prevOutputs []*prevOutput

func (p *prevOutputs) GetTxOut(op *catma.OutPoint) *catma.TxOut {
    for _, out := range *p {
        if op.Hash == out.prevHash && op.Index == out.prevIndex {
            return &catma.TxOut{0, out.pkScript}
        }
    }
    return nil
}

type txTestCase struct {
    pos prevOutputs
    tx *catma.Tx
    flags script.EvalFlag
}

func (c *txTestCase) valid() error {
    err := c.tx.FormatCheck()
    if err != nil {
        return err
    }

    for i, txin := range c.tx.TxIns {
        txo := c.pos.GetTxOut(&txin.PreviousOutput)
        err := catma.VerifyInputWithFlags(txo.PKScript, c.tx, i, c.flags)
        if err != nil {
            //fmt.Printf("tx Fail!\n")
            return err
        }
    }
    //fmt.Printf("tx OK!\n")
    return nil
}

func parseTestCase(t *testing.T, data interface{}) *txTestCase {
    srcStr := data.([]interface{})
    if len(srcStr) != 3 {
        return nil
    }
    prevOuts := parsePrevOutputs(t, srcStr[0])
    tx := parseTx(t, srcStr[1])
    flags := parseFlags(srcStr[2])
    return &txTestCase{prevOuts, &tx, flags,}
}

func parsePrevOutputs(t *testing.T, data interface{}) []*prevOutput {
    prevOuts := make([]*prevOutput, 0)
    inputs := data.([]interface{})
    for _, input := range inputs {
        ip := input.([]interface{})
        var po prevOutput
        po.prevHash.SetString(ip[0].(string))
        po.prevIndex = uint32(ip[1].(float64))
        script, err := parseScript(ip[2].(string))
        if err != nil {
            t.Errorf("error parsing script %s, data: %s", err, ip[2].(string))
        }
        po.pkScript = script
        prevOuts = append(prevOuts, &po)
    }
    return prevOuts
}

func parseTx(t *testing.T, data interface{}) catma.Tx {
    txData, err := hex.DecodeString(data.(string))
    if err != nil {
        t.Errorf("hex.DecodeString error %s", err)
    }
    var btx btcmsg.Tx
    r := bytes.NewReader(txData)
    err = btx.Deserialize(r)
    if err != nil {
        t.Errorf("tx deserialize error %s, data: %s", err, data.(string))
    }
    return catma.Tx(btx)
}

func parseFlags(data interface{}) script.EvalFlag {
    flagsStr := data.(string)
    fs := strings.Split(flagsStr, ",")
    flag := script.EvalFlagNone
    for _, f := range fs {
        switch f {
        case "P2SH": 
            flag |= script.EvalFlagP2SH
        case "NULLDUMMY":
            flag |= script.EvalFlagNullDummy
        }
    }
    return flag
} 

func TestTx(t *testing.T) {
    var f interface{}
    err := json.Unmarshal([]byte(validTxs), &f)
    if err != nil {
        t.Errorf("json.Unmarshal error %s", err)
    }
    cases := f.([]interface{})
    for _, c := range cases {
        tc := parseTestCase(t, c)
        if tc != nil {
            //fmt.Printf("Testing: %v\n", c)
            err = tc.valid()
            if err != nil {
                t.Errorf("valid tx deemed to be invalid: %s", err)
            }
        }
    }

    err = json.Unmarshal([]byte(invalidTxs), &f)
    if err != nil {
        t.Errorf("json.Unmarshal error %s", err)
    }
    cases = f.([]interface{})
    for _, c := range cases {
        tc := parseTestCase(t, c)
        if tc != nil {
            err = tc.valid()
            if err == nil {
                fmt.Printf("Testing: %v\n", c)
                t.Errorf("invalid tx deemed to be valid")
            }
        }
    }
}