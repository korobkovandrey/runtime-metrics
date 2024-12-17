package controller

import (
	"errors"
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"

	"log"
	"net/http"
)

// ValueHandlerFunc @todo test!!!
func ValueHandlerFunc(store *repository.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t := r.PathValue("type")
		name := r.PathValue("name")
		if t == "" {
			http.Error(w, "Type is required.", http.StatusBadRequest)
			return
		}
		if name == "" {
			http.NotFound(w, r)
			return
		}
		m, err := store.Get(t)
		if err != nil {
			log.Println(r.URL.Path, fmt.Errorf("store.Get(%v): %w", t, err))
			if errors.Is(err, repository.ErrTypeIsNotValid) {
				http.Error(w, fmt.Errorf("bad request: %w", err).Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		value, ok := m.GetStorageValue(name)
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, err = fmt.Fprint(w, value)
		if err != nil {
			log.Println(r.URL.Path, fmt.Errorf("fmt.Fprint(%v): %w", value, err))
			return
		}
	}
}
