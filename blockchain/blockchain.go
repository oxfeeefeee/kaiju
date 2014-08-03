package blockchain

import (
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/blockchain/cold"
)

const (
    InvTypeError = 0
    InvTypeTx = 1
    InvTypeBlock = 2
)

type InvElement struct {
    InvType     uint32
    Hash        klib.Hash256
}

// Initialize blockchain subsystem
func Init() error {
    return cold.Init()
}

// Destroy blockchain subsystem
func Destroy() error {
    return cold.Destroy()
}

// Get an array of InvElement to make a "getdata" message
func GetInv(heights []int) []*InvElement {
    headers := cold.TheHeaders()
    inv := make([]*InvElement, 0)
    for _, h := range heights {
        header := headers.Get(h)
        ele := &InvElement{InvTypeBlock, *(header.Hash())}
        inv = append(inv, ele)
    }
    return inv
}