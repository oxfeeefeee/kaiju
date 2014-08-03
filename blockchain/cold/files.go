// Manages files blockHeaders.dat and kdb.dat
package cold

import (
    "os"
    "path/filepath"
    "github.com/oxfeeefeee/kaiju"
    )

type files struct {
    header      *os.File
    kdb         *os.File
    dir         string
}

func newFiles(hname string, dbname string) (*files, error) {
    cfg := kaiju.GetConfig()
    path := filepath.Join(kaiju.GetConfigFileDir(), cfg.DataDir)
    err := os.MkdirAll(path, os.ModePerm)
    if err != nil {
        return nil, err
    }
    hpath := filepath.Join(path, hname)
    header, err := os.OpenFile(hpath, os.O_RDWR|os.O_CREATE, os.ModePerm)
    if err != nil {
        return nil, err
    }
    kpath := filepath.Join(path, dbname)
    kdb, err := os.OpenFile(kpath, os.O_RDWR|os.O_CREATE, os.ModePerm)
    if err != nil {
        return nil, err
    }
    return &files{header, kdb, path}, nil
}

func (f *files) headerFile() *os.File {
    return f.header
}

func (f *files) kdbFile() *os.File {
    return f.kdb
}

func (f *files) close() error {
    err := f.header.Close()
    if err != nil {
        return err
    }
    err = f.kdb.Close()
    return err
}
