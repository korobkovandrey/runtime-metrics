package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

func (c *Controller) updatesJSON(w http.ResponseWriter, r *http.Request) {
	mrs, err := model.UnmarshalMetricsRequestFromReader(r.Body)
	if err != nil {
		c.requestCtxWithLogMessageFromError(r, fmt.Errorf("failed unmarshal: %w", err))
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
	for _, mr := range mrs {
		err := mr.RequiredValue()
		if err != nil {
			c.requestCtxWithLogMessageFromError(r, fmt.Errorf("failed validate: %w", err))
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
	}
	ms, err := c.s.UpdateBatch(r.Context(), mrs)
	if err != nil {
		c.requestCtxWithLogMessageFromError(r, fmt.Errorf("failed update: %w", err))
		if errors.Is(err, model.ErrMetricNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	c.responseMarshaled(ms, w, r)
}
