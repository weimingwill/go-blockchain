package blockchain

import (
	"crypto/sha256"
)

// MerkleTree represents a merkle tree
type MerkleTree struct {
	Root *MerkleNode
}

// MerkleNode represents a node in merkle tree
type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

// NewMerkleNode creates a new merkle tree node
func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	if left != nil || right != nil {
		data = append(left.Data, right.Data...)
	}

	hash := sha256.Sum256(data)
	node := MerkleNode{
		Left:  left,
		Right: right,
		Data:  hash[:],
	}
	return &node
}

// NewMerkleTree creates a new merkle tree
// Todo - it seems there are some algorithm problem for creating tree in this way. e.g. 10 nodes
func NewMerkleTree(data [][]byte) *MerkleTree {
	var nodes []MerkleNode

	// If number of data is not even, duplicate the last one to make it even
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}

	for i := 0; i < len(data)/2; i++ {
		var newLevel []MerkleNode

		for j := 0; j < len(nodes); j += 2 {
			node := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			newLevel = append(newLevel, *node)
		}

		nodes = newLevel
	}

	mTree := MerkleTree{&nodes[0]}

	return &mTree

}
