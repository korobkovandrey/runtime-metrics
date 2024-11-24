package main

import (
	"github.com/korobkovandrey/runtime-metrics/internal/server"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware"
	"github.com/korobkovandrey/runtime-metrics/internal/storage"
	"github.com/korobkovandrey/runtime-metrics/internal/storage/memstorage"
	"net/http"
)

const (
	routeGauge   = `/update/gauge/`
	routeCounter = `/update/counter/`
)

func main() {
	mux := http.NewServeMux()
	memStorage := memstorage.NewMemStorage()

	mux.Handle(routeGauge,
		http.StripPrefix(routeGauge, middleware.BadRequestIfMethodNotEqualPOST(
			http.HandlerFunc(server.Handler(storage.Gauge{Storage: memStorage})),
		)),
	)

	mux.Handle(routeCounter,
		http.StripPrefix(routeCounter, middleware.BadRequestIfMethodNotEqualPOST(
			http.HandlerFunc(server.Handler(storage.Counter{Storage: memStorage})),
		)),
	)

	if err := http.ListenAndServe(`localhost:8080`, mux); err != nil {
		panic(err)
	}
}
