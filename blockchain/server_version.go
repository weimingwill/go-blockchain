package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

type version struct {
	Version int

	// RemoteAddr is the address where the version from
	RemoteAddr string

	BestHeight int
}

func sendVersion(addr string, bc *Blockchain) {
	log.Println("send version")
	bestHeight := bc.GetBestHeight()

	payload := GobEncode(
		version{
			Version:    nodeVersion,
			RemoteAddr: nodeAddress,
			BestHeight: bestHeight,
		})
	request := append(commandToBytes("version"), payload...)
	sendData(addr, request)
}

func handleVersion(request []byte, bc *Blockchain) {
	log.Println("handle version")
	var buff bytes.Buffer
	var payload version

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	logPanicErr(err)

	myBestHeight := bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	log.Println("myBestHeight", myBestHeight)
	log.Println("foreignerBestHeight", foreignerBestHeight)

	if myBestHeight < foreignerBestHeight {
		// If current blockchain height is smaller than the got-blockchain height
		// Send get blocks request
		sendGetBlocks(payload.RemoteAddr)
	} else if myBestHeight > foreignerBestHeight {
		sendVersion(payload.RemoteAddr, bc)
	}

	if !isNodeKnown(payload.RemoteAddr) {
		knownNodes = append(knownNodes, payload.RemoteAddr)
	}
}
