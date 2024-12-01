package server

import (
	"fmt"

	"github.com/go-chi/chi/v5"
	"github.com/korobkovandrey/runtime-metrics/internal/server/controller"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"

	"net/http"
)

type Config struct {
	Addr       string
	UpdatePath string
}

type Server struct {
	config Config
}

func New(config Config) *Server {
	return &Server{config}
}

func (s Server) NewHandler() http.Handler {
	store := repository.NewStoreMemStorage()
	r := chi.NewRouter()
	updateHandlerFunc := controller.UpdateHandlerFunc(store)
	r.Route(s.config.UpdatePath, func(r chi.Router) {
		r.Post("/", updateHandlerFunc)
		r.Route("/{type}", func(r chi.Router) {
			r.Post("/", updateHandlerFunc)
			r.Route("/{name}", func(r chi.Router) {
				r.Post("/", updateHandlerFunc)
				r.Post("/{value}", updateHandlerFunc)
			})
		})
	})
	return r
}

func (s Server) Run() error {
	fmt.Printf("Server listen: %s\n", `http://`+s.config.Addr+`/`)
	if err := http.ListenAndServe(s.config.Addr, s.NewHandler()); err != nil {
		return fmt.Errorf(`server.Run: %w`, err)
	}
	return nil
}
