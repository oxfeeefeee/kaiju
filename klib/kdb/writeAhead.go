package kdb

import (
    "io"
    "errors"
    "encoding/gob"
    "encoding/binary"
    )

// Write-ahead data, 
type waData struct {
    Keys        map[int64]keyData
    ValData     []byte
}

func (wa *waData) save(w io.Writer) error {
    enc := gob.NewEncoder(w)
    return enc.Encode(wa)
}

func (wa *waData) load(r io.Reader) error {
    dec := gob.NewDecoder(r)
    return dec.Decode(wa)
}

func (wa *waData) addKey(key keyData, slotNum int64) {
    wa.Keys[slotNum] = key
}

func (wa *waData) getKey(slotNum int64) (keyData, bool) {
    k, ok := wa.Keys[slotNum]
    return k, ok
}

func (wa *waData) addValue(value []byte) {
    wa.ValData = append(wa.ValData, value...)
}

func (wa *waData) clear() {
    wa.Keys = make(map[int64]keyData)
    wa.ValData = make([]byte,0)
}

func (db *KDB) saveWAData() error {
    if _, err := db.storage.Seek(HeaderSize, 0); err != nil {
        return err
    }
    if err := db.wa.save(db.storage); err != nil {
        return err
    }
    return db.storage.Sync()
}

func (db *KDB) loadWAData() error {
    if _, err := db.storage.Seek(HeaderSize, 0); err != nil {
        return err
    }
    return db.wa.load(db.storage)
}

func (db *KDB) commitWAData() error {
    for n, k := range db.wa.Keys {
        if _, err := writeAt(db.storage, db.slotsBeginPos() + n * SlotSize, k); err != nil {
            return err
        }
    }
    n, err := writeAt(db.storage, db.cursor, db.wa.ValData)
    if err != nil {
        return err
    }
    if err := db.storage.Sync(); err != nil {
        return err
    }
    db.cursor += n
    db.wa.clear()
    return nil
}

// Write begin_commit_tag
func (db *KDB) beginWACommit(tag uint32) error {
    p := make([]byte, 4)
    if _, err := readAt(db.storage, beginCommitTagPos, p); err != nil {
        return err
    }
    oldTag := binary.LittleEndian.Uint32(p)
    if tag == oldTag {
        return errors.New("Commit tag is the same as old tag.")
    }
    binary.LittleEndian.PutUint32(p, tag)
    if _, err := writeAt(db.storage, beginCommitTagPos, p); err != nil {
        return err
    }
    return db.storage.Sync()
}

// Write end_commit_tag
func (db *KDB) endWACommit(tag uint32) error {
    p := make([]byte, 12)
    binary.LittleEndian.PutUint32(p, tag)
    binary.LittleEndian.PutUint64(p[4:], uint64(db.cursor))
    if _, err := writeAt(db.storage, endCommitTagPos, p); err != nil {
        return err
    }
    return db.storage.Sync()
}