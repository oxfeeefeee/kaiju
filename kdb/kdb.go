package kdb

import (
    "io"
    "fmt"
    "errors"
)

// For value that with length of DefaultValueLen, we make a mark and do not record the length
const DefaultValueLen = 20

// The slot size is 10, in which 6 bytes is hash+flags and 4 bytes is the data pointer.
const SlotSize = 10

// File format version number
const Version = 1

// The size of header in bytes
const HeaderSize = 8

// How many slot we read at one time
const SlotBatchReadSize = 32

// Bitcoin uses a 256 bit hash to reference a privious TX as an input, to make it compact,
// we only use (6_bytes - 3_bits_flags) = 45 bit.
// With k slots and n keys, the expected number of collisions is n−k +k(1− 1/k)^n.
//
// It's expected for KDB to store thousands of txs with key collisions due to the 45bit key.
// The solution is straightforward: iterate through all the TXs with the same key to find the right one.
//
// ValuChecker is provided by the user to tell if the TxOutput is what's been looking for
type ValueChecker func(value []byte) bool

type KDB struct {
    // How many entries this DB is expected to store
    capacity            uint32
    // Size of the whole slot section in bytes
    slotSectionSize     int64
    // The ReadWriteSeeker of a file or a chuck of memory 
    rws                 io.ReadWriteSeeker
    // Current length, or the position to write next.
    cursor              int64
    // DB statistics
    stats               *Stats
}

type Stats struct {
    // How many slots are occupied, including slots that are marked as deleted
    occupiedSlotCount   int64
    // How many entries in this DB recordCount = occupiedSlotCount - slots_occupied_by_deleted_items
    recordCount         int64
    // How many scan operations did in total
    scanCount           int64
    // How many slot-read did
    slotReadCount       int64
}

func New(capacity uint32, rws io.ReadWriteSeeker) (*KDB, error) {
    s := new(Stats)
    db := &KDB{ capacity, 0, rws, 0, s}

    // Write header and init slots
    herr := db.writeHeader()
    if herr != nil {
        return nil, herr
    }
    serr := db.writeBlankSlots()
    if serr != nil {
        return nil, serr
    }
    return db, nil
}

// Add a record, KDB allows duplicated keys, i.e. one key for more than one value
func (db *KDB) AddRecord(key []byte, value []byte) error {
    if !keyData(key).validKey() {
        return errors.New("Invalid key format when add record to KDB")
    }
    // Calculate valueLoc: with which data will be found
    valuePtr := db.cursor - db.dataBeginPos()
    // DataLen unit is DefaultValueLen, so (valuePtr % DefaultValueLen) === 0
    if (valuePtr % DefaultValueLen) != 0 {
        panic("KDB.addRecord: (valuePtr % DefaultValueLen) != 0")
    } 
    valueLoc := uint32(valuePtr / DefaultValueLen)
    err := db.writeValue(value)
    if err != nil {
        return err
    }
    keyData(key).setDefaultFlags(len(value) == DefaultValueLen)
    err = db.writeSlot(key, valueLoc)
    // Restore the content of key
    keyData(key).clearFlags()
    return err
}

// Get record value with key "k"
func (db *KDB) GetRecord(key []byte, check ValueChecker) ([]byte, error) {
    if !keyData(key).validKey() {
        return nil, errors.New("Invalid key format when get record from KDB")
    } 
    var value []byte
    _, err := db.slotScan(db.getDefaultSlot(key), 
        findRecordFunc(key, db, check, 
            func(slotBuf []byte, offset int64, val []byte, slotNum int64) error {
                value = val
                return nil
            }))
    if err != nil {
        return nil, err
    }
    return value, nil
}

// Removing a record doesn't delete the value of the record in the data section
// It only modifies the slot section in two possible ways
// - If we know the next slot is empty, we can safely mark the deleted as empty
// - If we don't know the next slot is empty or not, we can only mark the deleted as deleted
func (db *KDB) RemoveRecord(key[]byte, check ValueChecker) (bool, error) {
    if !keyData(key).validKey() {
        return false, errors.New("Invalid key format when remove record from KDB")
    }
    found := false
    _, err := db.slotScan(db.getDefaultSlot(key), 
        findRecordFunc(key, db, check, 
            func(slotBuf []byte, offset int64, val []byte, slotNum int64) error {
                found = true
                next := offset + SlotSize
                if next < int64(len(slotBuf)) && keyData(slotBuf[next:]).empty() {
                    keyData(slotBuf[offset:]).setEmpty()
                } else {
                    keyData(slotBuf[offset:]).setDeleted()
                }
                db.rws.Seek(db.slotsBeginPos() + slotNum * SlotSize, 0)
                _, err := db.rws.Write(slotBuf[offset:offset+1])
                return err
            }))
    if err != nil {
        return false, err
    }
    return found, nil
}

func (db *KDB) String() string {
    return fmt.Sprintf("KDB:[capacity:%v, cursor:%v, scanCount:%v, slotReadCount:%v]",
        db.capacity, db.cursor, db.stats.scanCount, db.stats.slotReadCount)
}

