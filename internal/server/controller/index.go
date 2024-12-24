package controller

import (
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"

	"html/template"
	"log"
	"net/http"
)

func IndexHandlerFunc(store *repository.Store) (func(w http.ResponseWriter, r *http.Request), error) {
	tpl, err := template.ParseFiles("./web/template/index.html")
	if err != nil {
		return nil, fmt.Errorf("IndexHandlerFunc parse template: %w", err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		if err := tpl.Execute(w, store.GetAllData()); err != nil {
			log.Printf("IndexHandlerFunc tpl.Execute: %v", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}, nil
}
