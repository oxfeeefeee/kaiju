package cold

import (
    "os"
    "bytes"
    "errors"
    "github.com/oxfeeefeee/kaiju"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/klib/kdb"
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

func (u *outputDB) Get(h *klib.Hash256, i uint32) ([]byte, error) {
    key := getKdbKey(h, i)
    return u.db.GetRecord(key)
}

func (u *outputDB) Use(h *klib.Hash256, i uint32, val []byte) error {
    key := getKdbKey(h, i)
    v, err := u.db.GetRecord(key)
    if err != nil {
        return err
    }
    if val != nil { // Verify val if provided
        result := bytes.Compare(v, val)
        if result != 0 {
            return errors.New("outputDB.UseOutput value doesn't match value in DB")
        }
    }
    _, err = u.db.RemoveRecord(key)
    return err
}

func (u *outputDB) Add(h *klib.Hash256, i uint32, val []byte) error {
    key := getKdbKey(h, i)
    return u.db.AddRecord(key, val)
}

func getKdbKey(h *klib.Hash256, i uint32) []byte {
    p := klib.Uint32ToBytes(i)
    return append(p, h[:]...)
}