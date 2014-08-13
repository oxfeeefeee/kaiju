// This project is offically started today on April 2nd, 2014. BTC price is now at 450USD. 
package main

import (
    "github.com/oxfeeefeee/kaiju"
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
    c := make(chan struct{})
    /*defer func() {
        if r := recover(); r != nil {
            logger().Printf("Main func paniced:", r)
            logger().Printf("Exiting ...")
            close(c)
        }
    }()*/

    err := kaiju.Init()
    if err != nil {
        return;
    }

    logger().Printf("Starting KIO...")
    <- kio.Start(10)
    logger().Printf("KIO initialized.")

    logger().Printf("Initializing Node...")
    err = node.Init()
    if err != nil {
        logger().Printf("Error initializing Node: %s", err.Error())
        return;
    }
    logger().Printf("Starting Node...")
    node.Start()
    logger().Printf("Node started.")

    // Don't quit
    _ = <- c
}

func main() {
    mainFunc()
}

// Handy function
func logger() *kaiju.Logger {
    return kaiju.MainLogger()
}