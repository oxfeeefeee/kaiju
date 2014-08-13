// The first three bits of the first byte of the key are used as flags
package kdb

import (
    "crypto/sha256"
)

// This slot is take either by key are deleted garbage key
const occupiedFlagBitMask uint8 = 0x80

// This slot is cccupied by garbage, the garbage cannot be removed because it's not followed by empty slot
const garbageFlagBitMask uint8 = 0x40

// The value of this record is not with standard length, so we need to make a mark
const nonDefaultLenFlagBitMask uint8 = 0x20

const maskBits uint8 = 
    occupiedFlagBitMask         |
    garbageFlagBitMask          | 
    nonDefaultLenFlagBitMask

type keyData []byte

// Calculate mask, this is adding record so OccupiedMarkBitMask is always set
func (key keyData) setFlags(defaultValLen bool) {
    flags := uint8(occupiedFlagBitMask)
    if !defaultValLen {
        flags |= nonDefaultLenFlagBitMask
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

func (key keyData) emptyOrDeleted() bool {
    return ((key[0] & occupiedFlagBitMask) == 0) || ((key[0] & garbageFlagBitMask) != 0)
}

func (key keyData) empty() bool {
    return (key[0] & occupiedFlagBitMask) == 0
}

func (key keyData) setEmpty() {
    key[0] = key[0] & (^occupiedFlagBitMask)
}

func (key keyData) deleted() bool {
    return (key[0] & garbageFlagBitMask) != 0
}

func (key keyData) setDeleted() {
    key[0] =  key[0] | garbageFlagBitMask
}

func (key keyData) defaultLenVaule() bool {
    return (key[0] & nonDefaultLenFlagBitMask) == 0
}

// Get a shorter 45bit internal key from full key
func toInternalKey(fullKey []byte) keyData {
    key := sha256.Sum256(fullKey[:])
    //key := make([]byte, len(fullKey))
    //copy(key, fullKey)
    //key[2], key[3] , key[4] , key[5] = 0,0,0, 0

    key[0] = key[0] & (^maskBits)
    return key[:InternalKeySize]
}