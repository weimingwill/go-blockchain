package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"go-blockchain/base58"

	"golang.org/x/crypto/ripemd160"
)

const (
	verzion            = byte(0x00)
	addressChecksumLen = 4
)

// Wallet stores private and public keys
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// NewWallet creates and returns a Wallet
func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{private, public}
	return &wallet
}

// GetAddress returns wallet address
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)
	versionPaylod := append([]byte{verzion}, pubKeyHash...)
	checksum := checksum(versionPaylod)

	fullPayload := append(versionPaylod, checksum...)
	address := base58.Encode(fullPayload)
	return address
}

// HashPubKey hashes public key
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

// GetPubKeyHash returns public key hash with address
func GetPubKeyHash(address []byte) []byte {
	fullPayload := base58.Decode(address)
	return fullPayload[1 : len(fullPayload)-addressChecksumLen]
}

// ValidateAddress check if address if valid
func ValidateAddress(address string) bool {
	fullPayload := base58.Decode([]byte(address))
	expectChecksum := fullPayload[len(fullPayload)-addressChecksumLen:]
	version := fullPayload[0]
	pubKeyHash := fullPayload[1 : len(fullPayload)-addressChecksumLen]

	gotChecksum := checksum(append([]byte{version}, pubKeyHash...))
	return bytes.Compare(expectChecksum, gotChecksum) == 0
}

// Checksum generates a checksum for a public key
func checksum(payload []byte) []byte {
	firstSHA256 := sha256.Sum256(payload)
	secondSHA256 := sha256.Sum256(firstSHA256[:])

	return secondSHA256[:addressChecksumLen]
}

// newKeyPair genenrates a new pair of private key and public key.
// Private key is generated using elliptic curve criptography
func newKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pubKey
}
