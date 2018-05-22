package blockchain

import (
	"bytes"
)

// TXInput defines input of transactions
type TXInput struct {
	TxID      []byte
	Vout      int // output index in transaction
	PubKey    []byte
	Signature []byte
}

// UseKey checks whether the address initiated the transaction
func (in *TXInput) UseKey(pubKeyHash []byte) bool {
	inPubHashKey := HashPubKey(in.PubKey)
	return bytes.Compare(inPubHashKey, pubKeyHash) == 0
}
