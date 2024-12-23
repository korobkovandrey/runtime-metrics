package controller

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
)

func UpdateHandlerFunc(store *repository.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		t := r.PathValue("type")
		name := r.PathValue("name")
		value := r.PathValue("value")
		if t == "" {
			http.Error(w, strTypeIsRequired, http.StatusBadRequest)
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
			log.Println(r.URL.Path, fmt.Errorf("UpdateHandlerFunc store.Get(%v): %w", t, err))
			if errors.Is(err, repository.ErrTypeIsNotValid) {
				http.Error(w, fmt.Errorf(http.StatusText(http.StatusBadRequest)+": %w", err).Error(), http.StatusBadRequest)
				return
			}
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if err = m.Update(name, value); err != nil {
			log.Println(r.URL.Path, fmt.Errorf("UpdateHandlerFunc m.Update(%s, %s): %w", name, value, err))
			http.Error(w, http.StatusText(http.StatusBadRequest)+": invalid number", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
