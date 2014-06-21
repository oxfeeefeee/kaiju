// Manages files blockHeaders.dat, kdb.dat and blocks.dat
package blockchain

import (
    "os"
    "path/filepath"
    "github.com/oxfeeefeee/kaiju"
    )

const headersFileName = "headers.dat"

const kdbFileName = "kdb.dat"

type Files struct {
    headers     *os.File
    kdb         *os.File
    dir         string
}

var files *Files

func InitFiles() error {
    if files != nil {
        return nil
    }
    cfg := kaiju.GetConfig()
    path := filepath.Join(kaiju.GetConfigFileDir(), cfg.DataDir)
    err := os.MkdirAll(path, os.ModePerm)
    if err != nil {
        return err
    }
    hpath := filepath.Join(path, headersFileName)
    headers, err := os.OpenFile(hpath, os.O_RDWR|os.O_CREATE, os.ModePerm)
    if err != nil {
        return err
    }
    kpath := filepath.Join(path, kdbFileName)
    kdb, err := os.OpenFile(kpath, os.O_RDWR|os.O_CREATE, os.ModePerm)
    if err != nil {
        return err
    }
    files= &Files{headers, kdb, path}
    return nil
}

func CloseFiles() error {
    err := files.headers.Close()
    if err != nil {
        return err
    }
    err = files.kdb.Close()
    return err
}

func fileHeaders() *os.File {
    return files.headers
}

func fileKDB() *os.File {
    return files.kdb
}
