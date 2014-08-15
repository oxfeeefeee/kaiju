// I hate long source files
package kdb

import (
    "io"
    "errors"
    "bytes"
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
        _, err := readAt(db.storage, db.slotsBeginPos() + i * SlotSize, buf)
        if err != nil {
            return 0, err
        }
        for j := int64(0); j < size; j++ {
            offset := j*SlotSize
            // Get from write-ahead data first
            slotData, ok := db.wa.getKey(i+j)
            if !ok {
                slotData = keyData(buf[offset:offset+SlotSize])
            }
            if slotData.empty(){
                return i + j, nil
            } else if slotData.deleted() {
                continue
            }
            key0, skey0 := keyData(key).clearFlags(), slotData.clearFlags()
            result := bytes.Compare(key, slotData[:InternalKeySize])
            key[0], slotData[0] = key0, skey0
            if result == 0 {
                // A match is found, read the value
                defaultLen := keyData(slotData).defaultLenVaule()
                ptr := binary.LittleEndian.Uint32(slotData[InternalKeySize:])
                v, mv, err := db.readValue(ptr, defaultLen)
                if err != nil {
                    return -1, err
                } else {
                    if hi != nil {
                        ef := db.emptyFollow(buf[offset+SlotSize:], i + j) 
                        err = hi(slotData, i + j, ef, v, mv)
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

func (db *KDB) emptyFollow(buf []byte, slotNum int64) bool {
    if len(buf) > 0 {
        if d, ok := db.wa.getKey(slotNum + 1); ok {
            buf = d
        }
        return keyData(buf).empty()
    }
    return false
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
    var r io.ReadSeeker
    r = db.storage
    if pos >= db.cursor { // Need to read from Write-ahead-data
        pos -= db.cursor
        r = bytes.NewReader(db.wa.ValData)
    }
    if isDefaultLen {
        value := make([]byte, DefaultValueLen, DefaultValueLen)
        _, err := readAt(r, pos, value)
        if err != nil {
            return nil, false, err
        }
        return value, false, nil
    } else {
        _, err := r.Seek(pos, 0)
        if err != nil {
            return nil, false, err
        }
        hbuf := make([]byte, 2)
        err = binary.Read(r, binary.LittleEndian, hbuf)
        if err != nil {
            return nil, false, err
        }
        uintVal := binary.LittleEndian.Uint16(hbuf)
        valHeader := int16(uintVal)
        multiVal := valHeader < 0
        valueLen := valHeader
        if multiVal {
            valueLen = -valueLen
        }
        value := make([]byte, valueLen, valueLen)
        _, err = r.Read(value)
        if err != nil {
            return nil, false, err
        }
        return value, multiVal, nil
    }
}

func (db *KDB) writeValue(value []byte, multiVal bool) {
    if len(value) == DefaultValueLen && !multiVal {
        db.wa.addValue(value)
    } else {
        vl := len(value)
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
        db.wa.addValue(buf)
    }
}

func (db *KDB) writeKey(key keyData, slotNum int64) {
    db.wa.addKey(key, slotNum)
}

func (db *KDB) writeHeader() error {
    p := []byte{'K', 'D', 'B', Version, 0, 0, 0, 0}
    binary.LittleEndian.PutUint32(p[4:], uint32(db.capacity))
    p2 := make([]byte, HeaderSize)
    copy(p2, p)
    cursor := HeaderSize + WADataSize + int64(db.capacity) * 2 * SlotSize
    cp := savedCursorPos
    binary.LittleEndian.PutUint64(p2[cp:cp+8], uint64(cursor))
    n, err := writeAt(db.storage, 0, p2)
    if err == nil {
        db.cursor += int64(n)
    }
    return err
}

// Three sections to be written here:
// - Write ahead data (size = WADataSize)
// - Slot section (size = SlotSize * 2 * Capacity)
func (db *KDB) writeBlankSections() error {
    s := int64(db.capacity) * 2 * SlotSize
    s += int64(WADataSize)
    return db.writeBlank(1024 * 8, s)
}

func (db *KDB) writeBlank(batchSize int64, totalSize int64) error {
    left := totalSize
    cur := int64(0)
    for left > 0 {
        bs := batchSize
        if left < bs {
            bs = left
        }
        zeros := make([]byte, bs)
        _, err := writeAt(db.storage, db.slotsBeginPos() + cur, zeros)
        if err != nil {
            return err
        }
        left -= bs
        cur += bs
    }
    db.cursor += totalSize
    return nil
}

func (db *KDB) dataLoc() uint32 {
    // Calculate valueLoc: with which data will be found
    valuePtr := db.cursor - db.dataBeginPos() + int64(len(db.wa.ValData))
    // DataLen unit is DefaultValueLen, so (valuePtr % DefaultValueLen) === 0
    if (valuePtr % DefaultValueLen) != 0 {
        panic("KDB.addRecord: (valuePtr % DefaultValueLen) != 0")
    } 
    return uint32(valuePtr / DefaultValueLen)
}

func (db *KDB) slotsBeginPos() int64 {
    return HeaderSize + WADataSize
}

func (db *KDB) dataBeginPos() int64 {
    return db.slotsBeginPos() + int64(db.capacity) * 2 * SlotSize
}

// total_slot_count = db.capacity * 2
func (db *KDB) slotCount() int64 {
    return int64(db.capacity) * 2
}

func readAt(r io.ReadSeeker, c int64, p []byte) (int64, error) {
    _, err := r.Seek(c, 0)
    if err != nil {
        return 0, err
    }
    n, err := r.Read(p)
    if err != nil{
    }
    return int64(n), err
}

func writeAt(w io.WriteSeeker, c int64, p []byte) (int64, error) {
    _, err := w.Seek(c, 0)
    if err != nil {
        return 0, err
    }
    n, err := w.Write(p)
    return int64(n), err
}

// Returns capacity, beginCommitTag, endCommitTag, cursor
func readHeader(s Storage) (uint32, uint32, uint32, int64, error) {
    errInvalid := errors.New("Invalid KDB header")
    p := make([]byte, HeaderSize)
    if _, err := readAt(s, 0, p); err != nil {
        return 0, 0, 0, 0, err
    }
    if p[0] != 'K' || p[1] != 'D' || p[2] != 'B' {
        return 0, 0, 0, 0, errInvalid
    }
    if Version != p[3] {
        return 0, 0, 0, 0, errInvalid   
    }
    capacity := binary.LittleEndian.Uint32(p[4:8])
    if capacity <= 0 {
        return 0, 0, 0, 0, errInvalid
    }
    bc, ec, cu := beginCommitTagPos, endCommitTagPos, savedCursorPos
    bcTag := binary.LittleEndian.Uint32(p[bc:bc+4])
    ecTag := binary.LittleEndian.Uint32(p[ec:ec+4])
    cursor := int64(binary.LittleEndian.Uint64(p[cu:cu+8]))
    if cursor < HeaderSize + WADataSize + int64(capacity) * 2 * SlotSize {
        return 0, 0, 0, 0, errInvalid   
    }
    return capacity, bcTag, ecTag, cursor, nil
}

// Handy function
func logger() *kaiju.Logger {
    return kaiju.MainLogger()
}