package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
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

// TXOutputs collects TXOutput
type TXOutputs struct {
	Outputs []TXOutput
}

// Serialize serializes TXOutputs
func (outs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer

	encoder := gob.NewEncoder(&buff)
	err := encoder.Encode(outs)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// DeserializeOutputs deserializes TXOutputs
func DeserializeOutputs(data []byte) TXOutputs {
	var outs TXOutputs
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&outs)
	if err != nil {
		log.Panic(err)
	}
	return outs
}
