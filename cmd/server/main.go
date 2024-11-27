package main

import (
	"github.com/korobkovandrey/runtime-metrics/internal/server/controller"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
	"log"
	"net/http"
)

const (
	updateRoutePath = `/update/`
)

func main() {
	mux := http.NewServeMux()
	store := repository.NewStoreMemStorage()

	mux.Handle(updateRoutePath,
		http.StripPrefix(updateRoutePath, middleware.BadRequestIfMethodNotEqualPOST(
			http.HandlerFunc(controller.UpdateHandler(store)),
		)),
	)

	if err := http.ListenAndServe(`localhost:8080`, mux); err != nil {
		log.Fatal(err)
	}
}
