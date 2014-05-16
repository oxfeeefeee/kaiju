package blockchain

import (
    "testing"
)

func TestGenesisHeader(t *testing.T) {
    h := Chain()[0]
    s := h.Hash().String()
    logger().Debugf("genesis hash %s", s)
    if s != "000000000019d6689c085ae165831e934ff763ae46a2a6c172b3f1b60a8ce26f" {
        t.Errorf("Invalid genesis hash")
    }

    logger().Debugf("Locator %s", Chain().GetLocator()[0]) 
}