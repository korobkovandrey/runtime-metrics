package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

func NewValueJSONHandler(s Finder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mr, err := model.UnmarshalMetricRequestFromReader(r.Body)
		if err == nil {
			err = mr.ValidateType()
		}
		if err != nil {
			RequestCtxWithLogMessageFromError(r, fmt.Errorf("failed to unmarshal metric request: %w", err))
			errMsg := http.StatusText(http.StatusBadRequest)
			if errors.Is(err, model.ErrMetricNotFound) {
				errMsg += ": " + model.ErrMetricNotFound.Error()
			} else if errors.Is(err, model.ErrTypeIsNotValid) {
				errMsg += ": " + model.ErrTypeIsNotValid.Error()
			}
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}
		m, err := s.Find(r.Context(), mr)
		if err != nil {
			RequestCtxWithLogMessageFromError(r, fmt.Errorf("failed to find metric: %w", err))
			if errors.Is(err, model.ErrMetricNotFound) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		responseMarshaled(m, w, r)
	}
}
