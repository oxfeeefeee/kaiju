package cold

import (
    "os"
    "github.com/oxfeeefeee/kaiju"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/klib/kdb"
    //"github.com/oxfeeefeee/kaiju/catma"
)

// All unspent tx output stored in KDB
type outputDB struct {
    db  *kdb.KDB
}

func newOutputDB(f *os.File) (*outputDB, error) {
    db, err := kdb.New(kaiju.KDBCapacity, f)
    if err != nil {
        return nil, err
    }
    return &outputDB{db,}, nil
}

func (u *outputDB) HasOutput(hash *klib.Hash256, index int, value int64) bool {
    return false
}

func (u *outputDB) UseOutput(hash *klib.Hash256, index int, value int64) error {
    return nil
}

func (u *outputDB) AddOutput(hash *klib.Hash256, index int, value int64) error {
    return nil
}