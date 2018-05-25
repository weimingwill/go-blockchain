package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

type getblocks struct {
	RemoteAddr string
}

func sendGetBlocks(addr string) {
	log.Println("send get blocks")

	payload := GobEncode(getblocks{nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)

	sendData(addr, request)
}

func handleGetBlocks(request []byte, bc *Blockchain) {
	log.Println("handle get blocks")

	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	logPanicErr(err)

	blocks := bc.GetBlockHashes()
	sendInv(payload.RemoteAddr, "block", blocks)
}
