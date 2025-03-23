package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

//go:generate mockgen -source=updatejson.go -destination=mocks/mock_updater.go -package=handlers
type Updater interface {
	Update(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error)
}

func NewUpdateJSON(s Updater) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		mr, err := model.UnmarshalMetricRequestFromReader(r.Body)
		if err == nil {
			err = mr.RequiredValue()
		}
		if err != nil {
			requestCtxWithLogMessageFromError(r, fmt.Errorf("failed unmarshal: %w", err))
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
			requestCtxWithLogMessageFromError(r, fmt.Errorf("failed update: %w", err))
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
