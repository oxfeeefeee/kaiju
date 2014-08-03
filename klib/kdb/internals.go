// I hate long source files
package kdb

import (
    "errors"
    "bytes"
    "math"
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju"
)

// slotScanFunc's are called by KDB.slotScan, to operate on slot data
type slotScanFunc func(buf []byte, offset int64, slotNum int64) (bool, error)

// onRecordFoundFunc's are called by slotScanFunc when a match is found
type onRecordFoundFunc func(slotBuf []byte, offset int64, value []byte, slotNum int64) error

// Find an empty slot to add a record
func findEmptySlotFunc(buf []byte, offset int64, slotNum int64) (bool, error) {
    return keyData(buf[offset:]).emptyOrDeleted(), nil
}

func findRecordFunc(key []byte, db *KDB, check ValueChecker, f onRecordFoundFunc) slotScanFunc {
    return func(buf []byte, offset int64, slotNum int64) (bool, error) {
        // Exit if we hit an empty slot, deleted key doesn't stop the search
        if keyData(buf[offset:]).empty(){
            return true, nil
        } else if keyData(buf[offset:]).deleted() {
            return false, nil // Keep looking
        }
        slotData := buf[offset : offset + SlotSize]
        Key0, skey0 := keyData(key).clearFlags(), keyData(slotData).clearFlags()
        result := bytes.Compare(key, slotData[:6])
        key[0], slotData[0] = Key0, skey0
        if result == 0 {
            // A match is found, read the value and do the check
            defaultLen := keyData(slotData).defaultLenVaule()
            ptr := binary.LittleEndian.Uint32(slotData[6:])
            v, err := db.readValue(ptr, defaultLen)
            if err != nil {
                return false, err
            } else if check(v) {
                err := f(buf, offset, v, slotNum)
                return true, err
            }
        } 
        return false, nil
    }
}

// Scan for a slot, a.k.a. linear probing. we either look for empty slot or a specific key
// slotOp:
//      - receives the stored slot content and returns ture if the content matches
//      - does more operations with the matched slot content if needed
func (db *KDB) slotScan(slotNum int64, f slotScanFunc) (int64, error) {
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
        // "f" could change the current read position, seek is required
        db.stats.slotReadCount++
        _, err := db.rws.Seek(db.slotsBeginPos() + i * SlotSize, 0)
        if err != nil {
            return slotNum, err
        }
        _, err = db.rws.Read(buf)
        if err != nil {
            return slotNum, err
        }
        for j := int64(0); j < size; j++ {
            done, err := f(buf, j * SlotSize, i + j)
            if err != nil {
                return -1, err
            } else if done {
                return i + j, nil
            }
        }
        
        if reachedEnd {
            i = 0
        } else {
            i += size    
        }
    }
    return -1, errors.New("Could find an empty slot")
}

// Get the default slot number for a given key
func (db *KDB) getDefaultSlot(key []byte) int64 {
    Key0 := keyData(key).clearFlags()
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

func (db *KDB) writeSlot(key []byte, valueLoc uint32) error {
    n, err := db.slotScan(db.getDefaultSlot(key), findEmptySlotFunc)
    if err != nil {
        return err
    }
    _, err = db.rws.Seek(db.slotsBeginPos() + n * SlotSize, 0)
    if err != nil {
        return err
    }
    c := make([]byte, SlotSize, SlotSize)
    copy(c[:6], key[:])
    binary.LittleEndian.PutUint32(c[6:], valueLoc)
    _, err = db.rws.Write(c)
    return err
}

func (db *KDB) readValue(ptr uint32, isDefaultLen bool) ([]byte, error) {
    _, err := db.rws.Seek(db.dataBeginPos() + int64(ptr) * DefaultValueLen, 0)
    if err != nil {
        return nil, err
    }
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
    _, err := db.rws.Seek(db.cursor, 0)
    if err != nil {
        return err
    }
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
    _, err := db.rws.Seek(0, 0)
    if err != nil {
        return err
    }
    buffer := []byte{'K', 'D', 'B', Version, 0, 0, 0, 0}
    binary.LittleEndian.PutUint32(buffer[4:], uint32(db.capacity))
    n, err := db.rws.Write(buffer)
    if err == nil {
        db.cursor += int64(n)
    }
    return err
}

// The size of slot chuck is pre-defined: SlotSize * 2 * Capacity
// Only way to change the capacity is to rebuild a new DB
func (db *KDB) writeBlankSlots() error {
    _, err := db.rws.Seek(db.slotsBeginPos(), 0)
    if err != nil {
        return err
    }
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

// Handy function
func logger() *kaiju.Logger {
    return kaiju.MainLogger()
}