package sign

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMake(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		key      []byte
		expected []byte
	}{
		{
			name: "Valid input",
			data: []byte("testdata"),
			key:  []byte("secretkey"),
			expected: func() []byte {
				h := hmac.New(sha256.New, []byte("secretkey"))
				h.Write([]byte("testdata"))
				return h.Sum(nil)
			}(),
		},
		{
			name:     "Empty data",
			data:     []byte{},
			key:      []byte("secretkey"),
			expected: nil,
		},
		{
			name:     "Empty key",
			data:     []byte("testdata"),
			key:      []byte{},
			expected: nil,
		},
		{
			name:     "Both empty",
			data:     []byte{},
			key:      []byte{},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Make(tt.data, tt.key)
			assert.Equal(t, tt.expected, result, "Make() output should match expected")
		})
	}
}

func TestEncodeToString(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "Valid input",
			data:     []byte{0x01, 0x02, 0x03},
			expected: "010203",
		},
		{
			name:     "Empty input",
			data:     []byte{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeToString(tt.data)
			assert.Equal(t, tt.expected, result, "EncodeToString() output should match expected")
		})
	}
}

func TestDecodeString(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected []byte
		wantErr  bool
	}{
		{
			name:     "Valid hex string",
			data:     "010203",
			expected: []byte{0x01, 0x02, 0x03},
			wantErr:  false,
		},
		{
			name:     "Empty string",
			data:     "",
			expected: []byte{},
			wantErr:  false,
		},
		{
			name:     "Invalid hex string",
			data:     "invalid",
			expected: []byte{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeString(tt.data)
			if tt.wantErr {
				assert.Error(t, err, "DecodeString should return an error")
			} else {
				assert.NoError(t, err, "DecodeString should not return an error")
			}
			assert.Equal(t, tt.expected, result, "DecodeString() output should match expected")
		})
	}
}

func TestMakeToString(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		key      []byte
		expected string
	}{
		{
			name: "Valid input",
			data: []byte("testdata"),
			key:  []byte("secretkey"),
			expected: func() string {
				h := hmac.New(sha256.New, []byte("secretkey"))
				h.Write([]byte("testdata"))
				return hex.EncodeToString(h.Sum(nil))
			}(),
		},
		{
			name:     "Empty data",
			data:     []byte{},
			key:      []byte("secretkey"),
			expected: "",
		},
		{
			name:     "Empty key",
			data:     []byte("testdata"),
			key:      []byte{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MakeToString(tt.data, tt.key)
			assert.Equal(t, tt.expected, result, "MakeToString() output should match expected")
		})
	}
}

func TestValidate(t *testing.T) {
	validHash := func(data, key []byte) []byte {
		h := hmac.New(sha256.New, key)
		h.Write(data)
		return h.Sum(nil)
	}

	tests := []struct {
		name     string
		data     []byte
		key      []byte
		hash     []byte
		expected bool
	}{
		{
			name:     "Valid hash",
			data:     []byte("testdata"),
			key:      []byte("secretkey"),
			hash:     validHash([]byte("testdata"), []byte("secretkey")),
			expected: true,
		},
		{
			name:     "Invalid hash",
			data:     []byte("testdata"),
			key:      []byte("secretkey"),
			hash:     []byte("wronghash"),
			expected: false,
		},
		{
			name:     "Empty key",
			data:     []byte("testdata"),
			key:      []byte{},
			hash:     []byte("somehash"),
			expected: true,
		},
		{
			name:     "Empty hash",
			data:     []byte("testdata"),
			key:      []byte("secretkey"),
			hash:     []byte{},
			expected: true,
		},
		{
			name:     "Both empty",
			data:     []byte{},
			key:      []byte{},
			hash:     []byte{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Validate(tt.data, tt.key, tt.hash)
			assert.Equal(t, tt.expected, result, "Validate() output should match expected")
		})
	}
}
