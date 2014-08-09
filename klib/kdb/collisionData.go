package kdb

import (
    "bytes"
    "errors"
    "github.com/oxfeeefeee/kaiju/klib"
)

// Bitcoin uses 256bit hash as the ID of a transactions, the IDs don't collide.
// But KDB trim the ID to 45 bit to save space, and the internal keys could collide,
// when they do, the value stored in KDB becomes multiple-values
// 
// collisionData is used to represent the multiple-values.
// we cannot find the key for the first value, but for values after the first one, 
// the full key(original key) is stored to distinguish between them.
type collisionData struct {
    firstVal []byte
    otherKV [][2][]byte
}

func (d *collisionData) get(key []byte) []byte {
    for _, kv := range d.otherKV {
        //logger().Debugf("kv %v", kv)
        result := bytes.Compare(key, kv[0])
        if result == 0 {
            return kv[1]
        }
    }
    return d.firstVal
}

func (d *collisionData) add(key []byte, val []byte) {
    d.otherKV = append(d.otherKV, [2][]byte{key, val})
}

func (d *collisionData) remove(key []byte) {
    for i, kv := range d.otherKV {
        result := bytes.Compare(key, kv[0])
        if result == 0 {
            d.otherKV = append(d.otherKV[:i], d.otherKV[i+1:]...)
            return
        }
    }
    // The value to remove is firstVal
    d.firstVal = d.otherKV[0][1]
    d.otherKV = d.otherKV[1:]
}

func (d *collisionData) len() int {
    return len(d.otherKV) + 1
}

func (d *collisionData) fromBytes(p []byte) error {
    reader := bytes.NewReader(p)
    var s [][]byte
    for reader.Len() > 0 {
        var val klib.VarString
        err := val.Deserialize(reader)
        if err != nil {
            return err
        }
        s = append(s, val)
    }
    if len(s) % 2 != 1 {
        errors.New("collisionData.fromBytes invalid data")
    }
    d.fromSlice(s)
    return nil
}

func (d *collisionData) toBytes() []byte {
    s := d.toSlice()
    var p bytes.Buffer
    for _, val := range s {
        p.Write(klib.VarString(val).Bytes())
    }
    return p.Bytes()
}

func (d *collisionData) toSlice() [][]byte {
    size := 1 + 2 * len(d.otherKV)
    s := make([][]byte, 0, size)
    s = append(s, d.firstVal)
    for _, kv := range d.otherKV {
        s = append(s, kv[:]...)
    }
    return s
}

func (d *collisionData) fromSlice(s [][]byte) {
    d.firstVal = s[0]
    var kvs [][2][]byte
    kvCount := (len(s) - 1) / 2
    for i := 0; i < kvCount; i++ {
        kv := [2][]byte{s[i*2+1], s[i*2+2]}
        kvs = append(kvs, kv)
    }
    d.otherKV = kvs
}

