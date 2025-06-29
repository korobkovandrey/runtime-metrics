package mcrypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/mlogger"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockHandler returns the next handler which returns the request body
func mockHandler(t *testing.T) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}
		_, err = w.Write(body)
		if err != nil {
			t.Fatalf("failed to write response: %v", err)
		}
	})
}

// createRSAKeyPair creates an RSA key pair
func createRSAKeyPair(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return privateKey
}

// createEncryptedAESKey creates an RSA-encrypted AES key
func createEncryptedAESKey(t *testing.T, publicKey *rsa.PublicKey, aesKey []byte) string {
	t.Helper()
	if aesKey == nil {
		aesKey = make([]byte, 32)
		_, err := rand.Read(aesKey)
		require.NoError(t, err)
	}
	encryptedAESKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, aesKey, nil)
	require.NoError(t, err)
	return base64.StdEncoding.EncodeToString(encryptedAESKey)
}

// createEncryptedRequest creates an encrypted request with AES-GCM and RSA-encrypted key
func createEncryptedRequest(t *testing.T, publicKey *rsa.PublicKey, plaintext []byte) ([]byte, string) {
	t.Helper()
	aesKey := make([]byte, 32)
	_, err := rand.Read(aesKey)
	require.NoError(t, err)

	encryptedAESKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, aesKey, nil)
	require.NoError(t, err)
	encodedKey := base64.StdEncoding.EncodeToString(encryptedAESKey)

	block, err := aes.NewCipher(aesKey)
	require.NoError(t, err)
	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)
	nonce := make([]byte, 12)
	_, err = rand.Read(nonce)
	require.NoError(t, err)
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)
	body := append(nonce, ciphertext...)

	return body, encodedKey
}

func TestMiddleware(t *testing.T) {
	privateKey := createRSAKeyPair(t)
	l, err := logging.NewZapLogger(zap.InfoLevel)
	require.NoError(t, err)
	defer l.Sync()

	tests := []struct {
		privateKey     *rsa.PrivateKey
		name           string
		headerKey      string
		route          string
		wantBody       string
		wantLogMessage string
		body           []byte
		wantStatus     int
		encrypt        bool
	}{
		{
			name:       "no encryption",
			headerKey:  "",
			route:      "/add/",
			body:       []byte("plain text"),
			privateKey: privateKey,
			wantStatus: http.StatusOK,
			wantBody:   "plain text",
		},
		{
			name:           "without header",
			headerKey:      "",
			route:          "/updates/",
			body:           []byte("plain text"),
			privateKey:     privateKey,
			wantStatus:     http.StatusBadRequest,
			wantLogMessage: "X-Encrypted-Key is required",
		},
		{
			name:       "without header, disable crypt",
			headerKey:  "",
			route:      "/updates/",
			body:       []byte("plain text"),
			wantStatus: http.StatusOK,
			wantBody:   "plain text",
		},
		{
			name:       "valid encrypted request",
			encrypt:    true,
			headerKey:  "",
			route:      "/updates/",
			body:       []byte("Hello, world!"),
			privateKey: privateKey,
			wantStatus: http.StatusOK,
			wantBody:   "Hello, world!",
		},
		{
			name:           "invalid base64 header",
			encrypt:        true,
			headerKey:      "not-base64-!!!",
			route:          "/updates/",
			body:           []byte("body"),
			privateKey:     privateKey,
			wantStatus:     http.StatusBadRequest,
			wantLogMessage: "failed to get AES key: failed to decode X-Encrypted-Key: illegal base64 data at input byte 3",
		},
		{
			name:           "invalid length AES key",
			encrypt:        true,
			headerKey:      createEncryptedAESKey(t, &privateKey.PublicKey, []byte("short")),
			route:          "/updates/",
			body:           []byte("body"),
			privateKey:     privateKey,
			wantStatus:     http.StatusBadRequest,
			wantLogMessage: "failed to get AES key: invalid AES key length",
		},
		{
			name:           "body too short",
			headerKey:      createEncryptedAESKey(t, &privateKey.PublicKey, nil),
			route:          "/updates/",
			body:           []byte("short"),
			privateKey:     privateKey,
			wantStatus:     http.StatusBadRequest,
			wantLogMessage: "encrypted body too short",
		},
		{
			name:           "invalid ciphertext",
			encrypt:        true,
			route:          "/updates/",
			body:           []byte("Hello, world!"),
			privateKey:     privateKey,
			wantStatus:     http.StatusBadRequest,
			wantLogMessage: "failed to decrypt body: cipher: message authentication failed",
		},
		{
			name:           "no private key",
			encrypt:        true,
			route:          "/updates/",
			body:           []byte("Hello, world!"),
			privateKey:     nil,
			wantStatus:     http.StatusBadRequest,
			wantLogMessage: "crypto is disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var headerKey string
			var body []byte

			if tt.encrypt {
				body, headerKey = createEncryptedRequest(t, &privateKey.PublicKey, tt.body)
				if tt.name == "invalid ciphertext" {
					body[len(body)-1] ^= 0xFF
				}
				if tt.headerKey != "" {
					headerKey = tt.headerKey
				}
			} else {
				headerKey = tt.headerKey
				body = tt.body
			}

			req, err := http.NewRequest("POST", tt.route, bytes.NewReader(body))
			require.NoError(t, err)
			if headerKey != "" {
				req.Header.Set("X-Encrypted-Key", headerKey)
			}

			rr := httptest.NewRecorder()
			ctx := t.Context()
			req = req.WithContext(ctx)

			middleware := Middleware(tt.privateKey, "/updates/")(mockHandler(t))
			middleware.ServeHTTP(rr, req)

			assert.Equal(t, tt.wantStatus, rr.Code)
			if tt.wantStatus == http.StatusOK {
				assert.Equal(t, tt.wantBody, rr.Body.String())
			}
			if tt.wantLogMessage != "" {
				m, ok := req.Context().Value(mlogger.LogMessageKey).(string)
				assert.True(t, ok)
				assert.Equal(t, tt.wantLogMessage, m)
			} else {
				_, ok := req.Context().Value(mlogger.LogMessageKey).(string)
				assert.False(t, ok)
			}
		})
	}
}
