package klib

import (
    //"bytes"
    "testing"
)

func encodeAndDecodeScriptInt(t *testing.T, v int64) {
    logger().Debugf("Testing ScriptInt %v ...", v)
    
    var i ScriptInt
    i = ScriptInt(v)
    i.SetBytes(i.Bytes())
    if i != ScriptInt(v) {
        t.Errorf("ScriptInt %v encode decode error %v", v, i)
    }
}

func TestScriptInt(t *testing.T) {
    encodeAndDecodeScriptInt(t, 123456);
    encodeAndDecodeScriptInt(t, -123456);
    //encodeAndDecodeScriptInt(t, 12345600000);
    //encodeAndDecodeScriptInt(t, -12345600000);
}