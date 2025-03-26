package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

func NewUpdateURIHandler(s Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := r.PathValue("type")
		name := r.PathValue("name")
		value := r.PathValue("value")
		mr, err := model.NewMetricRequest(t, name, value)
		if err != nil {
			requestCtxWithLogMessageFromError(r, fmt.Errorf("controller.updateURI: %w", err))
			errMsg := http.StatusText(http.StatusBadRequest)
			if errors.Is(err, model.ErrTypeIsNotValid) {
				errMsg += ": " + model.ErrTypeIsNotValid.Error()
			} else if errors.Is(err, model.ErrValueIsNotValid) {
				errMsg += ": " + model.ErrValueIsNotValid.Error()
			}
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}
		_, err = s.Update(r.Context(), mr)
		if err != nil {
			requestCtxWithLogMessageFromError(r, fmt.Errorf("controller.updateURI: %w", err))
			if errors.Is(err, model.ErrMetricNotFound) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
