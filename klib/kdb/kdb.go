package kdb

import (
    "io"
    "os"
    "fmt"
    "sync"
    "encoding/binary"
)

// For value that with length of DefaultValueLen, we make a mark and do not record the length
const DefaultValueLen = 20

// Internal key size
const InternalKeySize = 6

// The slot size is 10, in which 6 bytes is KeySize and 4 bytes is the data pointer.
const SlotSize = 10

// File format version number
const Version = 1

// The size of header in bytes
const HeaderSize = 8

// How many slot we read at one time
const SlotBatchReadSize = 32

// Bitcoin uses a 256 bit hash to reference a privious TX as an input, to make it compact,
// we only use (6_bytes - 3_bits_flags) = 45 bit as the "internal key".
// With k slots and n keys, the expected number of collisions is n−k +k(1− 1/k)^n.
//
// It's expected for KDB to store thousands of txs with internal key collisions due to the 
// 45bit key.
// The solution is storing full key when there is a internal key collision.
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
    // Thread safety
    mutex               sync.RWMutex
}

func New(capacity uint32, rws io.ReadWriteSeeker) (*KDB, error) {
    s := new(Stats)
    db := &KDB{ 
        capacity: capacity,
        rws: rws,
        stats: s,
    }
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

// Add a record
func (db *KDB) AddRecord(key []byte, value []byte) error {
    kdata := toInternalKey(key)
    collision := false
    n, err := db.slotScan(kdata, nil, 
        func(val []byte, mv bool) error {
            collision = true // We hit an internal key collision 
            logger().Debugf("Internal key collision happened")
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
    kdata.setFlags(len(value) == DefaultValueLen)
    copy(c[:InternalKeySize], kdata[:])
    binary.LittleEndian.PutUint32(c[InternalKeySize:], db.dataLoc())
    _, err = db.write(db.slotsBeginPos() + n * SlotSize, c)
    if err != nil {
        return err
    }
    return db.writeValue(value, collision)
}

// Get record value with key "k"
func (db *KDB) GetRecord(key []byte) ([]byte, error) {
    kdata := toInternalKey(key)
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
// For internal-key-collision cases, we in-place change the slotData and value 
func (db *KDB) RemoveRecord(key []byte) (bool, error) {
    kdata := toInternalKey(key)
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
                ptr := binary.LittleEndian.Uint32(slotData[InternalKeySize:])
                // Overwrite the old value in-place
                return db.writeValueAt(ptr, val, mv)
            } else {
                //Update first byte of slot content
                if emptyFollow {
                    slotData.setEmpty()
                } else {
                    slotData.setDeleted()
                }
                _, err := db.write(db.slotsBeginPos() + slotNum * SlotSize, slotData[:1])
                return err
            }
        }, nil)
    if err != nil {
        return false, err
    }
    return found, nil
}

func (db *KDB) String() string {
    return fmt.Sprintf("KDB:[capacity:%v, cursor:%v, scanCount:%v, slotReadCount:%v]",
        db.capacity, db.cursor, db.stats.ScanCount(), db.stats.SlotReadCount())
}

func FromFile(f *os.File) (*KDB, error) {
    //fi, err := f.Stat()
    //if err != nil {
    //    return nil, nil
    //}  
    return nil, nil 
}