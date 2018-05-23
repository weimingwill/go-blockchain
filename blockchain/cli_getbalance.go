package blockchain

import (
	"fmt"
	"log"
)

func (cli *CLI) getBalance(address string) {
	if !ValidateAddress(address) {
		log.Panic("Invalid wallet address")
	}

	bc := NewBlockchain()
	utxoSet := UTXOSet{bc}
	defer bc.DB.Close()

	balance := 0
	pubKeyHash := GetPubKeyHash([]byte(address))

	utxos := utxoSet.FindUTXO(pubKeyHash)
	for _, utxo := range utxos {
		balance += utxo.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}
