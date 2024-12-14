package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
)

var (
	errEmptyError = errors.New("empty error")
)

func readMetricsFromRequest(w http.ResponseWriter, r *http.Request) (metrics model.Metrics, err error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, strTypeIsRequired, http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		http.Error(w, strTypeIsRequired, http.StatusBadRequest)
		err = errEmptyError
		return
	}
	err = json.Unmarshal(body, &metrics)
	if err != nil {
		http.Error(w, strTypeIsRequired, http.StatusBadRequest)
		return
	}
	if len(metrics.MType) == 0 {
		http.Error(w, strTypeIsRequired, http.StatusBadRequest)
		err = errEmptyError
		return
	}
	if len(metrics.ID) == 0 {
		http.Error(w, "ID is required.", http.StatusBadRequest)
		err = errEmptyError
		return
	}
	return
}

func responseMetricsJSON(metrics *model.Metrics, w http.ResponseWriter) (err error) {
	response, err := json.Marshal(metrics)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		return
	}
	return
}

func UpdateJSONHandlerFunc(store *repository.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		metrics, err := readMetricsFromRequest(w, r)
		if err != nil {
			if !errors.Is(err, errEmptyError) {
				log.Println(r.URL.Path, fmt.Errorf("UpdateHandlerFunc: %w", err))
			}
			return
		}

		err = store.UpdateMetrics(&metrics)
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
		err = responseMetricsJSON(&metrics, w)
		if err != nil {
			log.Println(r.URL.Path, fmt.Errorf("UpdateJSONHandlerFunc: %w", err))
			return
		}
	}
}
