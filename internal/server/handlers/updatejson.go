package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

// Updater updates metrics
//
//go:generate mockgen -source=updatejson.go -destination=mocks/mock_updater.go -package=mocks
type Updater interface {
	Update(context.Context, *model.MetricRequest) (*model.Metric, error)
}

// NewUpdateJSONHandler returns a handler for updating metrics
func NewUpdateJSONHandler(s Updater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mr, err := model.UnmarshalMetricRequestFromReader(r.Body)
		if err == nil {
			err = mr.RequiredValue()
		}
		if err != nil {
			RequestCtxWithLogMessageFromError(r, fmt.Errorf("failed to unmarshal metric request: %w", err))
			errMsg := http.StatusText(http.StatusBadRequest)
			switch {
			case errors.Is(err, model.ErrMetricNotFound):
				errMsg += ": " + model.ErrMetricNotFound.Error()
			case errors.Is(err, model.ErrTypeIsNotValid):
				errMsg += ": " + model.ErrTypeIsNotValid.Error()
			case errors.Is(err, model.ErrValueIsNotValid):
				errMsg += ": " + model.ErrValueIsNotValid.Error()
			}
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}
		m, err := s.Update(r.Context(), mr)
		if err != nil {
			RequestCtxWithLogMessageFromError(r, fmt.Errorf("failed to update metric: %w", err))
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
