// Generate RSA keys and save them to files.
// Flags:
//
//	-b bits     default 2048
//	-private    default private_key.pem     path to private key file
//	-public	    default public_key.pem      path to public key file
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
)

// generateAndSaveKeys generates RSA keys and saves them to specified files.
func generateAndSaveKeys(bits int, privatePath, publicPath string) error {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return fmt.Errorf("error generating RSA keys: %w", err)
	}
	publicKey := &privateKey.PublicKey

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	privateKeyBlock := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privateKeyBytes}
	privateFile, err := os.Create(privatePath)
	if err != nil {
		return fmt.Errorf("error creating private key file: %w", err)
	}
	defer func() {
		_ = privateFile.Close()
	}()
	if err = pem.Encode(privateFile, privateKeyBlock); err != nil {
		return fmt.Errorf("error encoding private key: %w", err)
	}
	fmt.Printf("Private key saved to: %v\n", privateFile.Name())

	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return fmt.Errorf("error marshaling public key: %w", err)
	}
	publicKeyBlock := &pem.Block{Type: "RSA PUBLIC KEY", Bytes: publicKeyBytes}
	publicFile, err := os.Create(publicPath)
	if err != nil {
		return fmt.Errorf("error creating public key file: %w", err)
	}
	defer func() {
		_ = publicFile.Close()
	}()
	if err = pem.Encode(publicFile, publicKeyBlock); err != nil {
		return fmt.Errorf("error encoding public key: %w", err)
	}
	fmt.Printf("Public key saved to: %v\n", publicFile.Name())
	return nil
}

func main() {
	const defaultBits = 2048
	fs := flag.NewFlagSet("genrsa", flag.ExitOnError)
	bits := fs.Int("b", defaultBits, "bits")
	privatePath := fs.String("private", "private_key.pem", "path to private key file")
	publicPath := fs.String("public", "public_key.pem", "path to public key file")
	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Printf("Error parsing flags: %v\n", err)
		return
	}

	if err := generateAndSaveKeys(*bits, *privatePath, *publicPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
}
