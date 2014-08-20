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
    return cold.Get().Init()
}

// Destroy blockchain subsystem
func Destroy() error {
    return cold.Get().Destroy()
}

func GetInvElem(h int) *InvElement {
    headers := cold.Get().Headers()
    header := headers.Get(h)
    return &InvElement{InvTypeBlock, *(header.Hash())}
}