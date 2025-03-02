package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

func (c *Controller) valueURI(w http.ResponseWriter, r *http.Request) {
	t := r.PathValue("type")
	name := r.PathValue("name")
	value := "0"
	mr, err := model.NewMetricRequest(t, name, value)
	if err != nil {
		c.requestCtxWithLogMessageFromError(r, fmt.Errorf("controller.valueURI: %w", err))
		errMsg := http.StatusText(http.StatusBadRequest)
		if errors.Is(err, model.ErrTypeIsNotValid) {
			errMsg += ": " + model.ErrTypeIsNotValid.Error()
		}
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	m, err := c.s.Find(mr)
	if err != nil {
		c.requestCtxWithLogMessageFromError(r, fmt.Errorf("controller.valueURI: %w", err))
		if errors.Is(err, model.ErrMetricNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	_, err = fmt.Fprint(w, m.AnyValue())
	if err != nil {
		c.requestCtxWithLogMessageFromError(r, fmt.Errorf("controller.valueURI: %w", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (c *Controller) valueJSON(w http.ResponseWriter, r *http.Request) {
	mr, err := model.UnmarshalMetricRequestFromReader(r.Body)
	if err == nil {
		err = mr.ValidateType()
	}
	if err != nil {
		c.requestCtxWithLogMessageFromError(r, fmt.Errorf("controller.valueJSON: %w", err))
		errMsg := http.StatusText(http.StatusBadRequest)
		if errors.Is(err, model.ErrMetricNotFound) {
			errMsg += ": " + model.ErrMetricNotFound.Error()
		} else if errors.Is(err, model.ErrTypeIsNotValid) {
			errMsg += ": " + model.ErrTypeIsNotValid.Error()
		}
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	m, err := c.s.Find(mr)
	if err != nil {
		c.requestCtxWithLogMessageFromError(r, fmt.Errorf("controller.valueJSON: %w", err))
		if errors.Is(err, model.ErrMetricNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	c.responseMarshaled(m, w, r)
}
