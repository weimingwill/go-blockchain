package blockchain

import (
	"github.com/boltdb/bolt"
)

const (
	dbFile       = "blockchain.db"
	blocksBucket = "blocks"
)

var lastHashKey = []byte("l")

// Blockchain keeps sequence of blocks
type Blockchain struct {
	DB *bolt.DB

	tip []byte
}

// AddBlock adds a new block to blockchain
func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte
	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get(lastHashKey)

		return nil
	})

	block := NewBlock(data, lastHash)

	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err = b.Put(lastHashKey, block.Hash)
		err = b.Put(block.Hash, block.Serialize())
		bc.tip = block.Hash

		return err
	})
}

// Iterator initializes a new blockchain iterator
func (bc *Blockchain) Iterator() *Iterator {
	bci := &Iterator{bc.tip, bc.DB}

	return bci
}

// NewBlockchain creates and returns a blockchain
func NewBlockchain() *Blockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			genesis := NewGenesisBlock()
			b, err = tx.CreateBucket([]byte(blocksBucket))
			err = b.Put(genesis.Hash, genesis.Serialize())
			err = b.Put(lastHashKey, genesis.Hash)
			tip = genesis.Hash
		} else {
			tip = b.Get(lastHashKey)
		}
		return err
	})

	bc := Blockchain{
		DB:  db,
		tip: tip,
	}
	return &bc
}
