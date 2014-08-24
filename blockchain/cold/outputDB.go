package cold

import (
    "fmt"
    "bytes"
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju"
    "github.com/oxfeeefeee/kaiju/log"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/klib/kdb"
    "github.com/oxfeeefeee/kaiju/catma"
    "github.com/oxfeeefeee/kaiju/catma/script"
)

// All unspent tx output stored in KDB
type outputDB struct {
    db  *kdb.KDB
}

func newOutputDB(db *kdb.KDB) *outputDB {
    return &outputDB{db,}
}

func (u *outputDB) Get(h *klib.Hash256, i uint32) (*catma.TxOut, error) {
    key := getKdbKey(h, i)
    if val, err := u.db.Get(key); err != nil {
        return nil, err
    } else if val == nil {
        return nil, fmt.Errorf("outputDB.Get Cannot find tx input %s %d", h, i)
    } else {
        return decodeTxo(val)
    }
}

func (u *outputDB) Use(h *klib.Hash256, i uint32, txo *catma.TxOut) error {
    key := getKdbKey(h, i)
    v, err := u.db.Get(key)
    if err != nil {
        return err
    }
    if txo != nil { // Verify txo if provided
        val, err := encodeTxo(txo)
        if err != nil {
            return err
        }
        result := bytes.Compare(v, val)
        if result != 0 {
            return fmt.Errorf("outputDB.Use value doesn't match value in DB %s %d", h, i)
        }
    }
    if found, err := u.db.Remove(key); err != nil {
        return err
    } else if !found {
        return fmt.Errorf("outputDB.Use Cannot find tx input %s %d", h, i)
    } else {
        return nil
    }
}

func (u *outputDB) Add(h *klib.Hash256, i uint32, txo *catma.TxOut) error {
    key := getKdbKey(h, i)
    if val, err := encodeTxo(txo); err != nil {
        return err
    } else {
        return u.db.Add(key, val)
    }
}

func (u *outputDB) Commit(tag uint32, force bool) error {
    if force || u.db.WAValueLen() > kaiju.GetConfig().MaxKdbWAValueLen {
        log.Infof("Committing blocks up to number %d ...", tag)
        err := u.db.Commit(tag)
        log.Infof("Committed blocks up to number %d", tag)
        return err
    }
    return nil
}

func (u *outputDB) Tag() (uint32, error) {
    return u.db.Tag()
}

func encodeTxo(txo *catma.TxOut) ([]byte, error) {
    s := script.Script(txo.PKScript)
    if s.IsTypePubKeyHash() {
        ret := make([]byte, 20+1+8)
        ret[0] = byte(script.PKS_PubKeyHash)
        binary.LittleEndian.PutUint64(ret[1:], uint64(txo.Value))
        copy(ret[9:], txo.PKScript[3:23])
        return ret, nil
    } else if s.IsTypeScriptHash() {
        ret := make([]byte, 20+1+8)
        ret[0] = byte(script.PKS_ScriptHash)
        binary.LittleEndian.PutUint64(ret[1:], uint64(txo.Value))
        copy(ret[9:], txo.PKScript[2:22])
        return ret, nil
    } else {
        ret := make([]byte, len(txo.PKScript)+1+8)
        ret[0] = 0
        binary.LittleEndian.PutUint64(ret[len(txo.PKScript)+1:], uint64(txo.Value))
        copy(ret[9:], txo.PKScript)
        return ret, nil
    }
}

func decodeTxo(val []byte) (*catma.TxOut, error) {
    fb := val[0]
    if fb == byte(script.PKS_PubKeyHash) {
        v := binary.LittleEndian.Uint64(val[1:])
        s := make([]byte, 25)
        s[0] = byte(script.OP_DUP)
        s[1] = byte(script.OP_HASH160)
        s[2] = byte(script.OP_PUSHDATA14)
        s[23] = byte(script.OP_EQUALVERIFY)
        s[24] = byte(script.OP_CHECKSIG)
        copy(s[3:23], val[9:21])
        return &catma.TxOut{int64(v), s}, nil
    } else if fb == byte(script.PKS_ScriptHash) {
        v := binary.LittleEndian.Uint64(val[1:])
        s := make([]byte, 23)
        s[0] = byte(script.OP_HASH160)
        s[1] = byte(script.OP_PUSHDATA14)
        s[22] = byte(script.OP_EQUAL)
        copy(s[2:22], val[9:])
        return &catma.TxOut{int64(v), s}, nil
    } else {
        v := binary.LittleEndian.Uint64(val[1:])
        // NOTE: reusing memory of val
        return &catma.TxOut{int64(v), val[9:]}, nil
    }
}

func getKdbKey(h *klib.Hash256, i uint32) []byte {
    p := klib.Uint32ToBytes(i)
    return append(p, h[:]...)
}

