// A facade of "blockchain/cold" to hide all the implementation
package cold

import (
    "errors"
    "github.com/oxfeeefeee/kaiju"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/catma"
)

const headersFileName = "header.dat"

const kdbFileName = "kdb.dat"

var bcFiles *files

var theHeaders *headers

var theOutputDB *outputDB

type Headers interface {
    Len() int
    Get(height int) *catma.Header
    Append(hs []*catma.Header) error 
    GetLocator() []*klib.Hash256
}

type OutputDB interface {
    HasOutput(hash *klib.Hash256, index int, value int64) bool
    UseOutput(hash *klib.Hash256, index int, value int64) error
    AddOutput(hash *klib.Hash256, index int, value int64) error
}

func Init() error {
    if bcFiles != nil || theHeaders != nil || theOutputDB != nil {
        errors.New("Init seems to be called before")
    }
    bcFiles, err := newFiles(headersFileName, kdbFileName)
    if err != nil {
        return err
    }
    theHeaders = newHeaders(bcFiles.headerFile())
    theHeaders.loadHeaders()
    theOutputDB, err = newOutputDB(bcFiles.kdbFile())
    if err != nil {
        return err
    }
    return nil
}

func Destroy() error {
    err := bcFiles.close()
    if err != nil {
        return err
    }
    theHeaders = nil
    theOutputDB = nil
    bcFiles = nil
    return nil
}

func TheHeaders() Headers {
    return theHeaders
}

func TheOutputDB() OutputDB {
    return theOutputDB
}

// Handy function
func logger() *kaiju.Logger {
    return kaiju.MainLogger()
}