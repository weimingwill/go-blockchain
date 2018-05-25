package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/boltdb/bolt"
)

const (
	blocksBucket        = "blocks"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
	dbFile              = "blockchain_%s.db"
)

var (
	lastHashKey      = []byte("l")
	blocksBucketName = []byte(blocksBucket)
)

// Blockchain keeps sequence of blocks
type Blockchain struct {
	DB *bolt.DB

	tip []byte
}

// MineBlock mines a block with transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	for _, tx := range transactions {
		if !bc.VerifyTransaction(tx) {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get(lastHashKey)

		blockData := b.Get(lastHash)
		block := DeserializeBlock(blockData)

		lastHeight = block.Height

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash, lastHeight+1)

	err = bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err = b.Put(lastHashKey, newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash

		return err
	})

	if err != nil {
		log.Panic(err)
	}

	return newBlock
}

// Iterator initializes a new blockchain iterator
func (bc *Blockchain) Iterator() *Iterator {
	bci := &Iterator{bc.tip, bc.DB}

	return bci
}

// FindTransaction finds a transaction by its id in the blockchain
func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("Transaction is not found")
}

// SignTransaction signs a transaction with wallet private key
func (bc *Blockchain) SignTransaction(tx *Transaction, privateKey ecdsa.PrivateKey) {
	prevTxs := bc.getPreviousTransactions(tx)

	tx.Sign(privateKey, prevTxs)
}

// VerifyTransaction verifies whether transaction is valid
func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	// After adding reward for mining, we are creating new coinbase transaction everytime.
	// No need to verify coinbase transaction.
	if tx.IsCoinbase() {
		return true
	}

	prevTxs := bc.getPreviousTransactions(tx)

	return tx.Verify(prevTxs)
}

// FindUnspentTransactions returns all unspent transactions of the blockchain
func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTXs []Transaction
	var spendTXOs = make(map[string][]int)

	bci := bc.Iterator()
	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for txOutID, out := range tx.Vout {
				if spendTXOs[txID] != nil {
					for _, spendOutID := range spendTXOs[txID] {
						if spendOutID == txOutID {
							continue Outputs
						}
					}
				}

				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}

			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					if in.UseKey(pubKeyHash) {
						inTXID := hex.EncodeToString(in.TxID)
						spendTXOs[inTXID] = append(spendTXOs[inTXID], in.Vout)
					}
				}
			}

		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// FindUTXO returns unspent transaction outputs
func (bc *Blockchain) FindUTXO() map[string]TXOutputs {
	utxos := make(map[string]TXOutputs)
	var spendTXOs = make(map[string][]int)

	bci := bc.Iterator()
	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for txOutID, out := range tx.Vout {
				if spendTXOs[txID] != nil {
					for _, spendOutID := range spendTXOs[txID] {
						if spendOutID == txOutID {
							continue Outputs
						}
					}
				}

				outs := utxos[txID]
				outs.Outputs = append(outs.Outputs, out)
				utxos[txID] = outs
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					inTXID := hex.EncodeToString(in.TxID)
					spendTXOs[inTXID] = append(spendTXOs[inTXID], in.Vout)
				}
			}

		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return utxos
}

// GetBestHeight returns the height of blockchain
func (bc *Blockchain) GetBestHeight() int {
	var lastBlock Block

	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(blocksBucketName)

		lastHash := b.Get(lastHashKey)
		blockData := b.Get(lastHash)
		lastBlock = *DeserializeBlock(blockData)

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return lastBlock.Height
}

// GetBlockHashes returns block hashes
func (bc *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte
	bci := bc.Iterator()

	for {
		block := bci.Next()
		blocks = append(blocks, block.Hash)

		if len(block.PrevBlockHash) == 0 {
			break
		}

	}
	return blocks
}

// GetBlock finds a block by its hash and returns it
func (bc *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	// Get block from db
	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(blocksBucketName)

		blockData := b.Get(blockHash)

		if blockData == nil {
			return errors.New("Block is not found")
		}

		block = *DeserializeBlock(blockData)

		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

// AddBlock saves the block into the blockchain
func (bc *Blockchain) AddBlock(block *Block) {
	err := bc.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(blocksBucketName)
		blockInDb := b.Get(block.Hash)

		if blockInDb != nil {
			return nil
		}

		blockData := block.Serialize()
		err := b.Put(block.Hash, blockData)
		if err != nil {
			log.Panic(err)
		}

		lastHash := b.Get(lastHashKey)
		lastBlockData := b.Get(lastHash)
		lastBlock := DeserializeBlock(lastBlockData)

		if block.Height > lastBlock.Height {
			err = b.Put(lastHashKey, block.Hash)
			if err != nil {
				log.Panic(err)
			}
			bc.tip = block.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func dbExists(dbFile string) bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// NewBlockchain creates and returns a blockchain
func NewBlockchain(nodeID string) *Blockchain {
	// Use unique db for different ndoes
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})

	bc := Blockchain{
		DB:  db,
		tip: tip,
	}
	return &bc
}

// CreateBlockchain creates and returns a blockchain
func CreateBlockchain(address, nodeID string) *Blockchain {
	dbFile := fmt.Sprintf(dbFile, nodeID)
	if dbExists(dbFile) == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			coinbaseTX := NewCoinbaseTX(address, genesisCoinbaseData)
			genesis := NewGenesisBlock(coinbaseTX)
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

func (bc *Blockchain) getPreviousTransactions(tx *Transaction) map[string]Transaction {
	prevTxs := make(map[string]Transaction)

	for _, in := range tx.Vin {
		prevTx, err := bc.FindTransaction(in.TxID)
		if err != nil {
			log.Panic(err)
		}
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}
	return prevTxs
}
