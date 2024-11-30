package controller

import (
	"errors"
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"

	"log"
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
			log.Println(r.URL.Path, fmt.Errorf(`store.Get(%v): %w`, raw[0], err))
			responseText := ``
			if errors.Is(err, repository.ErrTypeIsNotValid) {
				responseText = fmt.Errorf(`bad request: %w`, err).Error()
			}
			http.Error(w, responseText, http.StatusBadRequest)
			return
		}

		if err = m.Update(raw[1], raw[2]); err != nil {
			log.Println(r.URL.Path, fmt.Errorf(`m.Update(%s, %s): %w`, raw[1], raw[2], err))
			http.Error(w, `bad request: invalid number`, http.StatusBadRequest)
			return
		}

		if v, ok := m.GetStorageValue(raw[1]); ok {
			log.Printf(`%s OK %s: %s[%s] = %v`, r.URL.Path, raw[2], raw[0], raw[1], v)
		} else {
			log.Printf(`%s FAIL %s: %s[%s] not found in storage`, r.URL.Path, raw[2], raw[0], raw[1])
		}
	}
}
