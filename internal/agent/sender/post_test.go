package sender

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createRSAKeyPair(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return privateKey, &privateKey.PublicKey
}

func decryptAESData(t *testing.T, ciphertext []byte, encryptedAES string, privateKey *rsa.PrivateKey) []byte {
	t.Helper()
	encryptedAESKey, err := base64.StdEncoding.DecodeString(encryptedAES)
	require.NoError(t, err)
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedAESKey, nil)
	require.NoError(t, err)
	require.Equal(t, 32, len(aesKey))

	require.GreaterOrEqual(t, len(ciphertext), 12)
	nonce := ciphertext[:12]
	ciphertext = ciphertext[12:]

	block, err := aes.NewCipher(aesKey)
	require.NoError(t, err)

	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	require.NoError(t, err)
	return plaintext
}

func TestMakeCryptData(t *testing.T) {
	privateKey, publicKey := createRSAKeyPair(t)

	tests := []struct {
		publicKey      *rsa.PublicKey
		name           string
		wantErrMessage string
		data           []byte
		wantError      bool
	}{
		{
			name:      "successful encryption",
			data:      []byte("Hello, world!"),
			publicKey: publicKey,
			wantError: false,
		},
		{
			name:      "empty data",
			data:      nil,
			publicKey: publicKey,
			wantError: false,
		},
		{
			name:           "invalid public key",
			data:           []byte("test"),
			publicKey:      &rsa.PublicKey{N: nil, E: 0}, // Некорректный ключ
			wantError:      true,
			wantErrMessage: "failed to encrypt AES key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataBytes, encryptedAES, err := makeCryptData(tt.data, tt.publicKey)
			if tt.wantError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrMessage)
				assert.Nil(t, dataBytes)
				assert.Empty(t, encryptedAES)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, dataBytes)
			assert.NotEmpty(t, encryptedAES)

			plaintext := decryptAESData(t, dataBytes, encryptedAES, privateKey)
			assert.Equal(t, tt.data, plaintext)
		})
	}
}
