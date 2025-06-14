package handlers

import (
	"database/sql/driver"
	"fmt"
	"net/http"
)

// Pinger is an interface for database/sql.Pinger
//
//go:generate mockgen -source=ping.go -destination=mocks/mock_pinger.go -package=mocks
type Pinger interface {
	driver.Pinger
}

// NewPingHandler creates a new ping handler
func NewPingHandler(s Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.Ping(r.Context()); err != nil {
			RequestCtxWithLogMessageFromError(r, fmt.Errorf("failed ping: %w", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
