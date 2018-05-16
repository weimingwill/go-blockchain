package blockchain

import (
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

// // SetHash calculates and sets block hash
// func (b *Block) SetHash() {
// 	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
// 	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
// 	hash := sha256.Sum256(headers)

// 	b.Hash = hash[:]
// }

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
