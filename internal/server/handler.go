package server

import (
	"github.com/korobkovandrey/runtime-metrics/internal/storage"
	"net/http"
	"strings"
)

func Handler(metric storage.Metric) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		raw := strings.Split(
			strings.Trim(
				r.URL.Path,
				`/`,
			),
			`/`,
		)
		lenRaw := len(raw)
		if lenRaw == 0 || len(raw[0]) == 0 {
			http.NotFound(w, r)
			return
		}
		if lenRaw != 2 || len(raw[1]) == 0 {
			http.Error(w, `Wrong arguments`, http.StatusBadRequest)
			return
		}

		err := metric.Handler(raw[0], raw[1])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}
