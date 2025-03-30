package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/handlers"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/mcompress"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/mlogger"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/msign"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository/pgxstorage"
	"github.com/korobkovandrey/runtime-metrics/internal/server/service"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
)

type Handler struct {
	chi.Router
	closers []func() error
}

func NewHandler() *Handler {
	return &Handler{Router: chi.NewRouter()}
}

func (h *Handler) Configure(ctx context.Context, cfg *config.Config, l *logging.ZapLogger) error {
	h.Use(mcompress.GzipCompressed(l), msign.Signer([]byte(cfg.Key)), mlogger.RequestLogger(l))
	var r interface {
		service.FinderRepository
		service.UpdaterRepository
		service.BatchUpdaterRepository
	}
	if cfg.DatabaseDSN != "" {
		ps, err := pgxstorage.NewPGXStorage(ctx, &pgxstorage.Config{
			DSN:         cfg.DatabaseDSN,
			PingTimeout: cfg.DatabasePingTimeout,
			RetryDelays: cfg.RetryDelays,
		})
		if err != nil {
			return fmt.Errorf("failed to create pgxstorage: %w", err)
		}
		h.closers = append(h.closers, ps.Close)
		h.setPingRoute(ps)
		r = ps
	} else {
		ms := repository.NewMemStorage()
		if cfg.FileStoragePath != "" {
			fs := repository.NewFileStorage(ms, cfg)
			if cfg.Restore {
				if err := fs.Restore(); err != nil {
					return fmt.Errorf("failed to restore: %w", err)
				}
			}
			h.closers = append(h.closers, fs.Close)
			go fs.Run(ctx, l)
			r = fs
		} else {
			r = ms
		}
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

func (h *Handler) Close() error {
	var errs []error
	for i := range h.closers {
		if err := h.closers[i](); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (h *Handler) setIndexRoute(s handlers.AllFinder) error {
	indexHandler, err := handlers.NewIndexHandler(s)
	if err != nil {
		return fmt.Errorf("failed to create index handler: %w", err)
	}
	h.Get("/", indexHandler)
	return nil
}

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

func (h *Handler) setUpdatesRoute(s handlers.BatchUpdater) {
	h.Post("/updates/", handlers.NewUpdatesHandler(s))
}

func (h *Handler) setValueRoutes(s handlers.Finder) *Handler {
	h.Route("/value", func(r chi.Router) {
		r.Post("/", handlers.NewValueJSONHandler(s))
		r.Get("/{type}/{name}", handlers.NewValueURIHandler(s))
	})
	return h
}
