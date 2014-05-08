package kdb

import (
    "io"
    "fmt"
    "errors"
    "bytes"
    "math"
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju/log"
)

// For value that with length of DefaultValueLen, we make a mark and do not record the length
const DefaultValueLen = 20

// The slot size is 10, in which 6 bytes is hash+metadata and 4 bytes is the data pointer.
const SlotSize = 10

// File format version number
const Version = 1

// The size of header in bytes
const HeaderSize = 8

const OccupiedMarkBitMask uint8 = 0x80

const NonDefaultValueLenBitMask uint8 = 0x40

// How many slot we read at one time
const SlotBatchReadSize = 32

type slotOp func(content []byte) bool

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

func (db *KDB) String() string {
    return fmt.Sprintf("KDB:[capacity:%v, cursor:%v, scanCount:%v, slotReadCount:%v]",
        db.capacity, db.cursor, db.stats.scanCount, db.stats.slotReadCount)
}

func (db *KDB) addRecord(key []byte, value []byte) error {
    if !validKey(key) {
        return errors.New("Invalid key format when add record to KDB")
    }

    l := len(value)
    defaultLen := l == DefaultValueLen
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

    // Calculate mask, we are adding record so OccupiedMarkBitMask is always set
    flags := uint8(OccupiedMarkBitMask)
    if !defaultLen {
        flags |= NonDefaultValueLenBitMask
    }
    // We use the first two bits of the first byte of the key as flags
    key[0] |= flags
    err = db.writeSlot(key, valueLoc)
    // Restore the content of key
    clearMaskBits(key) 
    return err
}

// KDB can store multiple record with the same key,
// if there are two record with key k, and you need to get the second one, use getRecord(k, 2)
func (db *KDB) getRecord(key []byte, whichOne int) ([]byte, error) {
    if !validKey(key)  {
        return nil, errors.New("Invalid key format when get record from KDB")
    } 
    c := make([]byte, SlotSize, SlotSize)

    f := func(content []byte) bool {
        // Exit if we hit an empty slot
        if content[0] & OccupiedMarkBitMask == 0 {
            return true
        }
        Key0, skey0 := key[0], content[0]
        clearMaskBits(key)
        clearMaskBits(content)
        result := bytes.Compare(key, content[:6])
        key[0], content[0] = Key0, skey0
        if result == 0 {
            copy(c[:], content[:])
            return true
        } 
        return false
    }
    _, err := db.slotScan(db.getDefaultSlot(key), f, whichOne)
    if err != nil {
        return nil, err
    }
    if c[0] & OccupiedMarkBitMask == 0 { // Did not find it
        return nil, nil
    }
    defaultLen := (c[0] & NonDefaultValueLenBitMask) == 0
    ptr := binary.LittleEndian.Uint32(c[6:])
    return db.readValue(ptr, defaultLen)
}

// Get the default slot number for a given key
func (db *KDB) getDefaultSlot(key []byte) int64 {
    Key0 := key[0]
    clearMaskBits(key)
    // Convert the byte slice to a uint64, it's a hash so we dont care about endianness
    // but let's make it look like big endian
    key64 := int64(key[0])
    for i := 1; i < 6; i += 1 {
        key64 = key64 << 8 + int64(key[i]) 
    }
    n := key64 % db.slotCount()
    key[0] = Key0
    return n
}

// Scan for a slot, a.k.a. linear probing. we either look for empty slot or a specific key
// slotOp:
//      - receives the stored slot content and returns ture if the content matches
//      - does more operations with the matched slot content if needed
// stopAt: how many times slotOp returns true before the scan stops
func (db *KDB) slotScan(slotNum int64, f slotOp, stopAt int) (int64, error) {
    db.rws.Seek(db.slotsBeginPos() + slotNum * SlotSize, 0)
    t := db.slotCount()
    i := slotNum
    if i >= t {
        panic("KDB.findEmptySlot: slot number >= slot count")
    }
    db.stats.scanCount++
    for {
        reachedEnd := false
        size := int64(SlotBatchReadSize)
        if i + SlotBatchReadSize >= t {
            size = t - i
            reachedEnd = true
        }
        
        buf := make([]byte, size * SlotSize, size * SlotSize)
        _, err := db.rws.Read(buf)
        if err != nil {
            return slotNum, err
        }
        db.stats.slotReadCount++
        for j := int64(0); j < size; j++ {
            if f(buf[j * SlotSize : (j + 1) * SlotSize]) {
                stopAt -= 1
                if stopAt == 0 {
                    return i + j, nil
                }
            }
        }
        
        if reachedEnd {
            i = 0
            db.rws.Seek(db.slotsBeginPos(), 0)
        } else {
            i += size    
        }
    }
    return -1, errors.New("Could find an empty slot")
}

func (db *KDB) writeSlot(key []byte, valueLoc uint32) error {
    f := func(content []byte) bool {
        return content[0] & OccupiedMarkBitMask == 0
    }
    n, err := db.slotScan(db.getDefaultSlot(key), f, 1)
    if err != nil {
        return err
    }
    db.rws.Seek(db.slotsBeginPos() + n * SlotSize, 0)
    c := make([]byte, SlotSize, SlotSize)
    copy(c[:6], key[:])
    binary.LittleEndian.PutUint32(c[6:], valueLoc)
    _, err = db.rws.Write(c)
    return err
}

func (db *KDB) readValue(ptr uint32, isDefaultLen bool) ([]byte, error) {
    db.rws.Seek(db.dataBeginPos() + int64(ptr) * DefaultValueLen, 0)
    if isDefaultLen {
        value := make([]byte, DefaultValueLen, DefaultValueLen)
        _, err := db.rws.Read(value)
        if err != nil {
            return nil, err
        }
        return value, nil
    } else {
        var valueLen uint16
        err := binary.Read(db.rws, binary.LittleEndian, &valueLen)
        if err != nil {
            return nil, err
        }
        value := make([]byte, valueLen, valueLen)
        _, err = db.rws.Read(value)
        if err != nil {
            return nil, err
        }
        return value, nil
    }
}

func (db *KDB) writeValue(value []byte) error {
    // Seek should cost nothing if it's already there right?
    db.rws.Seek(db.cursor, 0)
    if len(value) == DefaultValueLen {
        n, err := db.rws.Write(value)
        if err != nil {
            return err
        }
        db.cursor += int64(n)
        return nil
    } else {
        vl := len(value)
        if vl > int(math.MaxUint16) {
            panic("KDB:writeValue data too long!")
        }
        // The data length must be multiplies of DefaultValueLen
        // so we need to pad with 0 when needed
        dl := vl + 2
        count := dl / DefaultValueLen
        if dl % DefaultValueLen > 0 {
            count++
        }
        fullLen := count * DefaultValueLen
        buf := make([]byte, fullLen, fullLen)
        // First 2 bytes is for length
        binary.LittleEndian.PutUint16(buf, uint16(vl))
        copy(buf[2:dl], value[:])
        n, err := db.rws.Write(buf)
        if err != nil {
            return err
        }
        db.cursor += int64(n)
        return nil
    }
}

// Header = "KDB" + a_byte_of_version + 4_byte_of_capacity
func (db *KDB) writeHeader() error {
    db.rws.Seek(0, 0)
    buffer := []byte{'K', 'D', 'B', Version, 0, 0, 0, 0}
    binary.LittleEndian.PutUint32(buffer[4:], uint32(db.capacity))
    n, err := db.rws.Write(buffer)
    if err == nil {
        db.cursor += int64(n)
    }
    return err
}

func (db *KDB) writeBlankSlots() error {
    db.rws.Seek(db.slotsBeginPos(), 0)
    bs := 1024
    sCount := int(db.capacity) * 2
    sToGo := sCount
    zeros := make([]byte, bs * SlotSize, bs * SlotSize)
    oldCursor := db.cursor
    for sToGo > 0 {
        s := bs
        if sToGo < bs {
            s = sToGo
            zeros = make([]byte, s * SlotSize, s * SlotSize)
        }
        sToGo -= s
        n, err := db.rws.Write(zeros)
        if err != nil {
            return err
        } else {
            db.cursor += int64(n)
        }
    }
    db.slotSectionSize = db.cursor - oldCursor
    return nil
}

func (db *KDB) slotsBeginPos() int64 {
    return HeaderSize
}

func (db *KDB) dataBeginPos() int64 {
    return db.slotsBeginPos() + db.slotSectionSize
}

// total_slot_count = db.capacity * 2
func (db *KDB) slotCount() int64 {
    return int64(db.capacity) * 2
}

func clearMaskBits(key []byte) {
    key[0] = key[0] & (^(OccupiedMarkBitMask | NonDefaultValueLenBitMask))
}

func validKey(key []byte) bool {
    return len(key) > 0 && len(key) <= 6 && ((key[0] & (OccupiedMarkBitMask | NonDefaultValueLenBitMask)) == 0)
}

// Handy function
func logger() *log.Logger {
    return log.KDBLogger
}

