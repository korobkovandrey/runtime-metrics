package controller

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/__old/server/repository"
	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"go.uber.org/zap"
)

func (c *Controller) updateURI(w http.ResponseWriter, r *http.Request) {
	t := r.PathValue("type")
	name := r.PathValue("name")
	value := r.PathValue("value")
	mr, err := model.NewMetricRequest(t, name, value)
	if err != nil {
		c.l.RequestWithContextFields(r, zap.Error(fmt.Errorf("controller.updateURI: %w", err)))
		errMsg := http.StatusText(http.StatusBadRequest)
		if errors.Is(err, model.ErrTypeIsNotValid) {
			errMsg += ": " + model.ErrTypeIsNotValid.Error()
		} else if errors.Is(err, model.ErrValueIsNotValid) {
			errMsg += ": " + model.ErrValueIsNotValid.Error()
		}
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	_, err = c.s.Update(mr)
	if err != nil {
		c.l.RequestWithContextFields(r, zap.Error(fmt.Errorf("controller.updateURI: %w", err)))
		if errors.Is(err, repository.ErrMetricNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

//nolint:dupl // ignore
func (c *Controller) updateJSON(w http.ResponseWriter, r *http.Request) {
	mr, err := model.NewMetricRequestFromReader(r.Body)
	if err != nil {
		c.l.RequestWithContextFields(r, zap.Error(fmt.Errorf("controller.updateJSON: %w", err)))
		errMsg := http.StatusText(http.StatusBadRequest)
		if errors.Is(err, model.ErrMetricNotFound) {
			errMsg += ": " + model.ErrMetricNotFound.Error()
		}
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}
	m, err := c.s.Update(mr)
	if err != nil {
		c.l.RequestWithContextFields(r, zap.Error(fmt.Errorf("controller.updateJSON: %w", err)))
		if errors.Is(err, repository.ErrMetricNotFound) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	c.responseMarshaled(m, w, r)
}
