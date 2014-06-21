// Encoding/decoding bitcoin messages into/from data packages

package btcmsg

import (
    "io"
    "fmt"
    "errors"
    "bytes"
    "encoding/binary"
    "github.com/oxfeeefeee/kaiju"
    "github.com/oxfeeefeee/kaiju/klib"
)

// A map from message name to it's creator function
var messageRegistry map[string]func() Message = map[string]func() Message {
    "getblocks":    NewGetBlocksMsg,
    "getheaders":   NewGetHeadersMsg,
    "inv":          NewInvMsg,
    "getdata":      NewGetDataMsg,
    "tx":           NewTxMsg,
    "block":        NewBlockMsg,
    "headers":      NewHeadersMsg,
    "notfound":     NewNotFoundMsg,
    "version":      NewVerionMsg,
    "verack":       NewVerAckMsg,
    "getaddr":      NewGetAddrMsg,
    "addr":         NewAddrMsg,
    "alert":        NewAlertMsg,
    "ping":         NewPingMsg,
    "pong":         NewPongMsg,
}

// Write a btc message to a io.Writer
func WriteMsg(w io.Writer, content Message) error {
    msgType := content.Command()
    payload, err := content.Encode()
    if err != nil {
        return err
    }
    var header *PackageHeader
    header = newPackageHeader(msgType, payload)
    err = binary.Write(w, binary.LittleEndian, header)
    if err != nil {
        return err
    }
    err = binary.Write(w, binary.LittleEndian, payload)
    if err != nil {
        return err
    }
    return nil
}

// Read a btc message from a io.Reader
// Returns: type, message body and error
//
// "msg" and "err" could both be nil when we get an invlid message and want to ignore it
func ReadMsg(r io.Reader) (Message, error) {
    header := new(PackageHeader)
    err := binary.Read(r, binary.LittleEndian, header)
    if err != nil {
        return nil, err
    }
    if header.Length > kaiju.MaxMessagePayload {
        return nil, errors.New("Error reading BTC message data : package too big")
    }
    payload := make([]byte, header.Length)
    err = binary.Read(r, binary.LittleEndian, payload)
    if err != nil {
        return nil, err
    }
    if header.Checksum != getChecksumForPayload(payload) {
        return nil, errors.New("Error reading BTC message data : wrong checksum")
    }
    command := header.getCommand()
    msg, error := decodeMessageFromPayload(command, payload)
    return msg, error
}

// A fixed size header of btc protocol message
type PackageHeader struct {
    Magic uint32
    Command [12]byte
    Length uint32
    Checksum uint32
}

func newPackageHeader(msgType string, payload []byte) *PackageHeader {
    header := new(PackageHeader)
    header.Magic = kaiju.NetWorkMagicMain
    header.setCommand(msgType)
    header.Length = uint32(len(payload))
    header.Checksum = getChecksumForPayload(payload)
    return header
}

func (h *PackageHeader)setCommand(command string) {
    copy(h.Command[:len(command)], command)
}

func (h *PackageHeader)getCommand() string {
    c := h.Command[:]
    n := bytes.Index(c, []byte{0})
    return string(c[:n])
}

func decodeMessageFromPayload(command string, payload []byte) (Message, error) {
    createFunc, ok := messageRegistry[command]
    if ok {
        msg := createFunc()
        err := msg.Decode(payload)
        return msg, err
    }
    return nil, errors.New(fmt.Sprintf("Invalid command type %s", command))
}

// checksum is the first 4 bytes of sha256(payload)
func getChecksumForPayload(payload []byte) uint32 {
    hash := klib.Sha256Sha256(payload)
    return binary.LittleEndian.Uint32(hash[:4])
}

