package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"syscall"
	"time"

	"github.com/korobkovandrey/runtime-metrics/pkg/sign"
	"go.uber.org/zap"
)

// postData sends data to the server.
func (s *Sender) postData(ctx context.Context, url string, data any, crypt bool) error {
	b, hash, err := s.makeBodyWithHash(data)
	if err != nil {
		return fmt.Errorf("failed to make body with hash: %w", err)
	}

	var encryptedAES string
	if crypt && s.cfg.PublicKey != nil {
		b, encryptedAES, err = makeCryptData(b, s.cfg.PublicKey)
		if err != nil {
			return fmt.Errorf("failed to encrypt data: %w", err)
		}
	}

	postBody, err := makeGzipBuffer(b)
	if err != nil {
		return fmt.Errorf("failed to make gzip buffer: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, postBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Encoding", "gzip")
	if encryptedAES != "" {
		req.Header.Set("X-Encrypted-Key", encryptedAES)
	}
	if hash != "" {
		req.Header.Set("HashSHA256", hash)
	}
	resp, err := s.doRetry(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			s.l.WarnCtx(ctx, "failed to close the resp body", zap.Error(err))
		}
	}()
	return nil
}

// doRetry retries the request.
func (s *Sender) doRetry(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
	for i := 0; ; i++ {
		if err = ctx.Err(); err != nil {
			break
		}
		resp, err = s.client.Do(req)
		if err == nil {
			if resp.StatusCode >= http.StatusOK || resp.StatusCode < http.StatusInternalServerError {
				break
			}
			if err = resp.Body.Close(); err != nil {
				s.l.WarnCtx(ctx, "failed to close body", zap.Error(err))
			}
			err = fmt.Errorf("unexpected status code received: %d", resp.StatusCode)
		} else if !errors.Is(err, syscall.ECONNREFUSED) {
			break
		}
		if i == len(s.cfg.RetryDelays) || ctx.Err() != nil {
			break
		}
		s.l.WarnCtx(ctx, "failed to send request, will retry", zap.Int("attempt", i+1), zap.Error(err))
		time.Sleep(s.cfg.RetryDelays[i])
	}
	if resp != nil && resp.StatusCode >= http.StatusBadRequest {
		return resp, errors.New(http.StatusText(resp.StatusCode))
	}
	return resp, err
}

// makeBodyWithHash makes the body with hash.
func (s *Sender) makeBodyWithHash(data any) (dataBytes []byte, hash string, err error) {
	if data == nil {
		return nil, "", nil
	}
	dataBytes, err = json.Marshal(data)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal data: %w", err)
	}
	return dataBytes, sign.MakeToString(dataBytes, s.cfg.Key), nil
}

// makeGzipBuffer makes the gzip buffer.
func makeGzipBuffer(data []byte) (io.Reader, error) {
	if data == nil {
		return http.NoBody, nil
	}
	buf := bytes.NewBuffer(nil)
	gz := gzip.NewWriter(buf)
	if _, err := gz.Write(data); err != nil {
		return nil, fmt.Errorf("failed to gzip data: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}
	return buf, nil
}

func makeCryptData(data []byte, publicKey *rsa.PublicKey) (dataBytes []byte, encryptedAES string, err error) {
	const aesKeyLen = 32
	aesKey := make([]byte, aesKeyLen)
	if _, err = rand.Read(aesKey); err != nil {
		return nil, "", fmt.Errorf("failed to generate AES key: %w", err)
	}

	encryptedAESKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, aesKey, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to encrypt AES key: %w", err)
	}
	encryptedAES = base64.StdEncoding.EncodeToString(encryptedAESKey)
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create GCM cipher: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	return gcm.Seal(nonce, nonce, data, nil), encryptedAES, nil
}
