package compress

import (
	"compress/gzip"
	"errors"
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

func (c *Writer) Write(p []byte) (int, error) {
	if c.Compressible {
		n, err := c.zw.Write(p)
		if err != nil {
			return n, fmt.Errorf("compress[Writer].Write: %w", err)
		}
		return n, nil
	}
	n, err := c.w.Write(p)
	if err != nil {
		return n, fmt.Errorf("compress[Writer].Write: %w", err)
	}
	return n, nil
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

func (c *Writer) Close() error {
	if c.Compressible {
		if err := c.zw.Close(); err != nil {
			return fmt.Errorf("compress[Writer].Close: %w", err)
		}
	}
	return nil
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

func (c *Reader) Read(p []byte) (int, error) {
	n, err := c.zr.Read(p)
	if !errors.Is(err, io.EOF) {
		err = fmt.Errorf("compress[Reader].Read: %w", err)
	}
	return n, err
}

func (c *Reader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("compress[Reader].Close: %w", err)
	}
	if err := c.zr.Close(); err != nil {
		return fmt.Errorf("compress[Reader].Close: %w", err)
	}
	return nil
}
