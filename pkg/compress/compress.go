// Package compress provides a way to compress and decompress data.
//
// It provides a io.Writer implementation that can be used to compress data
// and a io.Reader implementation that can be used to decompress data.
//
// The package also provides a http.ResponseWriter implementation that can be
// used to compress responses of a HTTP server.
package compress

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Writer is a io.Writer that can be used to compress data.
type Writer struct {
	w            http.ResponseWriter
	zw           *gzip.Writer
	Compressible bool
}

// NewCompressWriter returns a new Writer that can be used to compress data.
func NewCompressWriter(w http.ResponseWriter) *Writer {
	return &Writer{
		w:            w,
		Compressible: false,
	}
}

// Header returns the http.Header of the underlying http.ResponseWriter.
func (w *Writer) Header() http.Header {
	return w.w.Header()
}

// Write writes data to the underlying http.ResponseWriter.
func (w *Writer) Write(p []byte) (int, error) {
	if w.Compressible {
		if w.zw == nil {
			w.zw = gzip.NewWriter(w.w)
		}
		n, err := w.zw.Write(p)
		if err != nil {
			return n, fmt.Errorf("compress[Writer].Write: %w", err)
		}
		return n, nil
	}
	n, err := w.w.Write(p)
	if err != nil {
		return n, fmt.Errorf("compress[Writer].Write: %w", err)
	}
	return n, nil
}

// WriteHeader writes the http.Header to the underlying http.ResponseWriter.
func (w *Writer) WriteHeader(statusCode int) {
	contentType := w.Header().Get("Content-Type")
	isHTML := strings.Contains(contentType, "text/html")
	isJSON := strings.Contains(contentType, "application/json")
	if statusCode < 300 && (isHTML || isJSON) {
		w.w.Header().Set("Content-Encoding", "gzip")
		w.w.Header().Del("Content-Length")
		w.Compressible = true
	}
	w.w.WriteHeader(statusCode)
}

// Close closes the underlying gzip.Writer.
func (w *Writer) Close() error {
	if w.zw != nil {
		if err := w.zw.Close(); err != nil {
			return fmt.Errorf("compress[Writer].Close: %w", err)
		}
	}
	return nil
}

// Reader is a io.Reader that can be used to decompress data.
type Reader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewCompressReader returns a new Reader that can be used to decompress data.
func NewCompressReader(r io.ReadCloser) (*Reader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("NewCompressReader: %w", err)
	}
	return &Reader{
		r:  r,
		zr: zr,
	}, nil
}

// Read reads data from the underlying gzip.Reader.
func (r *Reader) Read(p []byte) (int, error) {
	n, err := r.zr.Read(p)
	if err != nil && !errors.Is(err, io.EOF) {
		err = fmt.Errorf("compress[Reader].Read: %w", err)
	}
	return n, err
}

// Close closes the underlying gzip.Reader.
func (r *Reader) Close() error {
	if err := r.r.Close(); err != nil {
		return fmt.Errorf("compress[Reader].Close: %w", err)
	}
	if err := r.zr.Close(); err != nil {
		return fmt.Errorf("compress[Reader].Close: %w", err)
	}
	return nil
}
