package blockchain

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"
)

// Wallets stores a collection of wallets
type Wallets struct {
	Wallets map[string]*Wallet
}

// NewWallets creates wallets and load wallet information to it if wallet file exist
func NewWallets() (*Wallets, error) {
	ws := Wallets{}
	ws.Wallets = make(map[string]*Wallet)
	err := ws.LoadFromFile()
	return &ws, err
}

// CreateWallet creates a new wallet and save to it to wallet file
func (ws *Wallets) CreateWallet() string {
	wallet := NewWallet()
	address := string(wallet.GetAddress())

	ws.Wallets[address] = wallet
	ws.SaveToFile()
	return address
}

// GetAddresses returns list of wallet addresses
func (ws *Wallets) GetAddresses() []string {
	var addresses []string
	for k := range ws.Wallets {
		addresses = append(addresses, k)
	}
	return addresses
}

// GetWallet returns a wallet by wallet address
func (ws *Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

// LoadFromFile loads wallets from wallet file if file exists
func (ws *Wallets) LoadFromFile() error {
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	content, err := ioutil.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(content))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	ws.Wallets = wallets.Wallets

	return nil
}

// SaveToFile saves wallets to wallet file
func (ws Wallets) SaveToFile() {
	var content bytes.Buffer

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0600)
	if err != nil {
		log.Panic(err)
	}
}
