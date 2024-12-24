package mcompress

import (
	"log"
	"net/http"
	"strings"

	"github.com/korobkovandrey/runtime-metrics/internal/server/compress"
)

func GzipCompressed(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			cw := compress.NewCompressWriter(w)
			ow = cw
			defer func(cw *compress.Writer) {
				if cw.Compressible {
					if err := cw.Close(); err != nil {
						log.Println(err)
					}
				}
			}(cw)
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			cr, err := compress.NewCompressReader(r.Body)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer func(cr *compress.Reader) {
				if err := cr.Close(); err != nil {
					log.Println(err)
				}
			}(cr)
		}
		h.ServeHTTP(ow, r)
	})
}
