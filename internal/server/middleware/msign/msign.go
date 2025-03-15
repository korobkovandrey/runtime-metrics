package msign

import (
	"bytes"
	"errors"
	"io"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/pkg/sign"
)

type errorReadCloser struct {
	io.ReadCloser
	err error
}

func newErrorReadCloser(r io.ReadCloser, err error) *errorReadCloser {
	return &errorReadCloser{r, err}
}
func (r *errorReadCloser) Read(p []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}
	return r.ReadCloser.Read(p)
}

type signWriter struct {
	http.ResponseWriter
	buf        *bytes.Buffer
	statusCode int
	key        []byte
}

func newSignWriter(w http.ResponseWriter, key []byte) *signWriter {
	return &signWriter{
		ResponseWriter: w,
		buf:            &bytes.Buffer{},
		key:            key,
	}
}

func (w *signWriter) Write(data []byte) (n int, err error) {
	return w.buf.Write(data)
}

func (w *signWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func (w *signWriter) close() {
	data := w.buf.Bytes()
	if hash := sign.MakeToString(data, w.key); hash != "" {
		w.ResponseWriter.Header().Set("HashSHA256", hash)
	}
	w.ResponseWriter.WriteHeader(w.statusCode)
	_, _ = w.ResponseWriter.Write(data)
}

func Signer(key []byte) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(key) == 0 {
				h.ServeHTTP(w, r)
				return
			}
			var body []byte
			bh, err := sign.DecodeString(r.Header.Get("HashSHA256"))
			if err == nil {
				body, err = io.ReadAll(r.Body)
				if err == nil && !sign.Validate(body, key, bh) {
					err = errors.New("invalid signature")
				}
			}
			if err != nil {
				r.Body = newErrorReadCloser(r.Body, err)
			} else {
				r.Body = io.NopCloser(bytes.NewBuffer(body))
			}
			sw := newSignWriter(w, key)
			defer sw.close()
			h.ServeHTTP(sw, r)
		})
	}
}
