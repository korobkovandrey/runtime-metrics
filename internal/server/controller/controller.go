package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"
	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/mcompress"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/mlogger"
	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"go.uber.org/zap"
)

type Service interface {
	Update(mr *model.MetricRequest) (*model.Metric, error)
	Find(mr *model.MetricRequest) (*model.Metric, error)
	FindAll() ([]*model.Metric, error)
}

type Controller struct {
	cfg *config.Config
	s   Service
	l   *logging.ZapLogger
	r   chi.Router
}

func NewController(cfg *config.Config, service Service, logger *logging.ZapLogger) *Controller {
	return &Controller{
		cfg: cfg,
		s:   service,
		l:   logger,
		r:   chi.NewRouter(),
	}
}

func (c *Controller) routes() error {
	c.r.Use(mcompress.GzipCompressed(c.l), mlogger.RequestLogger(c.l))
	c.r.Route("/update", func(r chi.Router) {
		r.Post("/", c.updateJSON)
		r.Route("/{type}", func(r chi.Router) {
			r.Post("/", http.NotFound)
			r.Route("/{name}", func(r chi.Router) {
				r.Post("/", func(w http.ResponseWriter, r *http.Request) {
					c.l.RequestWithContextFields(r, zap.String("errorText", "Value is required."))
					http.Error(w, "Value is required.", http.StatusBadRequest)
				})
				r.Post("/{value}", c.updateURI)
			})
		})
	})
	c.r.Route("/value", func(r chi.Router) {
		r.Post("/", c.valueJSON)
		r.Get("/{type}/{name}", c.valueURI)
	})
	indexFunc, err := c.indexFunc()
	if err != nil {
		return fmt.Errorf("controller.routes: %w", err)
	}
	c.r.Get("/", indexFunc)
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
		c.l.RequestWithContextFields(r, zap.Error(fmt.Errorf("controller.responseMarshaled: %w", err)))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func (c *Controller) ServeHTTP(ctx context.Context) error {
	err := c.routes()
	if err != nil {
		return fmt.Errorf("controller.ServeHTTP: %w", err)
	}
	c.l.InfoCtx(ctx, "Server started on http://"+c.cfg.Addr+"/")
	server := http.Server{
		Addr:     c.cfg.Addr,
		ErrorLog: c.l.Std(),
		Handler:  c.r,
	}
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		<-stop
		c.l.InfoCtx(ctx, "Shutting down the HTTP server...")
		if err := server.Shutdown(ctx); err != nil {
			c.l.ErrorCtx(ctx, "controller.ServeHTTP", zap.Error(err))
		}
	}()

	err = server.ListenAndServe()
	if err != nil {
		return fmt.Errorf("controller.ServeHTTP: %w", err)
	}
	return nil
}
