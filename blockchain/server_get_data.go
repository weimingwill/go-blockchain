package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"log"
)

type getdata struct {
	RemoteAddr string
	Type       string
	ID         []byte
}

func sendGetData(addr, kind string, id []byte) {
	log.Println("send get data")

	payload := GobEncode(getdata{nodeAddress, kind, id})
	request := append(commandToBytes("getdata"), payload...)

	sendData(addr, request)
}

func handleGetData(request []byte, bc *Blockchain) {
	log.Println("handle get data")

	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		sendBlock(payload.RemoteAddr, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]

		sendTx(payload.RemoteAddr, &tx)
		// delete(mempool, txID)
	}
}
