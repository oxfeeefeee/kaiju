package kdb

import (
    "io"
    "encoding/gob"
    )

// Write-ahead data
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

func (db *KDB) commit(tag uint32) error {
    if err := db.saveWAData(tag); err != nil {
        return err
    }
    return db.commitWAData(tag)
}

func (db *KDB) saveWAData(tag uint32) error {
    if _, err := db.wafile.Seek(HeaderSize, 0); err != nil {
        return err
    }
    if err := db.wa.save(db.wafile); err != nil {
        return err
    }
    if err := db.wafile.Sync(); err != nil {
        return err
    }
    // Now write the header of kdb
    if _, err := db.wafile.Seek(0, 0); err != nil {
        return err
    }
    cursor := db.cursor + int64(len(db.wa.ValData))
    if err := writeHeader(db.wafile, db.Stats, tag, cursor); err != nil {
        return err
    }
    return db.wafile.Sync()
}

func (db *KDB) loadWAData() error {
    if _, err := db.wafile.Seek(HeaderSize, 0); err != nil {
        return err
    }
    return db.wa.load(db.wafile)
}

func (db *KDB) commitWAData(tag uint32) error {
    for n, k := range db.wa.Keys {
        if _, err := writeAt(db.file, db.slotsBeginPos() + n * SlotSize, k); err != nil {
            return err
        }
    }
    n, err := writeAt(db.file, db.cursor, db.wa.ValData)
    if err != nil {
        return err
    }
    if err := db.file.Sync(); err != nil {
        return err
    }
    db.cursor += n
    db.wa.clear()
    // Update header to the end of committing
    if _, err := db.file.Seek(0, 0); err != nil {
        return err
    }
    if err := writeHeader(db.file, db.Stats, tag, db.cursor); err != nil {
        return err
    }
    return db.file.Sync()
}
