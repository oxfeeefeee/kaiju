// Sliding window download congestion avoidance, very rudimentary.
package catchUp 

import (
    //"github.com/oxfeeefeee/kaiju/log"
    )

// max consective nops can be sent
const maxConsectiveNops = 20

// System is considered in congestion if health is below congThreshold
const congThreshold = 0.80

// See swdlca.update
const healthConst = 0.98

type swdlca struct {
    // The health of downloading stream
    health  float32
    // NOP count have been sent to ease congestion
    nops    int
}

func newSwdlca() *swdlca {
    return &swdlca{
        health: 1.0,
    }
}

// Accepts latest download success rate, returns if a nop should send
func (ca *swdlca) update(got, total int) bool {
    ca.health = ca.health * healthConst + (float32(got)/float32(total)) * (1.0 - healthConst)
    if ca.health < congThreshold && ca.nops < maxConsectiveNops {
        ca.nops++
        return true
    } else {
        ca.nops = 0
        return false
    }
}