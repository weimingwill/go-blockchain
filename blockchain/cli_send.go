package blockchain

import (
	"fmt"
	"log"
)

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := NewBlockchain(nodeID)
	utxoSet := UTXOSet{bc}
	defer bc.DB.Close()

	wallets, err := NewWallets(nodeID)
	logPanicErr(err)
	wallet := wallets.GetWallet(from)

	tx := NewUTXOTransaction(&wallet, to, amount, &utxoSet)

	if mineNow {
		// Give reward to the mining
		cbTx := NewCoinbaseTX(from, "")
		newBlock := bc.MineBlock([]*Transaction{cbTx, tx})
		utxoSet.Update(newBlock)
	} else {
		sendTx(knownNodes[0], tx)
	}

	fmt.Println("Success!")
}
