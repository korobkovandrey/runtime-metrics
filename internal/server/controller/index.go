package controller

import (
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
	"net/http"
)

func IndexHandlerFunc(store *repository.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
