// This project is offically started today on April 2nd, 2014. BTC price is now at 450USD. 
package main

import (
    "runtime"
    "github.com/oxfeeefeee/kaiju/config"
    "github.com/oxfeeefeee/kaiju/log"
    "github.com/oxfeeefeee/kaiju/profiling"
    "github.com/oxfeeefeee/kaiju/kio"
)

func main() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    profiling.RunProfiler()

    err := config.ReadJsonConfigFile()
    if err != nil {
        log.MainLogger().Printf("Failed to ready config file: %s", err.Error())
        return;
    }

    kio := kio.New()
    kio.Go()
}
