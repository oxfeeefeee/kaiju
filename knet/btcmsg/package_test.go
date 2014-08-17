package btcmsg

import (
    "bytes"
    "testing"
    "encoding/binary"
    "encoding/hex"
    "crypto/sha256"
)

func TestVersionMsg(t *testing.T) {
    buf := new(bytes.Buffer)
    vmsg := NewLocalVersionMsg(NewPeerInfo())
    log.Debugf("Sending data: %+v", vmsg)
    err := WriteMsg(buf, vmsg)
    log.Debugf("Encoded data: %s", hex.EncodeToString(buf.Bytes()))
    if err != nil {
        t.Errorf("Encode error : %s", err.Error())
    }

    rmsg, rerr := ReadMsg(buf)
    log.Debugf("Decoded data: %+v", rmsg)
    //vermsg := vmsg.(*Message_version)

    if rerr != nil {
        t.Errorf("Decode error : %s", rerr.Error())
    }
}

func TestChecksum(t *testing.T) {
    data := []byte{1,2,3,4,5,6,7,8,9}
    a := sha256.Sum256(data)
    aslice := make([]byte, sha256.Size)
    copy(aslice, a[:])
    b := sha256.Sum256(aslice)
    c := getChecksumForPayload(data)
    copy(aslice, b[:])
    d := binary.LittleEndian.Uint32(b[0:4])

    log.Debugf("sha256sha256 %v, %v", c, d)
    if d != c {
        t.Errorf("Checksum unmatch")
    }
    
}