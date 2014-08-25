// A facade of "blockchain/storage" to hide all the implementation
package storage

import (
    "os"
    "path/filepath"
    "github.com/oxfeeefeee/kaiju"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/catma"
    "github.com/oxfeeefeee/kaiju/klib/kdb"
)

var storage Storage

type HeaderArray interface {
    Len() int
    Get(height int) *catma.Header
    Append(hs []*catma.Header) error 
    GetLocator() []*klib.Hash256
}

type UtxoDB interface {
    catma.UtxoSet
    Commit(tag uint32, force bool) error
    Tag() (uint32, error)
}

type Storage struct {
    hfile   *os.File
    dbFile  *os.File
    waFile  *os.File
    h       *headers
    db      *outputDB
}

func Get() *Storage {
    return &storage
}

func (c *Storage) Init() error {
    path, err := initFilePath()
    if err != nil {
        return err
    }

    f, _, err := openFile(path, kaiju.GetConfig().HeadersFileName)
    if err != nil {
        return err
    }
    c.hfile = f
    c.h = newHeaders(f)
    c.h.loadHeaders()

    dbf, dbfi, err := openFile(path, kaiju.GetConfig().KdbFileName)
    if err != nil {
        return err
    }
    waf, _, err := openFile(path, kaiju.GetConfig().KdbWAFileName)
    if err != nil {
        return err
    }
    var db *kdb.KDB
    if dbfi.Size() == 0 {
        db, err = kdb.New(kaiju.GetConfig().KDBCapacity, dbf, waf)
        if err != nil {
            return err
        }
    } else {
        db, err = kdb.Load(dbf, waf)
        if err != nil {
            return err
        }
    }
    c.dbFile = dbf
    c.waFile = waf
    c.db = newOutputDB(db)
    return nil
}

func (c *Storage) Destroy() error {
    if err := c.hfile.Close(); err != nil {
        return err
    }
    if err := c.dbFile.Close(); err != nil {
        return err
    }
    c.hfile, c.dbFile, c.waFile = nil, nil, nil
    c.h, c.db = nil, nil
    return nil
}

func (c *Storage) Headers() HeaderArray {
    return c.h
}

func (c *Storage) OutputDB() UtxoDB {
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
