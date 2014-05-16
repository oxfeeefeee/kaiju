// This project is offically started today on April 2nd, 2014. BTC price is now at 450USD. 
package kaiju

import (
    "time"
    "runtime"
    "math/rand"
    "github.com/oxfeeefeee/kaiju/config"
    "github.com/oxfeeefeee/kaiju/log"
    "github.com/oxfeeefeee/kaiju/profiling"
    "github.com/oxfeeefeee/kaiju/kio"
    "github.com/oxfeeefeee/kaiju/brain"
)

func mainFunc() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    profiling.RunProfiler()

    rand.Seed(time.Now().UTC().UnixNano())

    err := config.ReadJsonConfigFile()
    if err != nil {
        log.MainLogger().Printf("Failed to ready config file: %s", err.Error())
        return;
    }

    log.MainLogger().Printf("starting kio...")
    <- kio.Start(3)
    log.MainLogger().Printf("kio initialized.")

    brain.Start()

    // Don't quit
    c := make(chan struct{})
    _ = <- c
}

func main() {
    mainFunc()
}