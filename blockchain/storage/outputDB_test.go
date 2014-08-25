package storage

import (
    "testing"
    "fmt"
    "github.com/oxfeeefeee/kaiju"
)

func TestOutputDB(t *testing.T) {
    err := kaiju.Init()
    if err != nil {
        t.Errorf(fmt.Sprintf("Failed to call kaiju.Init: %s", err))
    }

    fs, err := newFiles(headersFileName, kdbFileName)
    if err != nil {
        t.Errorf(fmt.Sprintf("newFiles error: %s", err))
    }

    _, err = newOutputDB(fs.kdbFile())
    if err != nil {
        t.Errorf(fmt.Sprintf("newUtxoDB error: %s", err))
    }

    fs.close()
}