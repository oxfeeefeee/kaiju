package btcmsg

import (
    "bytes"
    "testing"
)

func encodeAndDecodeVarUnit(t *testing.T, v uint64) {
    logger().Debugf("Testing VarUint %v ...", v)
    buf := new(bytes.Buffer)
    var err error;
    vi := VarUint(v)
    err = writeVarUint(buf, &vi, err)
    if err != nil {
        t.Errorf("Encode error : %s", err.Error())
    }
    logger().Debugf("Encoded: %v", buf.Bytes())
    err = readVarUint(buf, &vi, err)
    if err != nil {
        t.Errorf("Decode error : %s", err.Error())
    }
    if vi != VarUint(v) {
        t.Errorf("VarUint %v encode decode error", v)
    }
}

func TestVarUint(t *testing.T) {
    encodeAndDecodeVarUnit(t, 123)
    encodeAndDecodeVarUnit(t, 0xfffe)
    encodeAndDecodeVarUnit(t, 0xffff)
    encodeAndDecodeVarUnit(t, 65536)
    encodeAndDecodeVarUnit(t, 1234567890)
    encodeAndDecodeVarUnit(t, 1234567890123)
}