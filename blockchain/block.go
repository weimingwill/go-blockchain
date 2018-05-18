package blockchain

import (
	"bytes"
	"encoding/gob"
	"time"
)

// Block represents a block in blockchain
type Block struct {
	PrevBlockHash []byte
	Hash          []byte
	Data          []byte
	Timestamp     int64
	Nonce         int
}

// Serialize serializes block data to bytes
func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	// Todo - handle error
	encoder.Encode(b)
	return result.Bytes()
}

// DeserializeBlock converts serialized block bytes to block
func DeserializeBlock(b []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(b))
	// Todo - handle error
	decoder.Decode(&block)

	return &block
}

// NewBlock creates and returns Block
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		Data:          []byte(data),
		Timestamp:     time.Now().Unix(),
		PrevBlockHash: prevBlockHash,
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash
	block.Nonce = nonce

	return block
}

// NewGenesisBlock creates and returns the genesis block
func NewGenesisBlock() *Block {
	return NewBlock("Genesis block", []byte{})
}
