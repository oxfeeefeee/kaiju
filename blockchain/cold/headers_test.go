package cold

import (
    "testing"
    "fmt"
    "github.com/oxfeeefeee/kaiju"
)

func TestFiles(t *testing.T) {
    err := kaiju.Init()
    if err != nil {
        t.Errorf(fmt.Sprintf("Failed to call kaiju.Init: %s", err))
    }

    fs, err := newFiles(headersFileName, kdbFileName)
    if err != nil {
        t.Errorf(fmt.Sprintf("newFiles error: %s", err))
    }
    //fs.headerFile().Write([]byte{0,1,2,3,4,5,6,78})
    fs.close()
}

func TestGenesisHeader(t *testing.T) {
    fs, err := newFiles(headersFileName, kdbFileName)
    if err != nil {
        t.Errorf(fmt.Sprintf("newFiles error: %s", err))
    }
    headers := newHeaders(fs.headerFile())
    h := headers.data[0]
    s := h.Hash().String()
    log.Debugf("genesis hash %s", s)
    if s != "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f" {
        t.Errorf("Invalid genesis hash")
    }

    log.Debugf("Locator %s", headers.GetLocator()[0]) 

    fs.close()
}