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
		if t == "" {
			http.Error(w, "Type is required.", http.StatusBadRequest)
			return
		}
		if name == "" {
			http.NotFound(w, r)
			return
		}
		if value == "" {
			http.Error(w, "Value is required.", http.StatusBadRequest)
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

		if err = m.Update(name, value); err != nil {
			log.Println(r.URL.Path, fmt.Errorf("m.Update(%s, %s): %w", name, value, err))
			http.Error(w, "bad request: invalid number", http.StatusBadRequest)
			return
		}
	}
}
