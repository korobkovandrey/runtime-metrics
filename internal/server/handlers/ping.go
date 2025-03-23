package handlers

import (
	"database/sql/driver"
	"fmt"
	"net/http"
)

//go:generate mockgen -source=ping.go -destination=mock_pinger.go -package=handlers
type Pinger interface {
	driver.Pinger
}

func NewPing(s Pinger) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(s)
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
