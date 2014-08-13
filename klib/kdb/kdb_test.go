package kdb

import (
    "bytes"
    //"testing"
    "encoding/binary"
    //"encoding/hex"
    //"crypto/sha256"
    "os"
    "testing"
    "fmt"
    "path/filepath"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju"
)
  
func cookUint32(key uint32, value uint32)([]byte, []byte) {
    kbuf, vbuf := make([]byte, 8, 8), make([]byte, 4, 4)
    binary.LittleEndian.PutUint32(kbuf[:], uint32(0))
    binary.LittleEndian.PutUint32(kbuf[4:], key)
    //hash := sha256.Sum256(kbuf[0:4])
    //copy(kbuf[:], hash[0:7])
    binary.LittleEndian.PutUint32(vbuf, value)  
    return kbuf, vbuf
}

func writeUint32(t *testing.T, db *KDB, key uint32, value uint32) {
    kbuf, vbuf := cookUint32(key, value)
    db.Add(kbuf, vbuf)
}

func removeUint32(t *testing.T, db *KDB, key uint32, value uint32) {
    kbuf, _ := cookUint32(key, value)
    db.Remove(kbuf)
}

func testUint32(t *testing.T, db *KDB, key uint32, value uint32) {
    kbuf, vbuf := cookUint32(key, value)
    v, getErr := db.Get(kbuf)
    if getErr != nil {
        t.Errorf("Failed to getRecord KDB: %s", getErr)
    }
    if bytes.Compare(vbuf, v) != 0 {
        t.Errorf("Did not get %v %v %v %v", key, value, vbuf, v)
    }
}

func testNotUint32(t *testing.T, db *KDB, key uint32, value uint32) {
    kbuf, vbuf := cookUint32(key, value)
    v, getErr := db.Get(kbuf)
    if getErr != nil {
        t.Errorf("Failed to getRecord KDB: %s", getErr)
    }
    if bytes.Compare(vbuf, v) == 0 {
        t.Errorf("Did get %v", v)
    }
}

func commit(t *testing.T, db *KDB, tag uint32) {
    err := db.Commit(tag)
    if err != nil {
        logger().Debugf("errerrerrerrerr, %v", err)
        t.Errorf("Failed to Commit: %s", err)
    }
}

func _TestMemoryKDB(t *testing.T) {
    err := kaiju.Init()
    if err != nil {
        t.Errorf("Failed to call kaiju.Init: %s", err)
    }
    buf := klib.NewMemFile(50 * 1024 * 1024)

    //capacity := uint32(1024 * 102)
    capacity := uint32(200000)
    db, dberr := New(capacity, buf)
    if dberr != nil {
        t.Errorf("Failed to create KDB: %s", dberr)
    }

    for i:=uint32(0); i < capacity; i++ {
        writeUint32(t, db, uint32(i), uint32(i))  
    }


    for i:=uint32(0); i < capacity; i++ {
        testUint32(t, db, uint32(i), uint32(i))  
    }

    fmt.Printf("db: %s\n", db)
}

func TestKDB(t *testing.T) {
    err := kaiju.Init()
    if err != nil {
        t.Errorf("Failed to call kaiju.Init: %s", err)
    }

    cfg := kaiju.GetConfig()
    path := filepath.Join(kaiju.GetConfigFileDir(), cfg.TempDataDir)
    os.MkdirAll(path, os.ModePerm)

    path = filepath.Join(path, "testdb.dat")
    exists, _ := fileExists(path)
    if exists {
        fmt.Printf("File already there: %s", path)
    }

    fmt.Printf("File Path: %s\n", path)
    f, openErr := os.Create(path)
    if openErr != nil {
        t.Errorf("Failed to create file: %s", openErr)
    }

    if true {
        capacity := uint32(200000)
        db, dberr := New(capacity, f)
        if dberr != nil {
            t.Errorf("Failed to create KDB: %s", dberr)
        }

        for i:=uint32(0); i < capacity; i++ {
            writeUint32(t, db, uint32(i), uint32(i))  
        }


        for i:=uint32(0); i < capacity; i++ {
            testUint32(t, db, uint32(i), uint32(i))  
        }
        return  
    }


    capacity := uint32(1000)
    db, dberr := New(capacity, f)
    if dberr != nil {
        t.Errorf("Failed to create KDB: %s", dberr)
    }

    for i:=uint32(0); i < capacity/2; i++ {
        writeUint32(t, db, uint32(i), uint32(i))  
    }

    commit(t, db, 1)

    for i:=uint32(0); i < capacity/2; i+=3 {
        removeUint32(t, db, uint32(i), uint32(i))  
    }

    commit(t, db, 2)

    for i:=uint32(0); i < capacity/2; i+=3 {
        writeUint32(t, db, uint32(i), uint32(i))  
    }

    commit(t, db, 3)

    for i:=capacity/2; i < capacity; i++ {
        writeUint32(t, db, uint32(i), uint32(i))  
    }

    commit(t, db, 4)

    for i:=uint32(0); i < capacity; i++ {
        testUint32(t, db, uint32(i), uint32(i))  
    }


    for i:=uint32(0); i < capacity/10000; i++ {
        removeUint32(t, db, uint32(i), uint32(i))  
    }

    for i:=uint32(0); i < capacity/10000; i++ {
        testNotUint32(t, db, uint32(i), uint32(i))  
    }

    for  i:=uint32(0); i < capacity/10000; i++ {
        writeUint32(t, db, uint32(i), uint32(i))  
    }

    for i:=uint32(0); i < capacity; i++ {
        testUint32(t, db, uint32(i), uint32(i))  
    }

    fmt.Printf("db: %s\n", db)

    db, dberr = Load(f)
    if dberr != nil {
        t.Errorf("Failed to load KDB: %s", dberr)
    } else {
        for i:=uint32(0); i < capacity; i++ {
            testUint32(t, db, uint32(i), uint32(i))  
        }
    }
    
    if closeErr := f.Close(); closeErr != nil {
        t.Errorf("Error closing file: %s", closeErr)
    }
}

func fileExists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return false, err
}