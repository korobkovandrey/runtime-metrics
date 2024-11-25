package main

import (
	"github.com/korobkovandrey/runtime-metrics/internal/server/controller"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
	"github.com/korobkovandrey/runtime-metrics/internal/storage/memstorage"
	"net/http"
)

const (
	updateRoutePath = `/update/`
)

func main() {
	mux := http.NewServeMux()
	memStorage := memstorage.NewMemStorage()
	store := repository.NewStore(
		repository.NewGauge(memStorage),
		repository.NewCounter(memStorage),
	)

	mux.Handle(updateRoutePath,
		http.StripPrefix(updateRoutePath, middleware.BadRequestIfMethodNotEqualPOST(
			http.HandlerFunc(controller.UpdateHandler(store)),
		)),
	)

	if err := http.ListenAndServe(`localhost:8080`, mux); err != nil {
		panic(err)
	}
}
