package blockchain

import (
	"fmt"
	"log"
)

func (cli *CLI) send(from, to string, amount int) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := NewBlockchain()
	utxoSet := UTXOSet{bc}
	defer bc.DB.Close()

	tx := NewUTXOTransaction(from, to, amount, &utxoSet)

	// Give reward to the mining
	cbTx := NewCoinbaseTX(from, "")
	newBlock := bc.MineBlock([]*Transaction{cbTx, tx})
	utxoSet.Update(newBlock)

	fmt.Println("Success!")
}
