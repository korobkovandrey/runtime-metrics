package server

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/korobkovandrey/runtime-metrics/internal/server/logger"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware"

	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/controller"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"

	"net/http"
)

type Server struct {
	config *config.Config
}

func New(cfg *config.Config) *Server {
	return &Server{cfg}
}

func (s Server) NewHandler() (http.Handler, error) {
	store := repository.NewStoreMemStorage()
	r := chi.NewRouter()
	r.Use(middleware.SugarRequestLogger(logger.Sugar()))

	updateHandlerFunc := controller.UpdateHandlerFunc(store)
	r.Route("/update", func(r chi.Router) {
		r.Post("/", updateHandlerFunc)
		r.Route("/{type}", func(r chi.Router) {
			r.Post("/", updateHandlerFunc)
			r.Route("/{name}", func(r chi.Router) {
				r.Post("/", updateHandlerFunc)
				r.Post("/{value}", updateHandlerFunc)
			})
		})
	})
	r.Get("/value/{type}/{name}", controller.ValueHandlerFunc(store))

	indexHandlerFunc, err := controller.IndexHandlerFunc(store)
	if err != nil {
		return r, fmt.Errorf("NewHandler: %w", err)
	}
	r.Get("/", indexHandlerFunc)
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
