// This project is offically started today on April 2nd, 2014. BTC price is now at 450USD. 
package main

import (
    "time"
    "runtime"
    "math/rand"
    "github.com/oxfeeefeee/kaiju"
    "github.com/oxfeeefeee/kaiju/profiling"
    "github.com/oxfeeefeee/kaiju/kio"
    "github.com/oxfeeefeee/kaiju/node"
    "github.com/oxfeeefeee/kaiju/blockchain"
)

func mainCleanUp(){
    kaiju.MainLogger().Printf("Cleaning up...")
    err := blockchain.CloseFiles()
    if err != nil {
        kaiju.MainLogger().Printf("Failed to close files: %s", err.Error())
    }
}

func mainFunc() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    profiling.RunProfiler()

    rand.Seed(time.Now().UTC().UnixNano())

    err := kaiju.ReadJsonConfigFile()
    if err != nil {
        kaiju.MainLogger().Printf("Failed to ready config file: %s", err.Error())
        return;
    }

    err = blockchain.InitFiles()
    if err != nil {
        kaiju.MainLogger().Printf("Failed to init files: %s", err.Error())
        return;
    }

    kaiju.MainLogger().Printf("starting kio...")
    <- kio.Start(10)
    kaiju.MainLogger().Printf("kio initialized.")
    node.Start()
    kaiju.MainLogger().Printf("node started.")

    // Don't quit
    c := make(chan struct{})
    _ = <- c
}

func main() {
    mainFunc()
}