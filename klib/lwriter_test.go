package klib

import (
    "bytes"
    "testing"
)

func TestLWriter(t *testing.T) {
    var buf bytes.Buffer

    lw := NewLWriter(&buf,5)
    n, err := lw.Write(make([]byte, 5))
    if n != 5 || err != nil {
        t.Errorf("write error")
    }
    n, err = lw.Write(make([]byte, 1))
    if err == nil {
        t.Errorf("should error but not")
    } else {
        t.Logf("output: %v %v", n, err)
    }
}