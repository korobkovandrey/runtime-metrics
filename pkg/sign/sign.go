// Package sign contains the sign logic.
package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Make returns the HMAC-SHA256 hash of the given data and key.
func Make(data, key []byte) []byte {
	if len(data) == 0 || len(key) == 0 {
		return nil
	}
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// EncodeToString returns the hexadecimal encoding of the given data.
func EncodeToString(data []byte) string {
	return hex.EncodeToString(data)
}

// DecodeString decodes the given hexadecimal string into a byte slice.
func DecodeString(data string) ([]byte, error) {
	//nolint:wrapcheck // ignore
	return hex.DecodeString(data)
}

// MakeToString returns the hexadecimal encoding of the HMAC-SHA256 hash of the given data and key.
func MakeToString(data, key []byte) string {
	return EncodeToString(Make(data, key))
}

// Validate checks if the given hash is valid for the given data and key.
func Validate(data, key, hash []byte) bool {
	if len(key) == 0 || len(hash) == 0 {
		return true
	}
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return hmac.Equal(hash, h.Sum(nil))
}
