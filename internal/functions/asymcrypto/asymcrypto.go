package asymcrypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
)

var errCast = errors.New("cast err")

func Encrypt(data *[]byte, key *[]byte) (*[]byte, error) {
	publicKeyBlock, _ := pem.Decode(*key)

	publicKey, err := x509.ParsePKIXPublicKey(
		publicKeyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf(
			"Encrypt->ParsePKIXPublicKey: %w", err)
	}

	rsakey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errCast
	}

	enc, err := rsa.EncryptPKCS1v15(
		rand.Reader,
		rsakey,
		*data)
	if err != nil {
		return nil, fmt.Errorf(
			"Encrypt->EncryptPKCS1v15: %w", err)
	}

	return &enc, nil
}

func Decrypt(data *[]byte, key *[]byte) (*[]byte, error) {
	privateKeyBlock, _ := pem.Decode(*key)

	privateKey, err := x509.ParsePKCS1PrivateKey(
		privateKeyBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf(
			"Decrypt->ParsePKCS1PrivateKey: %w", err)
	}

	dec, err := rsa.DecryptPKCS1v15(
		rand.Reader,
		privateKey,
		*data)
	if err != nil {
		return nil, fmt.Errorf(
			"Decrypt->DecryptPKCS1v15: %w", err)
	}

	return &dec, nil
}
