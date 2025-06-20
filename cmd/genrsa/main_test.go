package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGenerateKeys tests the generation and saving of RSA keys
func TestGenerateKeys(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "rsa_test")
	require.NoError(t, err, "Ошибка создания временной директории")
	defer func() {
		assert.NoError(t, os.RemoveAll(tempDir))
	}()

	privatePath := filepath.Join(tempDir, "private_key.pem")
	publicPath := filepath.Join(tempDir, "public_key.pem")

	runMainTesting(2048, privatePath, publicPath)

	_, err = os.Stat(privatePath)
	assert.NoError(t, err)
	_, err = os.Stat(publicPath)
	assert.NoError(t, err)

	privateKeyPEM, err := os.ReadFile(privatePath)
	require.NoError(t, err)
	privateBlock, _ := pem.Decode(privateKeyPEM)
	assert.NotNil(t, privateBlock)
	assert.Equal(t, "RSA PRIVATE KEY", privateBlock.Type)
	privateKey, err := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
	require.NoError(t, err)

	publicKeyPEM, err := os.ReadFile(publicPath)
	require.NoError(t, err)
	publicBlock, _ := pem.Decode(publicKeyPEM)
	assert.NotNil(t, publicBlock)
	assert.Equal(t, "RSA PUBLIC KEY", publicBlock.Type)
	publicKey, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	require.NoError(t, err)
	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	assert.True(t, ok)

	message := []byte("Тестовое сообщение")
	ciphertext, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPublicKey, message, nil)
	require.NoError(t, err)
	plaintext, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, ciphertext, nil)
	require.NoError(t, err)
	assert.Equal(t, message, plaintext)
}

// TestInvalidBits tests the generation of RSA keys with invalid bit size
func TestInvalidBits(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "rsa_test")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, os.RemoveAll(tempDir))
	}()

	privatePath := filepath.Join(tempDir, "private_key.pem")
	publicPath := filepath.Join(tempDir, "public_key.pem")

	err = generateAndSaveKeys(512, privatePath, publicPath)
	assert.Error(t, err)
}

// TestFileCreationError tests the error handling when trying to create files in a read-only directory
func TestFileCreationError(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "rsa_test")
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, os.RemoveAll(tempDir))
	}()

	err = os.Chmod(tempDir, 0400)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, os.Chmod(tempDir, 0755))
	}()

	privatePath := filepath.Join(tempDir, "private_key.pem")
	publicPath := filepath.Join(tempDir, "public_key.pem")

	err = generateAndSaveKeys(2048, privatePath, publicPath)
	assert.Error(t, err)
}

func runMainTesting(bits int, privatePath, publicPath string) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"genrsa", "-b", fmt.Sprintf("%d", bits), "-private", privatePath, "-public", publicPath}
	main()
}
