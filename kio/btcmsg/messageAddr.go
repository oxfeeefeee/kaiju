package btcmsg

import (
    "io"
    "fmt"
    "net"
    "time"
    "bytes"
    "errors"
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/catma/cst"
)

type PeerIP struct {
    //e.g. 00 00 00 00 00 00 00 00 00 00 FF FF 0A 00 00 01 - IPv6: ::ffff:a00:1 or IPv4: 10.0.0.1
    Data            [16]byte 
}

func (pip *PeerIP) ToNetIP() net.IP {
    nip := make([]byte, 16)
    copy(nip[:], pip.Data[:])
    return nip
}

func FromNetIP(nip *net.IP) PeerIP{
    var pip PeerIP
    copy(pip.Data[:], (*nip)[:])
    return pip
}

// The network address data structure used by bitcoin network
type PeerInfo struct {
    // Last time seem
    Time            uint32
    Services        uint64 
    IP              PeerIP
    Port            uint16
}

func NewPeerInfo() *PeerInfo {
    // TODO: cache it to reduce gc overhead
    return &PeerInfo{
        uint32(time.Now().Unix()),
        cst.NodeServices,
        PeerIP{},
        0,
    }
}

func (p *PeerInfo) ToTCPAddr() *net.TCPAddr{
    return &net.TCPAddr{
        p.IP.ToNetIP(),
        int(p.Port),
        "",
    }
}

func writePeerInfo(w io.Writer, p *PeerInfo, withTimestamp bool, lastError error) error {
    if withTimestamp {
        lastError = writeData(w, &p.Time, lastError) 
    }
    lastError = writeData(w, &p.Services, lastError)
    lastError = writeData(w, &p.IP, lastError)
    if lastError == nil {
        lastError = binary.Write(w, binary.BigEndian, &p.Port)
    }
    return lastError
}

func readPeerInfo(r io.Reader, p *PeerInfo, withTimestamp bool, lastError error) error {
    if withTimestamp {
        lastError = readData(r, &p.Time, lastError)  
    }
    lastError = readData(r, &p.Services, lastError)
    lastError = readData(r, &p.IP, lastError)
    if lastError == nil {
        lastError = binary.Read(r, binary.BigEndian, &p.Port)
    }
    return lastError
}

// Bitcoin protocol message: "addr"
type Message_addr struct {
    Addresses       []*PeerInfo
}

func NewAddrMsg() Message {
    return &Message_addr{
        make([]*PeerInfo, 0),
    }
}

func (m *Message_addr) Command() string {
    return "addr"
}

func (m *Message_addr) Encode() ([]byte, error) {
    buf := new(bytes.Buffer)
    var err error;
    listSize := klib.VarUint(len(m.Addresses))
    err = writeData(buf, &listSize, err)
    if err != nil {
        return nil, err
    }

    for _, addr := range m.Addresses {
        err = writePeerInfo(buf, addr, true, err)
    }
    if err != nil {
        return nil, err;
    }
    return buf.Bytes(), nil
}

func (m *Message_addr) Decode(payload []byte) error {
    buf := bytes.NewBuffer(payload)
    var err error;
    var listSize klib.VarUint;

    err = readData(buf, &listSize, err)
    if err != nil {
        return err
    }
    
    if listSize > klib.VarUint(cst.MaxAddrListSize) {
        return errors.New(fmt.Sprintf("Message_addr list too long: %v", listSize))
    }

    addresses := make([]*PeerInfo, listSize)
    for i := uint64(0); i < uint64(listSize); i++ {
        addresses[i] = new(PeerInfo)
        err = readPeerInfo(buf, addresses[i], true, err)
    } 
    m.Addresses = addresses
    return err
}