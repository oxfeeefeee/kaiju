package kdb

import (
    "bytes"
    "testing"
    "fmt"
)

func writeCDUint32(t *testing.T, cd *collisionData, key uint32, value uint32) {
    kbuf, vbuf := cookUint32(key, value)
    cd.add(kbuf, vbuf)
}

func removeCDUint32(t *testing.T, cd *collisionData, key uint32, value uint32) {
    kbuf, _ := cookUint32(key, value)
    cd.remove(kbuf)
}

func verifyCDUint32(t *testing.T, cd *collisionData, key uint32, value uint32) {
    kbuf, vbuf := cookUint32(key, value)
    v := cd.get(kbuf)
    if bytes.Compare(vbuf, v) != 0 {
        t.Errorf(fmt.Sprintf("Didn't get what we set: %d-%d", key, value))
    }
}
  
func TestCollisionData(t *testing.T) {
    cd := new(collisionData)
    writeCDUint32(t, cd, 1, 11)
    writeCDUint32(t, cd, 2, 12)
    writeCDUint32(t, cd, 3, 13)
    writeCDUint32(t, cd, 4, 14)

    removeCDUint32(t, cd, 3, 13)
    removeCDUint32(t, cd, 4, 14)
    removeCDUint32(t, cd, 1, 11)
    writeCDUint32(t, cd, 1, 11)
    writeCDUint32(t, cd, 3, 13)
    writeCDUint32(t, cd, 4, 14)

    p := cd.toBytes()
    cd2 := new(collisionData)
    cd2.fromBytes(p)
    cd = cd2

    verifyCDUint32(t, cd, 1,11)
    verifyCDUint32(t, cd, 2,12)
    verifyCDUint32(t, cd, 3,13)
    verifyCDUint32(t, cd, 4,14)
}