package controller

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
)

func UpdateJSONHandlerFunc(store *repository.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metric, ok, err := readMetricFromRequest(w, r)
		if err != nil {
			log.Println(r.URL.Path, fmt.Errorf("UpdateHandlerFunc: %w", err))
		}
		if !ok {
			return
		}

		err = store.UpdateMetric(&metric)
		if err != nil {
			log.Println(r.URL.Path, fmt.Errorf("UpdateJSONHandlerFunc: %w", err))
			if errors.Is(err, repository.ErrTypeIsNotValid) || errors.Is(err, repository.ErrValueIsRequired) {
				http.Error(w, fmt.Errorf(
					http.StatusText(http.StatusBadRequest)+": %w", errors.Unwrap(err)).Error(),
					http.StatusBadRequest)
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
