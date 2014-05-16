// All the messages in the bitcoin network protocol
package btcmsg

import (
    "bytes"
    "time"
    "github.com/oxfeeefeee/kaiju/cst"
)

// Bitcoin protocol message: "version"
type Message_version struct {
    Version         uint32
    Services        uint64
    Timestamp       int64
    Addr_recv       *PeerInfo
    Addr_from       *PeerInfo
    Nonce           uint64
    User_agent      VarString
    Start_height    int32
    Relay           byte
}

func NewLocalVersionMsg(addrRecv *PeerInfo) *Message_version {
    addrFrom := NewPeerInfo()// We don't accept incoming connections
    return &Message_version{
        cst.ProtocolVersion,
        cst.NodeServices,
        time.Now().Unix(),
        addrRecv,
        addrFrom,
        cst.NounceInVersionMsg,
        VarString(cst.UserAgent),
        1,
        0,
    }
}

func NewVerionMsg() Message {
    return &Message_version{
        Addr_recv: new(PeerInfo),
        Addr_from: new(PeerInfo), 
    }
}

func (m *Message_version) Command() string {
    return "version"
}

func (m *Message_version) Encode() ([]byte, error) {
    buf := new(bytes.Buffer)
    var err error;

    err = writeData(buf, &m.Version, err)
    err = writeData(buf, &m.Services, err)
    err = writeData(buf, &m.Timestamp, err)
    err = writePeerInfo(buf, m.Addr_recv, false, err)
    err = writePeerInfo(buf, m.Addr_from, false, err)
    err = writeData(buf, &m.Nonce, err)
    err = writeVarString(buf, &m.User_agent, err)
    err = writeData(buf, &m.Start_height, err)
    err = writeData(buf, &m.Relay, err)

    if err != nil {
        return nil, err;
    }
    return buf.Bytes(), nil
}

func (m *Message_version) Decode(payload []byte) error {
    buf := bytes.NewBuffer(payload)
    var err error

    err = readData(buf, &m.Version, err)
    err = readData(buf, &m.Services, err)
    err = readData(buf, &m.Timestamp, err)
    err = readPeerInfo(buf, m.Addr_recv, false, err)
    err = readPeerInfo(buf, m.Addr_from, false, err)
    err = readData(buf, &m.Nonce, err)
    err = readVarString(buf, &m.User_agent, err)
    err = readData(buf, &m.Start_height, err)
    err = readData(buf, &m.Relay, err)
    
    // TODO: better error handling
    return err;
}