package mcrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"
)

// Middleware is a middleware for decrypting requests.
func Middleware(l *logging.ZapLogger, privateKey *rsa.PrivateKey, routes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost || (len(routes) != 0 && !slices.Contains(routes, r.URL.Path)) {
				next.ServeHTTP(w, r)
				return
			}
			encryptedKey := r.Header.Get("X-Encrypted-Key")
			if encryptedKey == "" {
				if privateKey != nil {
					logBadRequest(l, r, "X-Encrypted-Key is required")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				next.ServeHTTP(w, r)
				return
			}
			if privateKey == nil {
				logBadRequest(l, r, "crypto is disabled")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			aesKey, err := getAESKey(encryptedKey, privateKey)
			if err != nil {
				logBadRequest(l, r, fmt.Errorf("failed to get AES key: %w", err).Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			ciphertext, err := io.ReadAll(r.Body)
			if err == nil {
				err = r.Body.Close()
			}
			if err != nil {
				logBadRequest(l, r, fmt.Errorf("failed to read request body: %w", err).Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			const ciphertextLen = 12
			if len(ciphertext) < ciphertextLen {
				logBadRequest(l, r, "encrypted body too short")
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			nonce := ciphertext[:ciphertextLen]
			ciphertext = ciphertext[ciphertextLen:]
			block, err := aes.NewCipher(aesKey)
			if err != nil {
				logBadRequest(l, r, fmt.Errorf("failed to create AES cipher: %w", err).Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			gcm, err := cipher.NewGCM(block)
			if err != nil {
				logBadRequest(l, r, fmt.Errorf("failed to create GCM: %w", err).Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
			if err != nil {
				logBadRequest(l, r, fmt.Errorf("failed to decrypt body: %w", err).Error())
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(plaintext))
			next.ServeHTTP(w, r)
		})
	}
}

func logBadRequest(l *logging.ZapLogger, r *http.Request, msg string) {
	l.InfoCtx(r.Context(), msg, zap.Int("status", http.StatusBadRequest),
		zap.String("method", r.Method), zap.String("uri", r.RequestURI))
}

func getAESKey(encryptedKey string, privateKey *rsa.PrivateKey) ([]byte, error) {
	encryptedKeyBytes, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode X-Encrypted-Key: %w", err)
	}
	aesKey, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedKeyBytes, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt AES key: %w", err)
	}
	const aesKeyLen = 32
	if len(aesKey) != aesKeyLen {
		return nil, errors.New("invalid AES key length")
	}
	return aesKey, nil
}
