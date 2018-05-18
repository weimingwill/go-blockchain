package blockchain

import (
	"github.com/boltdb/bolt"
)

// Iterator is for iterating though blockchain
type Iterator struct {
	currentHash []byte
	db          *bolt.DB
}

// Next returns the next block in blockchain
func (i *Iterator) Next() *Block {
	var block *Block

	i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})

	i.currentHash = block.PrevBlockHash

	return block
}
