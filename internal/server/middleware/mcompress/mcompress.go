package mcompress

import (
	"net/http"
	"strings"

	"github.com/korobkovandrey/runtime-metrics/pkg/compress"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"
)

func GzipCompressed(logger *logging.ZapLogger) func(h http.Handler) http.Handler {
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
						logger.ErrorCtx(r.Context(), "GzipCompressed", zap.Error(err))
					}
				}(cw)
			}

			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				cr, err := compress.NewCompressReader(r.Body)
				if err != nil {
					logger.ErrorCtx(r.Context(), "GzipCompressed", zap.Error(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer func(cr *compress.Reader) {
					if err := cr.Close(); err != nil {
						logger.ErrorCtx(r.Context(), "GzipCompressed", zap.Error(err))
					}
				}(cr)
			}
			h.ServeHTTP(ow, r)
		})
	}
}
