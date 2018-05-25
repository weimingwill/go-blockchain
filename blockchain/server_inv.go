package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

type inv struct {
	RemoteAddr string
	Type       string
	Items      [][]byte
}

func sendInv(addr string, kind string, items [][]byte) {
	log.Println("send inv")

	// Todo: why send to nodeAddrsess ?
	payload := GobEncode(inv{nodeAddress, kind, items})
	request := append(commandToBytes("inv"), payload...)

	sendData(addr, request)
}

func handleInv(request []byte, bc *Blockchain) {
	log.Println("handle inv")

	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		sendGetData(payload.RemoteAddr, "block", blockHash)

		// Remove the block hash that has sent get data
		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.RemoteAddr, "tx", txID)
		}
	}

}
