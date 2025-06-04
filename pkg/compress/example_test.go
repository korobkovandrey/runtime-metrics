package compress_test

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	"github.com/korobkovandrey/runtime-metrics/pkg/compress"
)

func Example() {
	// Create a test HTTP response writer
	w := httptest.NewRecorder()

	// Initialize compress Writer
	cw := compress.NewCompressWriter(w)

	// Set content type to enable compression
	cw.Header().Set("Content-Type", "text/html")
	cw.WriteHeader(http.StatusOK)

	// Write and compress data
	data := []byte("<html>Hello, World!</html>")
	_, err := cw.Write(data)
	if err != nil {
		fmt.Printf("Error writing data: %v\n", err)
		return
	}
	if err = cw.Close(); err != nil {
		fmt.Printf("Error closing writer: %v\n", err)
		return
	}

	// Decompress the written data
	r := io.NopCloser(bytes.NewReader(w.Body.Bytes()))
	cr, err := compress.NewCompressReader(r)
	if err != nil {
		fmt.Printf("Error creating reader: %v\n", err)
		return
	}
	defer func() {
		_ = cr.Close()
	}()

	// Read decompressed data
	decompressed := make([]byte, 100)
	n, err := cr.Read(decompressed)
	if err != nil && !errors.Is(err, io.EOF) {
		fmt.Printf("Error reading data: %v\n", err)
		return
	}

	fmt.Printf("Decompressed %d bytes: %s\n", n, string(decompressed[:n]))
	// Output: Decompressed 26 bytes: <html>Hello, World!</html>
}

func ExampleWriter() {
	// Create a test HTTP response writer
	w := httptest.NewRecorder()

	// Initialize compress Writer
	cw := compress.NewCompressWriter(w)

	// Set content type to enable compression
	cw.Header().Set("Content-Type", "text/html")

	// Write HTTP header with status OK
	cw.WriteHeader(http.StatusOK)

	// Write sample data
	data := []byte("<html>Hello, World!</html>")
	n, err := cw.Write(data)
	if err != nil {
		fmt.Printf("Error writing data: %v\n", err)
		return
	}

	// Close the writer
	if err := cw.Close(); err != nil {
		fmt.Printf("Error closing writer: %v\n", err)
		return
	}

	fmt.Printf("Wrote %d bytes, Content-Encoding: %s\n", n, w.Header().Get("Content-Encoding"))
	// Output: Wrote 26 bytes, Content-Encoding: gzip
}

func ExampleWriter_json() {
	// Create a test HTTP response writer
	w := httptest.NewRecorder()

	// Initialize compress Writer
	cw := compress.NewCompressWriter(w)

	// Set content type to JSON to enable compression
	cw.Header().Set("Content-Type", "application/json")

	// Write HTTP header with status OK
	cw.WriteHeader(http.StatusOK)

	// Write sample JSON data
	data := []byte(`{"message": "Hello, World!"}`)
	n, err := cw.Write(data)
	if err != nil {
		fmt.Printf("Error writing data: %v\n", err)
		return
	}

	// Close the writer
	if err := cw.Close(); err != nil {
		fmt.Printf("Error closing writer: %v\n", err)
		return
	}

	fmt.Printf("Wrote %d bytes, Content-Encoding: %s\n", n, w.Header().Get("Content-Encoding"))
	// Output: Wrote 28 bytes, Content-Encoding: gzip
}

func ExampleReader() {
	// Create a compressed data buffer
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	_, err := gz.Write([]byte("Hello, World!"))
	if err != nil {
		fmt.Printf("Error compressing data: %v\n", err)
		return
	}
	if err = gz.Close(); err != nil {
		fmt.Printf("Error closing gzip writer: %v\n", err)
		return
	}

	// Create a compress Reader
	r := io.NopCloser(bytes.NewReader(buf.Bytes()))
	cr, err := compress.NewCompressReader(r)
	if err != nil {
		fmt.Printf("Error creating reader: %v\n", err)
		return
	}
	defer func() {
		_ = cr.Close()
	}()

	// Read decompressed data
	data := make([]byte, 100)
	n, err := cr.Read(data)
	if err != nil && !errors.Is(err, io.EOF) {
		fmt.Printf("Error reading data: %v\n", err)
		return
	}

	fmt.Printf("Read %d bytes: %s\n", n, string(data[:n]))
	// Output: Read 13 bytes: Hello, World!
}
