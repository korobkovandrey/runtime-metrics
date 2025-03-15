package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func Make(data, key []byte) []byte {
	if len(data) == 0 || len(key) == 0 {
		return nil
	}
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func EncodeToString(data []byte) string {
	return hex.EncodeToString(data)
}

func DecodeString(data string) ([]byte, error) {
	return hex.DecodeString(data)
}

func MakeToString(data, key []byte) string {
	return EncodeToString(Make(data, key))
}

func Validate(data, key, hash []byte) bool {
	if len(key) == 0 || len(hash) == 0 {
		return true
	}
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return hmac.Equal(hash, h.Sum(nil))
}
