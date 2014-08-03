package cold

import (
    "testing"
    "fmt"
    "github.com/oxfeeefeee/kaiju"
)

func TestOutputDB(t *testing.T) {
    err := kaiju.ReadJsonConfigFile()
    if err != nil {
        t.Errorf(fmt.Sprintf("Failed to read config file: %s", err))
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