package server

import (
	"fmt"

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
	mux := http.NewServeMux()
	store := repository.NewStoreMemStorage()

	updateBasePattern := http.MethodPost + ` ` + s.config.UpdatePath
	mux.Handle(updateBasePattern+`/`, http.HandlerFunc(controller.UpdateHandlerFunc(store)))
	mux.Handle(updateBasePattern+`/{type}`, http.HandlerFunc(controller.UpdateHandlerFunc(store)))
	mux.Handle(updateBasePattern+`/{type}/`, http.HandlerFunc(controller.UpdateHandlerFunc(store)))
	mux.Handle(updateBasePattern+`/{type}/{name}`, http.HandlerFunc(controller.UpdateHandlerFunc(store)))
	mux.Handle(updateBasePattern+`/{type}/{name}/`, http.HandlerFunc(controller.UpdateHandlerFunc(store)))
	mux.Handle(updateBasePattern+`/{type}/{name}/{value}`, http.HandlerFunc(controller.UpdateHandlerFunc(store)))

	return mux
}

func (s Server) Run() error {
	fmt.Printf("Server listen: %s\n", `http://`+s.config.Addr+`/`)
	if err := http.ListenAndServe(s.config.Addr, s.NewHandler()); err != nil {
		return fmt.Errorf(`server.Run: %w`, err)
	}
	return nil
}
