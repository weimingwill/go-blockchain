package blockchain

import (
	"fmt"
	"log"
)

func (cli *CLI) createBlockchain(address string) {
	if !ValidateAddress(address) {
		log.Panic("Invalid wallet address")
	}

	bc := CreateBlockchain(address)
	us := UTXOSet{bc}
	us.Reindex()
	defer bc.DB.Close()
	fmt.Println("Done!")
}
