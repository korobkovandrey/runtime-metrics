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
		metric, ok, err := readMetricFromRequest(w, r)
		if err != nil {
			log.Println(r.URL.Path, fmt.Errorf("ValueJSONHandlerFunc: %w", err))
		}
		if !ok {
			return
		}

		err = store.FillMetric(&metric)
		if err != nil {
			log.Println(r.URL.Path, fmt.Errorf("ValueJSONHandlerFunc: %w", err))
			if errors.Is(err, repository.ErrTypeIsNotValid) {
				http.Error(w,
					fmt.Errorf(http.StatusText(http.StatusBadRequest)+": %w", errors.Unwrap(err)).Error(),
					http.StatusBadRequest)
				return
			}
			if errors.Is(err, repository.ErrMetricNotFound) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		err = responseMetricJSON(&metric, w)
		if err != nil {
			log.Println(r.URL.Path, fmt.Errorf("UpdateJSONHandlerFunc: %w", err))
			return
		}
	}
}
