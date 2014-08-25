package blockchain

import (
    "github.com/oxfeeefeee/kaiju/klib"
    "github.com/oxfeeefeee/kaiju/blockchain/storage"
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
    return storage.Get().Init()
}

// Destroy blockchain subsystem
func Destroy() error {
    return storage.Get().Destroy()
}

func GetInvElem(h int) *InvElement {
    headers := storage.Get().Headers()
    header := headers.Get(h)
    return &InvElement{InvTypeBlock, *(header.Hash())}
}