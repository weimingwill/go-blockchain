package blockchain

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type block struct {
	RemoteAddr string
	Block      []byte
}

func sendBlock(addr string, b *Block) {
	// Todo: why node address here? always to the central?
	payload := GobEncode(block{nodeAddress, b.Serialize()})
	request := append(commandToBytes("block"), payload...)

	sendData(addr, request)
}

// TODO: Instead of trusting unconditionally,
// we should validate every incoming block before adding it to the blockchain.

// TODO: Instead of running UTXOSet.Reindex(),
// UTXOSet.Update(block) should be used, because if blockchain is big,
// it’ll take a lot of time to reindex the whole UTXO set.

func handleBlock(request []byte, bc *Blockchain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	logPanicErr(err)

	blockData := payload.Block
	block := DeserializeBlock(blockData)

	fmt.Println("Recevied a new block!")
	bc.AddBlock(block)

	fmt.Printf("Added block %x\n", block.Hash)

	// When blocksInTransit sending request to get block again
	// A smart way to do the iteration
	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.RemoteAddr, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := UTXOSet{bc}
		UTXOSet.Reindex()
	}
}
