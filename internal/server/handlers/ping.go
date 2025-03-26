package handlers

import (
	"database/sql/driver"
	"fmt"
	"net/http"
)

//go:generate mockgen -source=ping.go -destination=mocks/mock_pinger.go -package=mocks
type Pinger interface {
	driver.Pinger
}

func NewPingHandler(s Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s != nil {
			if err := s.Ping(r.Context()); err != nil {
				requestCtxWithLogMessageFromError(r, fmt.Errorf("controller.ping: %w", err))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}
