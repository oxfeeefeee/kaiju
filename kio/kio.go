// KIO stands for Kaiju IO. 
//
// KIO is responsible for interacting with other bitcoin nodes,
// send and receive bitcoin protocol messages.
// KIO can be considered as an communication interface for the "ledger" and "catchUp" module
// to communicate with the rest of the network
package kio

import (
    "github.com/oxfeeefeee/kaiju/config"
    "github.com/oxfeeefeee/kaiju/log"
    "github.com/oxfeeefeee/kaiju/cst"
    "github.com/oxfeeefeee/kaiju/kio/peer"
)

type KIO struct {
    pool    *peer.Pool
    cc      *peer.CC
    idmgr   *peer.IDManager
}

func New() *KIO {
    idm := peer.NewIDManager()
    p := peer.NewPool(idm)
    cc := peer.NewCC(idm)
    return &KIO{p, cc, idm}
}

func (kio *KIO) Go() {
    seeds := config.GetConfig().SeedPeers
    for _, ip := range seeds {
        kio.cc.AddSeedAddr(ip, cst.ListenPort)
    }

    kio.pool.Go()
    kio.cc.Go([]peer.Monitor{kio.pool, kio.cc})

    // Don't quit
    c := make(chan struct{})
    _ = <- c
}

// Handy function
func logger() *log.Logger {
    return log.KioLogger
}


