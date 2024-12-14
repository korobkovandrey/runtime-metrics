package controller

import (
	"errors"
	"fmt"
	"log"

	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"

	"net/http"
)

func ValueJSONHandlerFunc(store *repository.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := readMetricsFromRequest(w, r)
		if err != nil {
			if !errors.Is(err, errEmptyError) {
				log.Println(r.URL.Path, fmt.Errorf("ValueJSONHandlerFunc: %w", err))
			}
			return
		}
		err = store.FillMetrics(&metrics)
		if err != nil {
			log.Println(r.URL.Path, fmt.Errorf("ValueJSONHandlerFunc: %w", err))
			if errors.Is(err, repository.ErrTypeIsNotValid) {
				http.Error(w,
					fmt.Errorf(http.StatusText(http.StatusBadRequest)+": %w", errors.Unwrap(err)).Error(),
					http.StatusBadRequest)
				return
			}
			if errors.Is(err, repository.ErrMetricsNotFound) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		err = responseMetricsJSON(&metrics, w)
		if err != nil {
			log.Println(r.URL.Path, fmt.Errorf("UpdateJSONHandlerFunc: %w", err))
			return
		}
	}
}
