package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

const subsidy = 10

// Transaction defines a transaction in blockchain
type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

// IsCoinbase returns whether current transaction is coinbase transaction
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].TxID) == 0 && tx.Vin[0].Vout == -1
}

// SetID sets id of the transction by encoding all its data
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	encoder := gob.NewEncoder(&encoded)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// TXInput defines input of transactions
type TXInput struct {
	TxID      []byte
	Vout      int // output index in transaction
	ScriptSig string
}

// CanUnlockOutputWith checks whether the input can unlocked the input data
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

// TXOutput deines output of transactions
type TXOutput struct {
	Value        int
	ScriptPubKey string
}

// CanBeUnlockedWith checks whether the output can be unlocked by input data
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

// NewCoinbaseTX initialzes a new transaction which is the first transaction of the blockchain.
// It gives incentivce for mining the this genesis transaction
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Rewad to %s", to)
	}

	txin := TXInput{[]byte{}, -1, data}
	txout := TXOutput{subsidy, to}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()

	return &tx
}

// NewUTXOTransaction initializes a new unspent transction
func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	accumulated, unspentOutputs := bc.FindSpendableOutputs(from, amount)
	if accumulated < amount {
		log.Panic("ERROR: Not enough funds")
	}

	for txIDEncoded, outs := range unspentOutputs {
		txid, err := hex.DecodeString(txIDEncoded)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXInput{
				TxID:      txid,
				Vout:      out,
				ScriptSig: from,
			}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, TXOutput{amount, to})
	if accumulated > amount {
		outputs = append(outputs, TXOutput{accumulated - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()
	return &tx
}
