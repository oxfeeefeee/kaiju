// package "brain" is the brain of the node, it uses kio, kdb and blockchain to:
// - Download, process and save historical bitcoin block data
// - Interact with other nodes in the bitcoin network to work as a full node
// BTW, still remember the huge Kaiju brain from the movie? :)
package brain 

import (
    "github.com/oxfeeefeee/kaiju/log"
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
func logger() *log.Logger {
    return log.BrainLogger
}
