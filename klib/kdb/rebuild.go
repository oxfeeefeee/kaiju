
package kdb

import (
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju/log"
)

type kvVisitor func(i uint32, sd []byte, val []byte, mv bool) error

// enumerate db excluding not committed write ahead data
func (db *KDB) enumerate(vis kvVisitor) (records uint32, garbage uint32, err error) {
    t := db.slotCount()
    i := int64(0)
    reachedEnd := false
    for !reachedEnd {
        size := int64(SlotBatchReadSize)
        if i + size >= t {
            size = t - i
            reachedEnd = true
        }
        buf := make([]byte, size * SlotSize, size * SlotSize)
        _, err = readAt(db.file, db.slotsBeginPos() + i * SlotSize, buf)
        if err != nil {
            return
        }
        for j := int64(0); j < size; j++ {
            offset := j*SlotSize
            slotData := keyData(buf[offset:offset+SlotSize])
            if slotData.deleted() {
                garbage++
                continue
            }
            if !slotData.empty(){
                records++
                if vis != nil {
                    defaultLen := keyData(slotData).unitValLen()
                    ptr := binary.LittleEndian.Uint32(slotData[InternalKeySize:])
                    var val []byte
                    var mv bool
                    val, mv, err = db.readValue(ptr, defaultLen)
                    if err != nil {
                        return 
                    }
                    err = vis(records, slotData, val, mv)
                    if err != nil {
                        return 
                    }
                }
            }
        }
        i += size
    }
    return
}

func (db *KDB) Rebuild(capacity uint32, file File, wafile File) (*KDB, error) {
    newdb, err := New(capacity, file, wafile)
    if err != nil {
        return nil, err 
    }
    f := func(i uint32, sd []byte, val []byte, mv bool) error {
        n, err := newdb.slotScan(sd[:InternalKeySize], nil, nil)
        if err != nil {
            return err
        }
        binary.LittleEndian.PutUint32(sd[InternalKeySize:], newdb.dataLoc())
        newdb.writeKey(sd, n)
        newdb.writeValue(val, keyData(sd).unitValLen(), mv)
        if i % 100000 == 0 {
            newdb.commit(i)
            log.Infof("KDB.Rebuild: current key count:%d", i)
        }
        return nil
    }
    db.smutex.RLock()
    defer db.smutex.RUnlock()
    if _, _, err = db.enumerate(f); err != nil {
        return nil, err
    }
    if tag, err := db.tag(); err != nil {
        return nil, err
    } else {
        newdb.commit(tag)
    }
    return newdb ,nil
}