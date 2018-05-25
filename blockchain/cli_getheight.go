package blockchain

import "fmt"

func (cli *CLI) getBlockchainHeight(nodeID string) {
	bc := NewBlockchain(nodeID)
	fmt.Println("Blockchain height:", bc.GetBestHeight())
}
