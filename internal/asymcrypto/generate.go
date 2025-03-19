package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	err := GeneratePair()
	if err != nil {
		fmt.Println(err)
	}
}

func GeneratePair() error {
	const fmd os.FileMode = 0o666

	pkb := 4096
	privateFN := "keys/private.pem"
	publicFN := "keys/public.pem"

	privateKey, err := rsa.GenerateKey(rand.Reader, pkb)
	if err != nil {
		return fmt.Errorf("generatePair->GenerateKey: %w", err)
	}

	publicKey := &privateKey.PublicKey

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})

	err = os.WriteFile(privateFN, privateKeyPEM, fmd)
	if err != nil {
		return fmt.Errorf("generatePair->WriteFile->pri: %w", err)
	}

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("generatePair->MarshalPKIXPub: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	err = os.WriteFile(publicFN, publicKeyPEM, fmd)
	if err != nil {
		return fmt.Errorf("generatePair->WriteFile-publ: %w", err)
	}

	return nil
}
