package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/mcompress"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/mlogger"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"
)

//go:generate mockgen -source=controller.go -destination=../mocks/controller.go -package=mocks

type Service interface {
	Update(mr *model.MetricRequest) (*model.Metric, error)
	Find(mr *model.MetricRequest) (*model.Metric, error)
	FindAll() ([]*model.Metric, error)
	UpdateBatch(mrs []*model.MetricRequest) ([]*model.Metric, error)
}

type Pinger interface {
	repository.Pinger
}

type Controller struct {
	cfg    *config.Config
	s      Service
	pinger Pinger
	l      *logging.ZapLogger
	r      chi.Router
}

func NewController(cfg *config.Config, service Service, logger *logging.ZapLogger) *Controller {
	return &Controller{
		cfg: cfg,
		s:   service,
		l:   logger,
		r:   chi.NewRouter(),
	}
}

func (c *Controller) WithPinger(pinger Pinger) *Controller {
	c.pinger = pinger
	return c
}

func (c *Controller) routes() error {
	c.r.Use(mcompress.GzipCompressed(c.l), mlogger.RequestLogger(c.l))
	c.r.Route("/update", func(r chi.Router) {
		r.Post("/", c.updateJSON)
		r.Route("/{type}", func(r chi.Router) {
			r.Post("/", http.NotFound)
			r.Route("/{name}", func(r chi.Router) {
				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					c.requestCtxWithLogMessage(r, "Value is required.")
					http.Error(w, "Value is required.", http.StatusBadRequest)
				})
				r.Post("/{value}", c.updateURI)
			})
		})
	})
	c.r.Post("/updates/", c.updatesJSON)
	c.r.Route("/value", func(r chi.Router) {
		r.Post("/", c.valueJSON)
		r.Get("/{type}/{name}", c.valueURI)
	})
	c.r.Get("/ping", c.ping)
	indexFunc, err := c.indexFunc()
	if err != nil {
		return fmt.Errorf("controller.routes: %w", err)
	}
	c.r.Get("/", indexFunc)
	return nil
}

func (c *Controller) ListenAndServe(ctx context.Context) error {
	err := c.routes()
	if err != nil {
		return fmt.Errorf("controller.ListenAndServe: %w", err)
	}
	c.l.InfoCtx(ctx, "Server started on http://"+c.cfg.Addr+"/")
	server := http.Server{
		Addr:              c.cfg.Addr,
		ErrorLog:          c.l.Std(),
		Handler:           c.r,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func(ctx context.Context) {
		ctxWithoutCancel := context.WithoutCancel(ctx)
		<-ctx.Done()
		ctx, cancel := context.WithTimeout(ctxWithoutCancel, c.cfg.ShutdownTimeout)
		defer cancel()
		c.l.InfoCtx(ctx, "Shutting down the HTTP server...")
		if err := server.Shutdown(ctx); err != nil {
			c.l.ErrorCtx(ctx, "controller.ListenAndServe", zap.Error(err))
		}
	}(ctx)

	err = server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("controller.ListenAndServe: %w", err)
	}
	return nil
}

func (c *Controller) responseMarshaled(data any, w http.ResponseWriter, r *http.Request) {
	response, err := json.Marshal(data)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(response)
	}
	if err != nil {
		c.requestCtxWithLogMessageFromError(r, fmt.Errorf("failed response: %w", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

//nolint:godot // ignore
/* A func (c *Controller) requestCtxWithContextFields(r *http.Request, fields ...zap.Field) {
	*r = *r.WithContext(c.l.WithContextFields(r.Context(), fields...))
}*/

func (c *Controller) requestCtxWithLogMessage(r *http.Request, msg string) {
	*r = *r.WithContext(context.WithValue(r.Context(), mlogger.LogMessageKey, msg))
}

func (c *Controller) requestCtxWithLogMessageFromError(r *http.Request, err error) {
	c.requestCtxWithLogMessage(r, err.Error())
}
