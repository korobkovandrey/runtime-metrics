package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

// Finder is a finder for metrics.
//
//go:generate mockgen -source=valueuri.go -destination=mocks/mock_finder.go -package=mocks
type Finder interface {
	Find(context.Context, *model.MetricRequest) (*model.Metric, error)
}

// NewValueURIHandler returns a handler for the value URI.
func NewValueURIHandler(s Finder) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := r.PathValue("type")
		name := r.PathValue("name")
		value := "0"
		mr, err := model.NewMetricRequest(t, name, value)
		if err != nil {
			RequestCtxWithLogMessageFromError(r, fmt.Errorf("failed to create metric request: %w", err))
			errMsg := http.StatusText(http.StatusBadRequest)
			if errors.Is(err, model.ErrTypeIsNotValid) {
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
		_, err = fmt.Fprint(w, m.AnyValue())
		if err != nil {
			RequestCtxWithLogMessageFromError(r, fmt.Errorf("failed to write value: %w", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}
