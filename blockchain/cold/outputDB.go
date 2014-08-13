package cold

import (
    "bytes"
    "errors"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/klib/kdb"
)

// All unspent tx output stored in KDB
type outputDB struct {
    db  *kdb.KDB
}

func newOutputDB(db *kdb.KDB) *outputDB {
    return &outputDB{db,}
}

func (u *outputDB) Get(h *klib.Hash256, i uint32) ([]byte, error) {
    key := getKdbKey(h, i)
    return u.db.Get(key)
}

func (u *outputDB) Use(h *klib.Hash256, i uint32, val []byte) error {
    key := getKdbKey(h, i)
    v, err := u.db.Get(key)
    if err != nil {
        return err
    }
    if val != nil { // Verify val if provided
        result := bytes.Compare(v, val)
        if result != 0 {
            return errors.New("outputDB.UseOutput value doesn't match value in DB")
        }
    }
    _, err = u.db.Remove(key)
    return err
}

func (u *outputDB) Add(h *klib.Hash256, i uint32, val []byte) error {
    key := getKdbKey(h, i)
    return u.db.Add(key, val)
}

func (u *outputDB) Commit(tag uint32) error {
    return u.db.Commit(tag)
}

func (u *outputDB) Tag() (uint32, error) {
    return u.db.Tag()
}

func getKdbKey(h *klib.Hash256, i uint32) []byte {
    p := klib.Uint32ToBytes(i)
    return append(p, h[:]...)
}