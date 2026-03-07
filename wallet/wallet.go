package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"math/big"

	"golang.org/x/crypto/ripemd160"
)

const (
	checksumLength = 4
	version        = byte(0x00)
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

// Implement GobEncode for Wallet
func (w *Wallet) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	// Encode PrivateKey components
	if err := enc.Encode(w.PrivateKey.D); err != nil {
		return nil, err
	}
	if err := enc.Encode(w.PrivateKey.PublicKey.X); err != nil {
		return nil, err
	}
	if err := enc.Encode(w.PrivateKey.PublicKey.Y); err != nil {
		return nil, err
	}

	// Encode PublicKey
	if err := enc.Encode(w.PublicKey); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Implement GobDecode for Wallet
func (w *Wallet) GobDecode(data []byte) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)

	var D, X, Y big.Int

	// Decode PrivateKey components
	if err := dec.Decode(&D); err != nil {
		return err
	}
	if err := dec.Decode(&X); err != nil {
		return err
	}
	if err := dec.Decode(&Y); err != nil {
		return err
	}

	// Reconstruct ECDSA private key
	w.PrivateKey = ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: elliptic.P256(),
			X:     &X,
			Y:     &Y,
		},
		D: &D,
	}

	// Decode PublicKey
	if err := dec.Decode(&w.PublicKey); err != nil {
		return err
	}

	return nil
}

// Rest of your existing functions remain the same...

func (w Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublicKey)
	versionedHash := append([]byte{version}, pubHash...)
	checksum := Checksum(versionedHash)
	fullHash := append(versionedHash, checksum...)
	address := Base58Encode(fullHash)
	return address
}

func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pub
}

func MakeWallet() *Wallet {
	private, public := NewKeyPair()
	return &Wallet{PrivateKey: private, PublicKey: public}
}

func PublicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		log.Panic(err)
	}

	publicRipMD := hasher.Sum(nil)

	return publicRipMD
}

func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checksumLength]
}

func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-checksumLength:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checksumLength]
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}
