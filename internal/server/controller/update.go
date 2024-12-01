package controller

import (
	"errors"
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"

	"log"
	"net/http"
)

func UpdateHandlerFunc(store *repository.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t := r.PathValue("type")
		name := r.PathValue("name")
		value := r.PathValue("value")
		if t == `` {
			http.Error(w, `Type is required.`, http.StatusBadRequest)
			return
		}
		if name == `` {
			http.NotFound(w, r)
			return
		}
		if value == `` {
			http.Error(w, `Value is required.`, http.StatusBadRequest)
			return
		}
		m, err := store.Get(t)
		if err != nil {
			log.Println(r.URL.Path, fmt.Errorf(`store.Get(%v): %w`, t, err))
			responseText := ``
			if errors.Is(err, repository.ErrTypeIsNotValid) {
				responseText = fmt.Errorf(`bad request: %w`, err).Error()
			}
			http.Error(w, responseText, http.StatusBadRequest)
			return
		}

		if err = m.Update(name, value); err != nil {
			log.Println(r.URL.Path, fmt.Errorf(`m.Update(%s, %s): %w`, name, value, err))
			http.Error(w, `bad request: invalid number`, http.StatusBadRequest)
			return
		}

		if v, ok := m.GetStorageValue(name); ok {
			log.Printf(`%s OK %s: %s[%s] = %v`, r.URL.Path, value, t, name, v)
		} else {
			log.Printf(`%s FAIL %s: %s[%s] not found in storage`, r.URL.Path, value, t, name)
		}
	}
}
