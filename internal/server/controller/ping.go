package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func (c *Controller) ping(w http.ResponseWriter, r *http.Request) {
	if c.db != nil {
		const pingTimeout = 5 * time.Second
		ctx, cancel := context.WithTimeout(r.Context(), pingTimeout)
		defer cancel()
		if err := c.db.Ping(ctx); err != nil {
			c.requestCtxWithLogMessageFromError(r, fmt.Errorf("controller.ping: %w", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
}
