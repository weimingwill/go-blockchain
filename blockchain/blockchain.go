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
	dbFile              = "blockchain.db"
	blocksBucket        = "blocks"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

var lastHashKey = []byte("l")

// Blockchain keeps sequence of blocks
type Blockchain struct {
	DB *bolt.DB

	tip []byte
}

// MineBlock mines a block with transactions
func (bc *Blockchain) MineBlock(transactions []*Transaction) {
	for _, tx := range transactions {
		if !bc.VerifyTransaction(tx) {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	var lastHash []byte
	err := bc.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get(lastHashKey)

		return nil
	})

	block := NewBlock(transactions, lastHash)

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
func (bc *Blockchain) FindUTXO(pubKeyHash []byte) []TXOutput {
	var utxos []TXOutput
	utx := bc.FindUnspentTransactions(pubKeyHash)
	for _, tx := range utx {
		for _, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) {
				utxos = append(utxos, out)
			}
		}
	}
	return utxos
}

// FindSpendableOutputs finds spendable outputs of an address.
// The amount it returns either less than input amount or just exceed it.
// It also retunrs the transaction ids that accumulate to that amount.
func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	var unspentOutputs = make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outID, out := range tx.Vout {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outID)

				if accumulated > amount {
					break Work
				}
			}
		}
	}
	return accumulated, unspentOutputs
}

func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// NewBlockchain creates and returns a blockchain
// Todo: why need address as input ??
func NewBlockchain(address string) *Blockchain {
	if dbExists() == false {
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
func CreateBlockchain(address string) *Blockchain {
	if dbExists() == false {
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
