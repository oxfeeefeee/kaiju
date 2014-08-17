// A facade of "blockchain/cold" to hide all the implementation
package cold

import (
    "os"
    "path/filepath"
    "github.com/oxfeeefeee/kaiju"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/catma"
    "github.com/oxfeeefeee/kaiju/klib/kdb"
)

const headersFileName = "header.dat"

const kdbFileName = "kdb.dat"

var cold Cold

type HeaderArray interface {
    Len() int
    Get(height int) *catma.Header
    Append(hs []*catma.Header) error 
    GetLocator() []*klib.Hash256
}

type UtxoDB interface {
    catma.UtxoSet
    Commit(tag uint32) error
    Tag() (uint32, error)
}

type Cold struct {
    hfile   *os.File
    dbFile  *os.File
    h       *headers
    db      *outputDB
}

func Get() *Cold {
    return &cold
}

func (c *Cold) Init() error {
    path, err := initFilePath()
    if err != nil {
        return err
    }

    f, _, err := openFile(path, headersFileName)
    if err != nil {
        return err
    }
    c.hfile = f
    c.h = newHeaders(f)
    c.h.loadHeaders()

    f, fi, err := openFile(path, kdbFileName)
    if err != nil {
        return err
    }
    var db *kdb.KDB
    if fi.Size() == 0 {
        db, err = kdb.New(kaiju.KDBCapacity, f)
        if err != nil {
            return err
        }
    } else {
        db, err = kdb.Load(f)
        if err != nil {
            return err
        }
    }
    c.dbFile = f
    c.db = newOutputDB(db)
    return nil
}

func (c *Cold) Destroy() error {
    if err := c.hfile.Close(); err != nil {
        return err
    }
    if err := c.dbFile.Close(); err != nil {
        return err
    }
    c.hfile = nil
    c.dbFile = nil
    c.h = nil
    c.db = nil
    return nil
}

func (c *Cold) Headers() HeaderArray {
    return c.h
}

func (c *Cold) OutputDB() UtxoDB {
    return c.db
}

func initFilePath() (string ,error) {
    cfg := kaiju.GetConfig()
    path := filepath.Join(kaiju.ConfigFileDir(), cfg.DataDir)
    if err := os.MkdirAll(path, os.ModePerm); err != nil {
        return "", err
    } else {
        return path, nil
    }
}

func openFile(path string, name string) (*os.File, os.FileInfo, error) {
    fullp := filepath.Join(path, name)
    f, err := os.OpenFile(fullp, os.O_RDWR|os.O_CREATE, os.ModePerm)
    if err != nil {
        return nil, nil, err
    }
    if fi, err := f.Stat(); err != nil {
        return nil, nil, err
    } else {
        return f, fi, err
    }
}
