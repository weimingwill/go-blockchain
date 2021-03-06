package blockchain

import "fmt"

func (cli *CLI) reindexUTXO(nodeID string) {

	bc := NewBlockchain(nodeID)
	utxoSet := UTXOSet{bc}
	defer bc.DB.Close()

	utxoSet.Reindex()

	count := utxoSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}
