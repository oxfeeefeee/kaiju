// package "node" manages bitcoin node behavior, it uses kio, kdb and blockchain to:
// - Download, process and save historical bitcoin block data
// - Interact with other nodes in the bitcoin network to work as a full node
package node 

import (
    "github.com/oxfeeefeee/kaiju"
)

func Start() {
    go func() {
        // First make sure our blockchain is up to date
        catchUp()
        // Then run node
        runNode()
    }()
}

// Handy function
func logger() *kaiju.Logger {
    return kaiju.NodeLogger
}
