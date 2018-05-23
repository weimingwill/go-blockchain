package blockchain

import (
	"encoding/hex"
	"log"

	"github.com/boltdb/bolt"
)

const utxoSetBucket = "chainstate"

var utxoBucketName = []byte(utxoSetBucket)

// UTXOSet represents UTXO set
type UTXOSet struct {
	Blockchain *Blockchain
}

// Reindex rebuilds the UTXOSet
func (us UTXOSet) Reindex() {

	// Delete and create bucket content again

	err := us.Blockchain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(utxoBucketName)

		if b != nil {
			tx.DeleteBucket(utxoBucketName)
		}

		_, err := tx.CreateBucket(utxoBucketName)

		return err
	})

	if err != nil {
		log.Panic(err)
	}

	// Get unspent transactions from blockchain in format of map[string]Outputs,
	// where the key is the transaction id, value is transaction outputs
	utxo := us.Blockchain.FindUTXO()

	// Store the utxo set in db bucket, using put
	err = us.Blockchain.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(utxoBucketName)

		for txID, outs := range utxo {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}

// FindSpendableOutputs finds and returns unspent outputs to reference in inputs.
func (us UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	var unspentOutputs = make(map[string][]int)
	accumulated := 0

	// Get the unspent outputs from db by bucket
	err := us.Blockchain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(utxoBucketName)
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := DeserializeOutputs(v)

			for outID, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outID)
				}
			}
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return accumulated, unspentOutputs
}

// FindUTXO finds UTXO for a public key hash
func (us UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	var utxo []TXOutput

	err := us.Blockchain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(utxoBucketName)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					utxo = append(utxo, out)
				}
			}
		}

		return nil

	})

	if err != nil {
		log.Panic(err)
	}

	return utxo
}

// Update updates the UTXO set with transactions from the Block
// The Block is considered to be the tip of a blockchain
func (us UTXOSet) Update(block *Block) {

	err := us.Blockchain.DB.Update(func(tx *bolt.Tx) error {
		var err error
		b := tx.Bucket(utxoBucketName)

		for _, tx := range block.Transactions {
			// If it is not coin base transction, get outputs of every transaction in the block
			// Delete unspent outputs if they are spent and store new unspent outputs
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					updatedOuts := TXOutputs{}

					outBytes := b.Get(in.TxID)
					outs := DeserializeOutputs(outBytes)

					for outID, out := range outs.Outputs {
						if outID != in.Vout {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						err = b.Delete(in.TxID)
					} else {
						err = b.Put(in.TxID, updatedOuts.Serialize())
					}

				}
			}

			// If it is coin base transction, just store the unspent output
			updatedOuts := TXOutputs{}

			for _, out := range tx.Vout {
				updatedOuts.Outputs = append(updatedOuts.Outputs, out)
			}

			err = b.Put(tx.ID, updatedOuts.Serialize())
		}

		return err
	})

	if err != nil {
		log.Panic(err)
	}
}

// CountTransactions returns the number of transactions in the UTXO set
func (us UTXOSet) CountTransactions() int {
	count := 0
	err := us.Blockchain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(utxoBucketName)
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			count++
		}
		return nil
	})

	if err != nil {
		log.Panic(err)
	}
	return count
}
