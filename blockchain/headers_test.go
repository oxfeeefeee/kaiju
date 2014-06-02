package blockchain

import (
    "testing"
    "fmt"
    "github.com/oxfeeefeee/kaiju/config"
)

func TestFiles(t *testing.T) {
    err := config.ReadJsonConfigFile()
    if err != nil {
        t.Errorf(fmt.Sprintf("Failed to read config file: %s", err))
    }

    InitFiles()
    //FileHeaders().Write([]byte{0,1,2,3,4,5,6,78})
}

func TestGenesisHeader(t *testing.T) {
    h := chain()[0]
    s := h.Hash().String()
    logger().Debugf("genesis hash %s", s)
    if s != "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f" {
        t.Errorf("Invalid genesis hash")
    }

    logger().Debugf("Locator %s", GetLocator()[0]) 
}