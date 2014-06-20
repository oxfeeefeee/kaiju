package klib

import (
    //"bytes"
    "testing"
)

func encodeAndDecodeScriptInt(t *testing.T, v int64) {
    logger().Debugf("Testing ScriptInt %v ...", v)
    
    var i ScriptInt
    i = ScriptInt(v)
    err := i.SetBytes(i.Bytes())
    if err != nil {
        t.Errorf("Decode error : %s", err.Error())
    }
    if i != ScriptInt(v) {
        t.Errorf("ScriptInt %v encode decode error %v", v, i)
    }
}

func TestScriptInt(t *testing.T) {
    encodeAndDecodeScriptInt(t, 123456);
    encodeAndDecodeScriptInt(t, -123456);
    encodeAndDecodeScriptInt(t, 12345600000);
    encodeAndDecodeScriptInt(t, -12345600000);
}