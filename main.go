package main

import (
	"go-blockchain/blockchain"
)

func main() {
	bc := blockchain.NewBlockchain()
	defer bc.DB.Close()

	cli := blockchain.NewCLI(bc)
	cli.Run()
}
