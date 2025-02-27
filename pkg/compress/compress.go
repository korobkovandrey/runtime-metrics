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
		Compressible: false,
	}
}

func (w *Writer) Header() http.Header {
	return w.w.Header()
}

func (w *Writer) Write(p []byte) (int, error) {
	if w.Compressible {
		if w.zw == nil {
			w.zw = gzip.NewWriter(w.w)
		} else {
			w.zw.Reset(w.w)
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

func (w *Writer) Close() error {
	if w.zw != nil {
		if err := w.zw.Close(); err != nil {
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

func (r *Reader) Read(p []byte) (int, error) {
	n, err := r.zr.Read(p)
	if !errors.Is(err, io.EOF) {
		err = fmt.Errorf("compress[Reader].Read: %w", err)
	}
	return n, err
}

func (r *Reader) Close() error {
	if err := r.r.Close(); err != nil {
		return fmt.Errorf("compress[Reader].Close: %w", err)
	}
	if err := r.zr.Close(); err != nil {
		return fmt.Errorf("compress[Reader].Close: %w", err)
	}
	return nil
}
