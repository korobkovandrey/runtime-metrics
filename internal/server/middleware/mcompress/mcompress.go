// Package mcompress provides a middleware for compressing responses and decompressing requests.
package mcompress

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/korobkovandrey/runtime-metrics/pkg/compress"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
)

// GzipCompressed returns a middleware that compresses responses and decompresses requests.
func GzipCompressed(l *logging.ZapLogger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ow := w
			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				cw := compress.NewCompressWriter(w)
				ow = cw
				defer func(cw *compress.Writer) {
					if err := cw.Close(); err != nil {
						l.ErrorCtx(r.Context(), fmt.Errorf("failed to close compress writer: %w", err).Error())
					}
				}(cw)
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := compress.NewCompressReader(r.Body)
				if err != nil {
					l.ErrorCtx(r.Context(), fmt.Errorf("failed to create compress reader: %w", err).Error())
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer func(cr *compress.Reader) {
					if err := cr.Close(); err != nil {
						l.ErrorCtx(r.Context(), fmt.Errorf("failed to close compress reader: %w", err).Error())
					}
				}(cr)
			}
			h.ServeHTTP(ow, r)
		})
	}
}
