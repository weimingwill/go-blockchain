package blockchain

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

// Block represents a block in blockchain
type Block struct {
	PrevBlockHash []byte
	Hash          []byte
	Data          []byte
	Timestamp     int64
}

// SetHash calculates and sets block hash
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)

	b.Hash = hash[:]
}

// NewBlock creates and returns Block
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		Data:      []byte(data),
		Timestamp: time.Now().Unix(),
	}
	block.SetHash()

	return block
}

// NewGenesisBlock creates and returns the genesis block
func NewGenesisBlock() *Block {
	return NewBlock("Genesis block", []byte{})
}
