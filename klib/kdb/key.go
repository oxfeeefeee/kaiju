// The first three bits of the first byte of the key are used as flags
package kdb

import (
    "crypto/sha256"
    //"github.com/oxfeeefeee/kaiju/log"
)

// This slot is take either by key are deleted garbage key
const occupiedBit uint8 = 0x80

// This slot is cccupied by garbage, the garbage cannot be removed because it's not followed by empty slot
const garbageBit uint8 = 0x40

// The value of this record is not with standard length, so we need to make a mark
const nonUnitLenBit uint8 = 0x20

const maskBits uint8 = 
    occupiedBit     |
    garbageBit      | 
    nonUnitLenBit

type keyData []byte

// Calculate mask, this is adding record so OccupiedMarkBitMask is always set
func (key keyData) setFlags(defaultValLen bool) {
    key.clearFlags()
    flags := uint8(occupiedBit)
    if !defaultValLen {
        flags |= nonUnitLenBit
    } 
    key[0] |= flags
}

// Returns the original key0 in case we need to restore it
func (key keyData) clearFlags() byte {
    key0 := key[0]
    key[0] = key[0] & (^maskBits)
    return key0
}

func (key keyData) validKey() bool {
    return len(key) > 0 && len(key) <= 6 &&
        (key[0] & maskBits) == 0
}

func (key keyData) empty() bool {
    return (key[0] & occupiedBit) == 0
}

func (key keyData) setEmpty() {
    key[0] = key[0] & (^occupiedBit)
}

func (key keyData) deleted() bool {
    return (key[0] & garbageBit) != 0
}

func (key keyData) setDeleted() {
    key[0] =  key[0] | garbageBit
}

func (key keyData) unitValLen() bool {
    return (key[0] & nonUnitLenBit) == 0
}

// Get a shorter 45bit internal key from full key
func toInternal(fullKey []byte) keyData {
    key := sha256.Sum256(fullKey[:])
    //key := make([]byte, len(fullKey))
    //copy(key, fullKey[:5])

    key[0] = key[0] & (^maskBits)
    return key[:InternalKeySize]
}