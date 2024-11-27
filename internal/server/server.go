package server

import (
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/server/controller"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"

	"net/http"
)

const (
	updateRoutePath = `/update/`
)

type Server struct{}

func New() *Server {
	return &Server{}
}

func (s Server) Run() error {
	mux := http.NewServeMux()
	store := repository.NewStoreMemStorage()

	mux.Handle(updateRoutePath,
		http.StripPrefix(updateRoutePath, middleware.BadRequestIfMethodNotEqualPOST(
			http.HandlerFunc(controller.UpdateHandler(store)),
		)),
	)

	if err := http.ListenAndServe(`localhost:8080`, mux); err != nil {
		return fmt.Errorf(`server.Run: %w`, err)
	}
	return nil
}
