package kdb

import (
    "sync"
    "math"
    "errors"
    "encoding/binary"
    //"github.com/oxfeeefeee/kaiju/log"
)

// For value that with length of ValLenUnit, we make a mark and do not record the length
const ValLenUnit = 20

// Internal key size
const InternalKeySize = 6

// The slot size is 10, in which 6 bytes is KeySize and 4 bytes is the data pointer.
const SlotSize = 10

// File format version number
const Version = 1

// The size of header in bytes
//3_"KDB" + 1_version + 4*4_stats + commitTag + 8_cursor
const HeaderSize = 3 + 1 + 4 * 4 + 4 + 8

// How many slot we read at one time
const SlotBatchReadSize = 64

type Storage interface {
    Seek(offset int64, whence int) (ret int64, err error)
    Read(b []byte) (n int, err error)
    Write(b []byte) (n int, err error)
    Sync() (err error)
}

// Bitcoin uses a 256 bit hash to reference a privious TX as an input, to make it compact,
// we only use (6_bytes - 3_bits_flags) = 45 bit as the "internal key".
// With k slots and n keys, the expected number of collisions is n−k +k(1− 1/k)^n.
//
// It's expected for KDB to store thousands of txs with internal key collisions due to the 
// 45bit key.
// The solution is storing full key when there is a internal key collision.
type KDB struct {
    // Main storage of a file or a chuck of memory 
    storage             Storage
    // Storage for write-ahead data
    was                 Storage
    // Write ahead data
    wa                  waData
    // Current length, or the position to write next.
    cursor              int64
    // Mutex for the whole DB
    mutex               sync.RWMutex
    // Mutex for the main storage
    smutex              sync.RWMutex
    // DB statistics
    *Stats
}

func New(capacity uint32, s Storage, was Storage) (*KDB, error) {
    stats := &Stats{
        capacity: capacity,
        deadSlots: 0,
        deadValues: 0,
        records: 0,
    }
    cursor := HeaderSize + int64(capacity) * 2 * SlotSize
    db := &KDB{ 
        storage: s,
        was: was,
        wa: waData{map[int64]keyData{}, []byte{},},
        cursor: cursor,
        Stats: stats,
    }
    // Write header and init slots
    if err := writeHeader(s, stats, 0, cursor); err != nil {
        return nil, err
    }
    if err := db.writeBlankSections(); err != nil {
        return nil, err
    }
    if err := db.storage.Sync(); err != nil {
        return nil, err
    }
    // To generate a valid write ahead file
    if err := db.commit(0); err != nil {
        return nil, err
    }
    return db, nil
}

func Load(s Storage, was Storage) (*KDB, error) {
    stats, tag, cursor, err := readHeader(s)
    if err != nil {
        return nil, err
    }
    db := &KDB{ 
        storage: s,
        was: was,
        wa: waData{map[int64]keyData{}, []byte{},},
        Stats: stats,
    }
    db.cursor = cursor
    wastats, watag, _, err := readHeader(was)
    if err != nil {
        return nil, err
    }
    if tag != watag { // Need to re-commit write ahead data
        db.loadWAData()
        db.Stats = wastats
        if err := db.commit(watag); err != nil {
            return nil, err
        }
    }
    return db, nil
}

// Add a record
func (db *KDB) Add(key []byte, value []byte) error {
    if len(value) > int(math.MaxInt16) {
        return errors.New("KDB:Add data too long!")
    }
    kdata := toInternal(key)
    db.mutex.Lock()
    defer db.mutex.Unlock()
    db.smutex.RLock()
    defer db.smutex.RUnlock()
    collision := false
    n, err := db.slotScan(kdata, nil, 
        func(val []byte, mv bool) error {
            collision = true // We hit an internal key collision 
            //log.Debugf("Internal key collision happened")
            var cd collisionData
            if mv {
                cd.fromBytes(val)
            } else {
                cd.firstVal = val
            }
            cd.add(key, value)
            value = cd.toBytes()
            return nil
        })
    if err != nil {
        return err
    }
    c := make([]byte, SlotSize, SlotSize)
    kdata.setFlags(len(value) == ValLenUnit)
    copy(c[:InternalKeySize], kdata[:])
    binary.LittleEndian.PutUint32(c[InternalKeySize:], db.dataLoc())
    db.writeKey(c, n)
    db.writeValue(value, collision)
    return nil
}

// Get record value with key "k"
func (db *KDB) Get(key []byte) ([]byte, error) {
    kdata := toInternal(key)
    db.mutex.RLock()
    defer db.mutex.RUnlock()
    db.smutex.RLock()
    defer db.smutex.RUnlock()
    var value []byte
    _, err := db.slotScan(kdata, nil,
        func(val []byte, mv bool) error {
            if mv {
                var cd collisionData
                cd.fromBytes(val)
                value = cd.get(key)
            } else {
                value = val
            }
            return nil
        })
    if err != nil {
        return nil, err
    }
    return value, nil
}

// Removing a record doesn't delete the value of the record in the data section
// For non-internal-key-collision cases,it only modifies the slot section in two possible ways
// - If we know the next slot is empty, we can safely mark the deleted as empty
// - If we don't know the next slot is empty or not, we can only mark the deleted as deleted
// For internal-key-collision cases, we in-place change the slotData 
func (db *KDB) Remove(key []byte) (bool, error) {
    kdata := toInternal(key)
    db.mutex.Lock()
    defer db.mutex.Unlock()
    db.smutex.RLock()
    defer db.smutex.RUnlock()
    found := false
    _, err := db.slotScan(kdata, 
        func(slotData keyData, slotNum int64, emptyFollow bool, val []byte, mv bool) error {
            found = true
            if mv {
                var cd collisionData
                cd.fromBytes(val)
                cd.remove(key)
                if cd.len() == 1 {
                    val = cd.firstVal
                    mv = false
                } else {
                    val = cd.toBytes()
                }
                binary.LittleEndian.PutUint32(slotData[InternalKeySize:], db.dataLoc())
                db.writeKey(slotData, slotNum)
                db.writeValue(val, mv)
            } else {
                if emptyFollow {
                    slotData.setEmpty()
                } else {
                    slotData.setDeleted()
                }
                db.writeKey(slotData, slotNum)
            }
            return nil
        }, nil)
    if err != nil {
        return false, err
    }
    return found, nil
}

// Returns if write-ahead data is full
func (db *KDB) WAValueLen() int {
    db.mutex.RLock()
    defer db.mutex.RUnlock()
    return len(db.wa.ValData)
}

// Save data in memory to permanent storage
func (db *KDB) Commit(tag uint32) error {
    db.mutex.Lock()
    defer db.mutex.Unlock()
    db.smutex.Lock()
    defer db.smutex.Unlock()
    return db.commit(tag)
}

func (db *KDB) Tag() (uint32, error) {
    db.mutex.RLock()
    defer db.mutex.RUnlock()
    db.smutex.RLock()
    defer db.smutex.RUnlock()
    return db.tag()
}