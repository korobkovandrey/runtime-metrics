package controller

import (
	"fmt"
	"net/http"
)

func (c *Controller) ping(w http.ResponseWriter, r *http.Request) {
	if c.pinger != nil {
		if err := c.pinger.Ping(r.Context()); err != nil {
			c.requestCtxWithLogMessageFromError(r, fmt.Errorf("controller.ping: %w", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}
