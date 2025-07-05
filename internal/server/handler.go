// Package server provides a handler for the HTTP server.
package server

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/factory"
	"github.com/korobkovandrey/runtime-metrics/internal/server/handlers"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/mcompress"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/mcrypto"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/mlogger"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/msign"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/msubnet"
	"github.com/korobkovandrey/runtime-metrics/internal/server/service"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
)

// Handler is a handler for the HTTP server.
type Handler struct {
	chi.Router
}

// NewHandler returns a new Handler.
func NewHandler() *Handler {
	return &Handler{Router: chi.NewRouter()}
}

// Configure configures the handler.
func (h *Handler) Configure(cfg *config.Config, r factory.Repository, l *logging.ZapLogger) error {
	h.Use(
		mlogger.RequestLogger(l),
		msubnet.Middleware(cfg.IPNet),
		mcompress.GzipCompressed(l),
		mcrypto.Middleware(cfg.PrivateKey, "/updates/"),
		msign.Signer([]byte(cfg.Key)),
	)
	if cfg.Pprof {
		h.Mount("/debug", middleware.Profiler())
	}
	if rPinger, ok := r.(handlers.Pinger); ok {
		h.setPingRoute(rPinger)
	} else {
		h.setPingRoute(nil)
	}

	finder := service.NewFinder(r)
	if err := h.setIndexRoute(finder); err != nil {
		return fmt.Errorf("failed to set index route: %w", err)
	}
	h.setUpdateRoutes(service.NewUpdater(r))
	h.setUpdatesRoute(service.NewBatchUpdater(r))
	h.setValueRoutes(finder)
	return nil
}

// setIndexRoute sets the index route.
func (h *Handler) setIndexRoute(s handlers.AllFinder) error {
	indexHandler, err := handlers.NewIndexHandler(s)
	if err != nil {
		return fmt.Errorf("failed to create index handler: %w", err)
	}
	h.Get("/", indexHandler)
	return nil
}

// setPingRoute sets the ping route.
func (h *Handler) setPingRoute(s handlers.Pinger) {
	var pingHandler http.HandlerFunc
	if s == nil {
		pingHandler = func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}
	} else {
		pingHandler = handlers.NewPingHandler(s)
	}
	h.Get("/ping", pingHandler)
}

// setUpdateRoutes sets the update routes.
func (h *Handler) setUpdateRoutes(s handlers.Updater) {
	h.Route("/update", func(r chi.Router) {
		r.Post("/", handlers.NewUpdateJSONHandler(s))
		r.Route("/{type}", func(r chi.Router) {
			r.Post("/", http.NotFound)
			r.Route("/{name}", func(r chi.Router) {
				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					handlers.RequestCtxWithLogMessage(r, "Value is required.")
					http.Error(w, "Value is required.", http.StatusBadRequest)
				})
				r.Post("/{value}", handlers.NewUpdateURIHandler(s))
			})
		})
	})
}

// setUpdatesRoute sets the updates route.
func (h *Handler) setUpdatesRoute(s handlers.BatchUpdater) {
	h.Post("/updates/", handlers.NewUpdatesHandler(s))
}

// setValueRoutes sets the value routes.
func (h *Handler) setValueRoutes(s handlers.Finder) {
	h.Route("/value", func(r chi.Router) {
		r.Post("/", handlers.NewValueJSONHandler(s))
		r.Get("/{type}/{name}", handlers.NewValueURIHandler(s))
	})
}
