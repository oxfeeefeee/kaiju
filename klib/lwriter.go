package klib

import (
    "io"
    )

// LWriter stands for limited writer.
// LWriter is used to wrap another writer and limit the amount of data 
// to write
// LWriter returns io.EOF when limit is reached.

type LWriter struct { 
    w io.Writer
    limit int
    written int
}

func NewLWriter(w io.Writer, l int) *LWriter {
    return &LWriter{w, l, 0,}
}

func (lw *LWriter) Write(p []byte) (int, error) {
    if len(p) > (lw.limit - lw.written) {
        return 0, io.EOF
    } else {
        n, err := lw.w.Write(p)
        if err == nil {
            lw.written += n 
        }
        return n, err
    }
}