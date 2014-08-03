// ktree is a tree data structure used to store and maintain blockchain.
// blockchain is in fact a tree not a chain, because sometimes there can be two 
// or more competing chains
package klib

import (
    "fmt"
    "strconv"
    "bytes"
    "errors"
    )

// ktree uses a slice of slice to store the tree, with node in the same sub-slice
// if they are with same depth.
// And all node record the index of its parent.

type Node struct {
    // Parent index of the node
    parentIndex int16
    // Record how far to move when do the two passes deleting
    moveDist int16 
    // The value stored in the node
    Value interface{}
}

func (n *Node) String() string {
    return fmt.Sprintf("[%d->%v]", n.parentIndex, n.Value)
}

type nodeArray []*Node

type KTree []nodeArray

func NewKTree(rootVal interface{}) *KTree {
    return &KTree{nodeArray{&Node{0, 0, rootVal,},},}
}

func (t *KTree) String() string {
    var p bytes.Buffer
    for i, nodes := range *t {
        p.WriteString(strconv.Itoa(i))
        p.WriteByte('\t')
        for j, n := range nodes {
            p.WriteString(strconv.Itoa(j))
            p.WriteByte('-')
            p.WriteString(n.String())
            p.WriteByte('\t')
        }
        p.WriteByte('\n')
    }
    return p.String()
}

func (t *KTree) NodesByDepth(depth int) ([]*Node, error) {
    if depth < len(*t) {
        return (*t)[depth], nil
    } else {
        return nil, errors.New("KTree.NodesByDepth depth out of range.")
    }
}

func (t *KTree) Node(depth int, index int) (*Node, error) {
    if depth < len(*t) {
        nodes := (*t)[depth]
        if index < len(nodes) {
            return nodes[index], nil
        }
    }
    return nil, errors.New("KTree.Node depth out of range.")
}

// Add a child for the node at depth "depth" and with index "index"
func (t *KTree) AddChild(depth int, index int, value interface{}) error {
    if depth >= len(*t) || index >= len((*t)[depth]) {
        return errors.New("KTree.AddChild depth out of range.")
    }
    if depth == len(*t) - 1 {
        *t = append(*t, nodeArray{})
    }
    nodes := (*t)[depth+1]
    (*t)[depth+1] = append(nodes, &Node{int16(index), 0, value,})
    return nil
}

// Remove the node at depth "depth" and with index "index"
// and the sub-tree of the node
func (t *KTree) Remove(depth int, index int) error {
    if depth <= 0 || depth >= len(*t) || index >= len((*t)[depth]) {
        return errors.New("KTree.AddChild depth out of range.")
    }
    // First pass, mark all nodes that's to be deleted
    // and calculate move distance
    // 1.1 process the first depth level
    firstLevel := (*t)[depth]
    firstLevel[index].moveDist = -1
    for i := index + 1; i < len(firstLevel); i++ {
        firstLevel[i].moveDist = 1
    }
    // 1.2 the rest of the sub-tree
    deleted := map[int]bool{index:true,}
    for i := depth + 1; i < len(*t); i++ {
        newDeleted := make(map[int]bool)
        dist := 0
        for j, node := range (*t)[i] {
            if deleted[int(node.parentIndex)] {
                dist++
                newDeleted[j] = true
                node.moveDist = -1
            } else {
                node.moveDist = int16(dist)
                parentMove := (*t)[i-1][node.parentIndex].moveDist
                // update parentIndex to what-would-be after deletion
                node.parentIndex -= parentMove
            }
        }
        deleted = newDeleted
        if len(deleted) == 0 {
            break
        }
    }
    // Second pass, delete them
    for i := depth; i < len(*t); i++ {
        finalDist := 0
        nodes := (*t)[i]
        for j, node := range (*t)[i] {
            if node.moveDist > 0 {
                nodes[j-int(node.moveDist)] = node
                finalDist = int(node.moveDist)
            }
        }
        if finalDist == 0 {
            // Didn't move any node, we can stop now
            break
        } else {
            // Remove the space left by moved nodes
            (*t)[i] = nodes[:len(nodes)-finalDist]
        }
    }
    return nil
}

