package compress

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Writer struct {
	w            http.ResponseWriter
	zw           *gzip.Writer
	Compressible bool
}

func NewCompressWriter(w http.ResponseWriter) *Writer {
	return &Writer{
		w:            w,
		zw:           gzip.NewWriter(w),
		Compressible: false,
	}
}

func (c *Writer) Header() http.Header {
	return c.w.Header()
}

//nolint:wrapcheck // unnecessary
func (c *Writer) Write(p []byte) (int, error) {
	if c.Compressible {
		return c.zw.Write(p)
	}
	return c.w.Write(p)
}

func (c *Writer) WriteHeader(statusCode int) {
	contentType := c.Header().Get("Content-Type")
	isHTML := strings.Contains(contentType, "text/html")
	isJSON := strings.Contains(contentType, "application/json")
	if statusCode < 300 && (isHTML || isJSON) {
		c.w.Header().Set("Content-Encoding", "gzip")
		c.w.Header().Del("Content-Length")
		c.Compressible = true
	}
	c.w.WriteHeader(statusCode)
}

//nolint:wrapcheck // unnecessary
func (c *Writer) Close() error {
	return c.zw.Close()
}

type Reader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

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

//nolint:wrapcheck // unnecessary
func (c *Reader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

//nolint:wrapcheck // unnecessary
func (c *Reader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("compress.Close: %w", err)
	}
	return c.zr.Close()
}
