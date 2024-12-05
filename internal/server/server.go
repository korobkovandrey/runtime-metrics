package server

import (
	"fmt"

	"github.com/go-chi/chi/v5"

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

func (s Server) NewHandler() http.Handler {
	store := repository.NewStoreMemStorage()
	r := chi.NewRouter()

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
	r.Get("/", controller.IndexHandlerFunc(store))
	return r
}

func (s Server) Run() error {
	fmt.Printf("Server listen: %s\n", "http://"+s.config.Addr+"/")
	if err := http.ListenAndServe(s.config.Addr, s.NewHandler()); err != nil {
		return fmt.Errorf("server.Run: %w", err)
	}
	return nil
}
