package klib

import (
    "fmt"
    "testing"
)

func AddChild(t *testing.T, tree *KTree, depth int, index int, v interface{}) {
    err := tree.AddChild(depth, index, v)
    if err != nil {
        t.Errorf("err: %s", err)
    }
    fmt.Printf("tree:\n%s", tree)
}

func TestKTree(t *testing.T) {
    tree := NewKTree(1)
    AddChild(t, tree, 0, 0, 10)
    AddChild(t, tree, 0, 0, 11)
    AddChild(t, tree, 0, 0, 12)
    AddChild(t, tree, 1, 1, 20)
    AddChild(t, tree, 1, 1, 21)
    AddChild(t, tree, 0, 0, 13)
    AddChild(t, tree, 1, 2, 200)
    AddChild(t, tree, 1, 1, 22)
    AddChild(t, tree, 1, 1, 23)
    AddChild(t, tree, 1, 2, 201)
    AddChild(t, tree, 2, 3, 30)
    AddChild(t, tree, 2, 2, 31)
    AddChild(t, tree, 2, 3, 32)
    AddChild(t, tree, 2, 4, 33)
    AddChild(t, tree, 2, 5, 34)
    //AddChild(t, tree, 2, 6, 35)
    
    tree.Remove(1, 1)
    fmt.Printf("tree:\n%s", tree)
}