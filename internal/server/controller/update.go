package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

func (c *Controller) updateURI(w http.ResponseWriter, r *http.Request) {
	t := r.PathValue("type")
	name := r.PathValue("name")
	value := r.PathValue("value")
	mr, err := model.NewMetricRequest(t, name, value)
	if err != nil {
		c.requestCtxWithLogMessageFromError(r, fmt.Errorf("controller.updateURI: %w", err))
		errMsg := http.StatusText(http.StatusBadRequest)
		if errors.Is(err, model.ErrTypeIsNotValid) {
			errMsg += ": " + model.ErrTypeIsNotValid.Error()
		} else if errors.Is(err, model.ErrValueIsNotValid) {
			errMsg += ": " + model.ErrValueIsNotValid.Error()
		}
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	_, err = c.s.Update(r.Context(), mr)
	if err != nil {
		c.requestCtxWithLogMessageFromError(r, fmt.Errorf("controller.updateURI: %w", err))
		if errors.Is(err, model.ErrMetricNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (c *Controller) updateJSON(w http.ResponseWriter, r *http.Request) {
	mr, err := model.UnmarshalMetricRequestFromReader(r.Body)
	if err == nil {
		err = mr.RequiredValue()
	}
	if err != nil {
		c.requestCtxWithLogMessageFromError(r, fmt.Errorf("failed unmarshal: %w", err))
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
	m, err := c.s.Update(r.Context(), mr)
	if err != nil {
		c.requestCtxWithLogMessageFromError(r, fmt.Errorf("failed update: %w", err))
		if errors.Is(err, model.ErrMetricNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	c.responseMarshaled(m, w, r)
}
