package server

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/mcompress"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/mlogger"
	"go.uber.org/zap"

	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/controller"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"

	"net/http"
)

type Server struct {
	config *config.Config
	logger *zap.Logger
}

func New(cfg *config.Config, logger *zap.Logger) *Server {
	return &Server{cfg, logger}
}

func (s Server) NewHandler() (http.Handler, error) {
	store := repository.NewStoreMemStorage()
	r := chi.NewRouter()
	sugaredLogger := s.logger.Sugar().Named("request")
	updateHandlerFunc := controller.UpdateHandlerFunc(store)

	r.With(mlogger.SugarRequestLogger(sugaredLogger.Named("update")), mcompress.GzipCompressed).
		Route("/update", func(r chi.Router) {
			r.Post("/", controller.UpdateJSONHandlerFunc(store))
			r.Route("/{type}", func(r chi.Router) {
				r.Post("/", updateHandlerFunc)
				r.Route("/{name}", func(r chi.Router) {
					r.Post("/", updateHandlerFunc)
					r.Post("/{value}", updateHandlerFunc)
				})
			})
		})

	r.With(mlogger.SugarRequestLogger(sugaredLogger.Named("value")), mcompress.GzipCompressed).
		Route("/value", func(r chi.Router) {
			r.Post("/", controller.ValueJSONHandlerFunc(store))
			r.Get("/{type}/{name}", controller.ValueHandlerFunc(store))
		})

	indexHandlerFunc, err := controller.IndexHandlerFunc(store)
	if err != nil {
		return r, fmt.Errorf("NewHandler: %w", err)
	}
	r.With(mlogger.SugarRequestLogger(sugaredLogger.Named("index")), mcompress.GzipCompressed).
		Get("/", indexHandlerFunc)
	return r, nil
}

func (s Server) Run() error {
	handler, err := s.NewHandler()
	if err != nil {
		return fmt.Errorf("server.NewHandler: %w", err)
	}
	fmt.Printf("Server listen: %s\n", "http://"+s.config.Addr+"/")
	if err = http.ListenAndServe(s.config.Addr, handler); err != nil {
		return fmt.Errorf("server.Run: %w", err)
	}
	return nil
}
