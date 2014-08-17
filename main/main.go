// This project is offically started today on April 2nd, 2014. BTC price is now at 450USD. 
package main

import (
    _ "github.com/oxfeeefeee/kaiju"
    _ "github.com/oxfeeefeee/kaiju/profiling"
    "github.com/oxfeeefeee/kaiju/log"
    "github.com/oxfeeefeee/kaiju/knet"
    "github.com/oxfeeefeee/kaiju/node"
)

func mainCleanUp(){
    log.Infof("Cleaning up...")
    err := node.Destroy()
    if err != nil {
        log.Infof("Error destroying node: %s", err.Error())
    }
}

func mainFunc() {
    c := make(chan struct{})
    /*defer func() {
        if r := recover(); r != nil {
            log.Infof("Main func paniced:", r)
            log.Infof("Exiting ...")
            close(c)
        }
    }()*/

    log.Infof("Starting KNet...")
    ch, err := knet.Start(10)
    if err != nil {
        log.Infof("Error starting knet: %s", err)
    }
    <- ch
    log.Infof("KNet initialized.")

    log.Infof("Initializing Node...")
    err = node.Init()
    if err != nil {
        log.Infof("Error initializing Node: %s", err.Error())
        return;
    }
    log.Infof("Starting Node...")
    node.Start()
    log.Infof("Node started.")

    // Don't quit
    _ = <- c
}

func main() {
    mainFunc()
}
