package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

//go:generate mockgen -source=updates.go -destination=mocks/mock_batchupdater.go -package=mocks
type BatchUpdater interface {
	UpdateBatch(context.Context, []*model.MetricRequest) ([]*model.Metric, error)
}

func NewUpdatesHandler(s BatchUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mrs, err := model.UnmarshalMetricsRequestFromReader(r.Body)
		if err != nil {
			RequestCtxWithLogMessageFromError(r, fmt.Errorf("failed to unmarshal metrics request: %w", err))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if err = model.ValidateMetricsRequest(mrs); err != nil {
			RequestCtxWithLogMessageFromError(r, fmt.Errorf("failed to validate metrics request: %w", err))
			errMsg := http.StatusText(http.StatusBadRequest)
			switch {
			case errors.Is(err, model.ErrTypeIsNotValid):
				errMsg += ": " + model.ErrTypeIsNotValid.Error()
			case errors.Is(err, model.ErrValueIsNotValid):
				errMsg += ": " + model.ErrValueIsNotValid.Error()
			}
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}
		ms, err := s.UpdateBatch(r.Context(), mrs)
		if err != nil {
			RequestCtxWithLogMessageFromError(r, fmt.Errorf("failed to update metric batch: %w", err))
			if errors.Is(err, model.ErrMetricNotFound) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		responseMarshaled(ms, w, r)
	}
}
