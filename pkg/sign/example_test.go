package sign_test

import (
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/pkg/sign"
)

func Example() {
	data := []byte("Hello, World!")
	key := []byte("secret-key")

	// Create HMAC-SHA256 hash
	hash := sign.Make(data, key)

	// Encode hash to hexadecimal string
	hexHash := sign.EncodeToString(hash)

	// Decode hexadecimal string back to bytes
	decodedHash, err := sign.DecodeString(hexHash)
	if err != nil {
		fmt.Printf("Error decoding hash: %v\n", err)
		return
	}

	// Validate the hash
	valid := sign.Validate(data, key, decodedHash)
	fmt.Printf("Hash is valid: %v\n", valid)
	// Output: Hash is valid: true
}

func ExampleMake() {
	data := []byte("Hello, World!")
	key := []byte("secret-key")
	hash := sign.Make(data, key)
	fmt.Printf("HMAC-SHA256 hash length: %d bytes\n", len(hash))
	// Output: HMAC-SHA256 hash length: 32 bytes
}

func ExampleEncodeToString() {
	data := []byte{0x01, 0x02, 0x03}
	hexString := sign.EncodeToString(data)
	fmt.Println("Hex encoded string:", hexString)
	// Output: Hex encoded string: 010203
}

func ExampleDecodeString() {
	hexString := "010203"
	data, err := sign.DecodeString(hexString)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("Decoded bytes: %v\n", data)
	// Output: Decoded bytes: [1 2 3]
}

func ExampleMakeToString() {
	data := []byte("Hello, World!")
	key := []byte("secret-key")
	hexHash := sign.MakeToString(data, key)
	fmt.Println("Hex encoded HMAC-SHA256:", hexHash)
	// Output will be a 64-character hexadecimal string
}

func ExampleValidate() {
	data := []byte("Hello, World!")
	key := []byte("secret-key")
	hash := sign.Make(data, key)

	// Valid case
	valid := sign.Validate(data, key, hash)
	fmt.Println("Valid hash:", valid)

	// Invalid case
	invalidHash := []byte("invalid-hash")
	invalid := sign.Validate(data, key, invalidHash)
	fmt.Println("Invalid hash:", invalid)
	// Output:
	// Valid hash: true
	// Invalid hash: false
}
