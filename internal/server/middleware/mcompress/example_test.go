package mcompress_test

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/mcompress"
	"github.com/korobkovandrey/runtime-metrics/pkg/compress"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"
)

func Example() {
	logger, err := logging.NewZapLogger(zap.InfoLevel)
	if err != nil {
		fmt.Printf("Error creating logger: %v\n", err)
		return
	}

	// Create a simple handler that writes a response
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("<html>Hello, World!</html>")); err != nil {
			logger.ErrorCtx(r.Context(), fmt.Errorf("failed to write response: %w", err).Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	// Wrap the handler with GzipCompressed middleware
	middleware := mcompress.GzipCompressed(logger)(handler)

	// Create a test request with Accept-Encoding: gzip
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	// Serve the request
	middleware.ServeHTTP(w, req)

	// Check if response is compressed
	fmt.Printf("Content-Encoding: %s\n", w.Header().Get("Content-Encoding"))
	// Output: Content-Encoding: gzip
}

func ExampleGzipCompressed_response() {
	logger, err := logging.NewZapLogger(zap.InfoLevel)
	if err != nil {
		fmt.Printf("Error creating logger: %v\n", err)
		return
	}

	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(`{"message": "Hello, World!"}`))
		if err != nil {
			logger.ErrorCtx(r.Context(), fmt.Errorf("failed to write response: %w", err).Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	// Wrap the handler with GzipCompressed middleware
	middleware := mcompress.GzipCompressed(logger)(handler)

	// Create a test request with Accept-Encoding: gzip
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()

	// Serve the request
	middleware.ServeHTTP(w, req)

	// Verify the response is compressed
	fmt.Printf("Content-Encoding: %s\n", w.Header().Get("Content-Encoding"))

	// Decompress the response to verify content
	reader, err := compress.NewCompressReader(io.NopCloser(w.Body))
	if err != nil {
		fmt.Printf("Error creating compress reader: %v\n", err)
		return
	}
	defer func() {
		_ = reader.Close()
	}()

	data, err := io.ReadAll(reader)
	if err != nil {
		fmt.Printf("Error reading decompressed data: %v\n", err)
		return
	}

	fmt.Printf("Decompressed content: %s\n", string(data))
	// Output:
	// Content-Encoding: gzip
	// Decompressed content: {"message": "Hello, World!"}
}

func ExampleGzipCompressed_request() {
	logger, err := logging.NewZapLogger(zap.InfoLevel)
	if err != nil {
		fmt.Printf("Error creating logger: %v\n", err)
		return
	}

	// Create a handler that reads the request body
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data []byte
		data, err = io.ReadAll(r.Body)
		if err != nil {
			logger.ErrorCtx(r.Context(), fmt.Errorf("failed to read request body: %w", err).Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(data)
		if err != nil {
			logger.ErrorCtx(r.Context(), fmt.Errorf("failed to write response: %w", err).Error())
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	// Wrap the handler with GzipCompressed middleware
	middleware := mcompress.GzipCompressed(logger)(handler)

	// Create a compressed request body
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err = gz.Write([]byte("Compressed request data"))
	if err != nil {
		fmt.Printf("Error compressing data: %v\n", err)
		return
	}
	if err := gz.Close(); err != nil {
		fmt.Printf("Error closing gzip writer: %v\n", err)
		return
	}

	// Create a test request with Content-Encoding: gzip
	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip")
	w := httptest.NewRecorder()

	// Serve the request
	middleware.ServeHTTP(w, req)

	// Print the response body (should be the decompressed request data)
	fmt.Printf("Response body: %s\n", w.Body.String())
	// Output: Response body: Compressed request data
}
