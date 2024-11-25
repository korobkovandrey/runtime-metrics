package controller

import (
	"fmt"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
	"net/http"
	"strings"
)

func UpdateHandler(store *repository.Store) func(w http.ResponseWriter, r *http.Request) {
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
			http.Error(w, `Type is required.`, http.StatusBadRequest)
			return
		}
		if lenRaw == 1 || len(raw[1]) == 0 {
			http.NotFound(w, r)
			return
		}
		if lenRaw == 2 || len(raw[2]) == 0 {
			http.Error(w, `Value is required.`, http.StatusBadRequest)
			return
		}
		m, err := store.Get(raw[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = m.Update(raw[1], raw[2]); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		fmt.Fprintf(w, `%v`, m.GetStorage())
	}
}
