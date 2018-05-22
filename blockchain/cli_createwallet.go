package blockchain

import (
	"fmt"
)

func (cli *CLI) createWallet() {
	wallets, _ := NewWallets()
	address := wallets.CreateWallet()
	fmt.Printf("Your new wallet address: %s\n", address)
}
