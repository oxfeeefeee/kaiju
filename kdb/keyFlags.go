// The first three bits of the first byte of the key are used as flags
package kdb

// This slot is take either by key are deleted garbage key
const OccupiedFlagBitMask uint8 = 0x80

// This slot is cccupied by garbage, the garbage cannot be removed because it's not followed by empty slot
const GarbageFlagBitMask uint8 = 0x40

// The value of this record is not with standard length, so we need to make a mark
const NonDefaultLenFlagBitMask uint8 = 0x20


type keyData []byte

// Calculate mask, this is adding record so OccupiedMarkBitMask is always set
func (key keyData) setDefaultFlags(defaultValLen bool) {
    flags := uint8(OccupiedFlagBitMask)
    if !defaultValLen {
        flags |= NonDefaultLenFlagBitMask
    } 
    key[0] |= flags
}

// Returns the original key0 in case we need to restore it
func (key keyData) clearFlags() byte {
    key0 := key[0]
    key[0] = key[0] & (^(OccupiedFlagBitMask | GarbageFlagBitMask | NonDefaultLenFlagBitMask))
    return key0
}

func (key keyData) validKey() bool {
    return len(key) > 0 && len(key) <= 6 &&
        (key[0] & (OccupiedFlagBitMask | GarbageFlagBitMask | NonDefaultLenFlagBitMask)) == 0
}

func (key keyData) emptyOrDeleted() bool {
    return ((key[0] & OccupiedFlagBitMask) == 0) || ((key[0] & GarbageFlagBitMask) != 0)
}

func (key keyData) empty() bool {
    return (key[0] & OccupiedFlagBitMask) == 0
}

func (key keyData) setEmpty() {
    key[0] = key[0] & (^OccupiedFlagBitMask)
}

func (key keyData) deleted() bool {
    return (key[0] & GarbageFlagBitMask) != 0
}

func (key keyData) setDeleted() {
    key[0] =  key[0] | GarbageFlagBitMask
}

func (key keyData) defaultLenVaule() bool {
    return (key[0] & NonDefaultLenFlagBitMask) == 0
}
