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
)

func mainCleanUp(){
    logger().Printf("Cleaning up...")
    err := node.Destroy()
    if err != nil {
        logger().Printf("Error destroying node: %s", err.Error())
    }
}

func mainFunc() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    profiling.RunProfiler()

    rand.Seed(time.Now().UTC().UnixNano())

    err := kaiju.ReadJsonConfigFile()
    if err != nil {
        logger().Printf("Failed to ready config file: %s", err.Error())
        return;
    }

    kaiju.MainLogger().Printf("Starting KIO...")
    <- kio.Start(10)
    kaiju.MainLogger().Printf("KIO initialized.")

    kaiju.MainLogger().Printf("Initializing kio...")
    err = node.Init()
    if err != nil {
        logger().Printf("Error initializing node: %s", err.Error())
    }
    kaiju.MainLogger().Printf("Starting kio...")
    node.Start()
    kaiju.MainLogger().Printf("node started.")

    // Don't quit
    c := make(chan struct{})
    _ = <- c
}

func main() {
    mainFunc()
}

// Handy function
func logger() *kaiju.Logger {
    return kaiju.MainLogger()
}