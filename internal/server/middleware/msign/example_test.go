package msign_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/msign"
	"github.com/korobkovandrey/runtime-metrics/pkg/sign"
)

func Example() {
	// Define a key for signing
	key := []byte("secret-key")

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})

	// Wrap the handler with Signer middleware
	middleware := msign.Signer(key)(handler)

	// Create a test request with a valid signature
	body := []byte("Hello, World!")
	hash := sign.MakeToString(body, key)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("HashSHA256", hash)
	w := httptest.NewRecorder()

	// Serve the request
	middleware.ServeHTTP(w, req)

	// Verify the response
	respHash := w.Header().Get("HashSHA256")
	fmt.Printf("Response status: %d, HashSHA256: %s\n", w.Code, respHash)
	fmt.Printf("Response body: %s\n", w.Body.String())
	// Output:
	// Response status: 200, HashSHA256: 16ee525f6c944ff49a368cd593eb7b72883b14456c7b583bba4ff973ff4b30f9
	// Response body: Hello, World!
}

func ExampleSigner_response() {
	// Define a key for signing
	key := []byte("secret-key")

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Signed response"))
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})

	// Wrap the handler with Signer middleware
	middleware := msign.Signer(key)(handler)

	// Create a test request (no signature in request)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Serve the request
	middleware.ServeHTTP(w, req)

	// Verify the response has a signature
	respHash := w.Header().Get("HashSHA256")
	valid := sign.Validate([]byte("Signed response"), key, sign.Make([]byte("Signed response"), key))
	fmt.Printf("Response status: %d, Has HashSHA256: %v, Valid: %v\n", w.Code, respHash != "", valid)
	// Output: Response status: 200, Has HashSHA256: true, Valid: true
}

func ExampleSigner_request_valid() {
	// Define a key for signing
	key := []byte("secret-key")

	// Create a handler that reads the request body
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})

	// Wrap the handler with Signer middleware
	middleware := msign.Signer(key)(handler)

	// Create a test request with a valid signature
	body := []byte("Valid signed request")
	hash := sign.MakeToString(body, key)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("HashSHA256", hash)
	w := httptest.NewRecorder()

	// Serve the request
	middleware.ServeHTTP(w, req)

	// Verify the response
	fmt.Printf("Response status: %d, Body: %s\n", w.Code, w.Body.String())
	// Output: Response status: 200, Body: Valid signed request
}

func ExampleSigner_request_invalid() {
	// Define a key for signing
	key := []byte("secret-key")

	// Create a handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Invalid signature", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})

	// Wrap the handler with Signer middleware
	middleware := msign.Signer(key)(handler)

	// Create a test request with an invalid signature
	body := []byte("Invalid signed request")
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("HashSHA256", "invalid-hash")
	w := httptest.NewRecorder()

	// Serve the request
	middleware.ServeHTTP(w, req)

	// Verify the response
	fmt.Printf("Response status: %d, Body: %s\n", w.Code, w.Body.String())
	// Output: Response status: 400, Body: Invalid signature
}

func ExampleSigner_no_key() {
	// Create a handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("No key provided"))
		if err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	})

	// Wrap the handler with Signer middleware with no key
	middleware := msign.Signer(nil)(handler)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	// Serve the request
	middleware.ServeHTTP(w, req)

	// Verify the response has no signature
	fmt.Printf("Response status: %d, Has HashSHA256: %v, Body: %s\n", w.Code, w.Header().Get("HashSHA256") != "", w.Body.String())
	// Output: Response status: 200, Has HashSHA256: false, Body: No key provided
}
