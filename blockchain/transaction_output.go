package blockchain

import (
	"bytes"
)

// TXOutput deines output of transactions
type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

// LockWithKey signs the output
func (out *TXOutput) LockWithKey(address []byte) {
	out.PubKeyHash = GetPubKeyHash(address)
}

// IsLockedWithKey checks if the output can be used by the owner of the pubkey
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

// NewTXOutput initializes a new transaction output
func NewTXOutput(value int, address string) *TXOutput {
	out := &TXOutput{
		Value: value,
	}

	out.LockWithKey([]byte(address))
	return out
}
