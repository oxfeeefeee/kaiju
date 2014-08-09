// I hate long source files
package kdb

import (
    "errors"
    "bytes"
    "math"
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju"
)

type handleValue func(value []byte, multiVal bool) error

type handleItem func(slotData keyData, slotNum int64, emptyFollow bool,
    value []byte, multiVal bool) error

// Scan for a slot, a.k.a. linear probing, stops either when get an empty slot
// or get a key match
func (db *KDB) slotScan(key []byte, hi handleItem, hv handleValue) (int64, error) {
    t := db.slotCount()
    slotNum := db.getDefaultSlot(key)
    i := slotNum
    if i >= t {
        panic("KDB.findEmptySlot: slot number >= slot count")
    }
    db.stats.incScanCount()
    for {
        reachedEnd := false
        size := int64(SlotBatchReadSize)
        if i + SlotBatchReadSize >= t {
            size = t - i
            reachedEnd = true
        }
        
        buf := make([]byte, size * SlotSize, size * SlotSize)
        db.stats.incSlotReadCount()
        _, err := db.read(db.slotsBeginPos() + i * SlotSize, buf)
        if err != nil {
            return slotNum, err
        }
        for j := int64(0); j < size; j++ {
            offset := j*SlotSize
            slotData := keyData(buf[offset:offset+SlotSize])
            if slotData.empty(){
                return i + j, nil
            } else if slotData.deleted() {
                continue
            }
            Key0, skey0 := keyData(key).clearFlags(), slotData.clearFlags()
            result := bytes.Compare(key, slotData[:InternalKeySize])
            key[0], slotData[0] = Key0, skey0
            if result == 0 {
                // A match is found, read the value
                defaultLen := keyData(slotData).defaultLenVaule()
                ptr := binary.LittleEndian.Uint32(slotData[InternalKeySize:])
                v, mv, err := db.readValue(ptr, defaultLen)
                if err != nil {
                    return -1, err
                } else {
                    if hi != nil {
                        emptyFollow := (j < size - 1) && keyData(buf[offset+SlotSize:]).empty()
                        err = hi(slotData, i + j, emptyFollow, v, mv)
                    } else if hv != nil {
                        err = hv(v, mv)  
                    }
                    if err != nil {
                        return -1, err
                    }
                }
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
    for i := 1; i < InternalKeySize; i += 1 {
        key64 = key64 << 8 + int64(key[i]) 
    }
    n := key64 % db.slotCount()
    key[0] = Key0
    return n
}

func (db *KDB) readValue(ptr uint32, isDefaultLen bool) ([]byte, bool, error) {
    pos := db.dataBeginPos() + int64(ptr) * DefaultValueLen
    if isDefaultLen {
        value := make([]byte, DefaultValueLen, DefaultValueLen)
        _, err := db.read(pos, value)
        if err != nil {
            return nil, false, err
        }
        return value, false, nil
    } else {
        db.mutex.RLock()
        defer db.mutex.RUnlock()
        _, err := db.rws.Seek(pos, 0)
        if err != nil {
            return nil, false, err
        }
        var valHeader int16
        err = binary.Read(db.rws, binary.LittleEndian, &valHeader)
        if err != nil {
            return nil, false, err
        }
        multiVal := valHeader < 0
        valueLen := valHeader
        if multiVal {
            valueLen = -valueLen
        }
        value := make([]byte, valueLen, valueLen)
        _, err = db.rws.Read(value)
        if err != nil {
            return nil, false, err
        }
        return value, multiVal, nil
    }
}

func (db *KDB) writeValue(value []byte, multiVal bool) error {
    if len(value) == DefaultValueLen && !multiVal {
        n, err := db.write(db.cursor, value)
        if err != nil {
            return err
        }
        db.cursor += int64(n)
        return nil
    } else {
        vl := len(value)
        if vl > int(math.MaxInt16) {
            return errors.New("KDB:writeValue data too long!")
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
        // First 2 bytes is for length and multiVal flag
        if multiVal {
            vl = - vl
        }
        binary.LittleEndian.PutUint16(buf, uint16(vl))
        copy(buf[2:dl], value[:])
        n, err := db.write(db.cursor, buf)
        if err != nil {
            return err
        }
        db.cursor += int64(n)
        return nil
    }
}

func (db *KDB) writeValueAt(ptr uint32, value []byte, multiVal bool) error {
    oldCursor := db.cursor
    db.cursor = db.dataBeginPos() + int64(ptr) * DefaultValueLen
    err := db.writeValue(value, multiVal)
    db.cursor = oldCursor
    return err
}

// Header = "KDB" + a_byte_of_version + 4_byte_of_capacity
func (db *KDB) writeHeader() error {
    buffer := []byte{'K', 'D', 'B', Version, 0, 0, 0, 0}
    binary.LittleEndian.PutUint32(buffer[4:], uint32(db.capacity))
    n, err := db.write(0, buffer)
    if err == nil {
        db.cursor += int64(n)
    }
    return err
}

// The size of slot chuck is pre-defined: SlotSize * 2 * Capacity
// Only way to change the capacity is to rebuild a new DB
func (db *KDB) writeBlankSlots() error {
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
        n, err := db.write(db.slotsBeginPos(), zeros)
        if err != nil {
            return err
        } else {
            db.cursor += int64(n)
        }
    }
    db.slotSectionSize = db.cursor - oldCursor
    return nil
}

func (db *KDB) read(c int64, p []byte) (int64, error) {
    db.mutex.RLock()
    defer db.mutex.RUnlock()
    _, err := db.rws.Seek(c, 0)
    if err != nil {
        return 0, err
    }
    n, err := db.rws.Read(p)
    return int64(n), err
}

func (db *KDB) write(c int64, p []byte) (int64, error) {
    db.mutex.RLock()
    _, err := db.rws.Seek(c, 0)
    db.mutex.RUnlock()
    if err != nil {
        return 0, err
    }
    db.mutex.Lock()
    defer db.mutex.Unlock()
    n, err := db.rws.Write(p)
    return int64(n), err
}

func (db *KDB) dataLoc() uint32 {
    // Calculate valueLoc: with which data will be found
    valuePtr := db.cursor - db.dataBeginPos()
    // DataLen unit is DefaultValueLen, so (valuePtr % DefaultValueLen) === 0
    if (valuePtr % DefaultValueLen) != 0 {
        panic("KDB.addRecord: (valuePtr % DefaultValueLen) != 0")
    } 
    return uint32(valuePtr / DefaultValueLen)
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