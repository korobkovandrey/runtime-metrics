package main

import (
	"os"
	"path/filepath"
)

// ExampleGenerateAndSaveKeys demonstrates how to generate RSA keys and save them to files.
func Example_main() {
	const (
		privatePath = "example_private_key.pem"
		publicPath  = "example_public_key.pem"
	)

	runMainTesting(2048, privatePath, publicPath)
	_ = os.Remove(privatePath)
	_ = os.Remove(publicPath)

	// Output:
	// Private key saved to: example_private_key.pem
	// Public key saved to: example_public_key.pem
}

// Example_main_err_permission_denied demonstrates the error handling when trying to create files in a read-only directory.
func Example_main_err_permission_denied() {
	const dir = "rsa_example"
	runMainTesting(2048, filepath.Join(dir, "example_private_key.pem"), filepath.Join(dir, "example_public_key.pem"))

	// Output:
	// Error: error creating private key file: open rsa_example/example_private_key.pem: no such file or directory
}

// Example_main_err_incorrect_bits demonstrates the error handling when trying to generate RSA keys with an insecure bit size.
func Example_main_err_incorrect_bits() {
	runMainTesting(512, "example_private_key.pem", "example_public_key.pem")

	// Output:
	// Error: error generating RSA keys: crypto/rsa: 512-bit keys are insecure (see https://go.dev/pkg/crypto/rsa#hdr-Minimum_key_size)
}
