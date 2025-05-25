package handlers

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

// AllFinder is an interface for finding all metrics
//
//go:generate mockgen -source=index.go -destination=mocks/mock_allfinder.go -package=mocks
type AllFinder interface {
	FindAll(context.Context) ([]*model.Metric, error)
}

// NewIndexHandler creates a new index handler
func NewIndexHandler(s AllFinder) (http.HandlerFunc, error) {
	tpl, err := template.ParseFiles("./web/template/index.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data, err := s.FindAll(r.Context())
		if err != nil {
			RequestCtxWithLogMessageFromError(r, fmt.Errorf("failed to find all: %w", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		if err = tpl.Execute(w, data); err != nil {
			RequestCtxWithLogMessageFromError(r, fmt.Errorf("failed to execute template: %w", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}, nil
}
