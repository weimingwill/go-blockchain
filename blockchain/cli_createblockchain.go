package blockchain

import (
	"fmt"
	"log"
)

func (cli *CLI) createBlockchain(address string, nodeID string) {
	if !ValidateAddress(address) {
		log.Panic("Invalid wallet address")
	}

	bc := CreateBlockchain(address, nodeID)
	us := UTXOSet{bc}
	us.Reindex()
	defer bc.DB.Close()
	fmt.Println("Done!")
}
