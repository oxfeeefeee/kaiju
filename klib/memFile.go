package klib

import (
    "io"
    "os"
    "errors"
    )

type MemFile struct {
    buf         []byte
    off         int
}

func NewMemFile(size int64) *MemFile {
    return &MemFile{
        make([]byte, size, size),
        0,
    }
}

func (f *MemFile) Read(p []byte) (n int, err error) {
    fl, bl := len(f.buf), len(p)
    if bl > (fl - f.off) {
        return 0, io.EOF
    } else {
        copy(p, f.buf[f.off:])
        f.off += bl
        return bl, nil
    }
}

func (f *MemFile) Seek(offset int64, whence int) (int64, error) {
    fl := len(f.buf)
    if whence == os.SEEK_SET {
        offset = offset
    } else if whence == os.SEEK_CUR {
        offset += int64(f.off)
    } else if whence == os.SEEK_END {
        offset = int64(fl) + offset
    } else {
        return offset, errors.New("MemFile Seek out of range.")
    }

    if offset < int64(fl) {
        f.off = int(offset)
        return int64(f.off), nil
    } else {
        return int64(f.off), io.EOF
    }
}

func (f *MemFile) Write(p []byte) (n int, err error) {
    fl, bl := len(f.buf), len(p)
    if bl > (fl - f.off) {
        return 0, errors.New("MemFile Write out of range.")
    } else {
        copy(f.buf[f.off:], p)
        f.off += bl
        return bl, nil
    }
}

func (f *MemFile) Sync() (err error) {
    return nil
}