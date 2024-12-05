package controller

import (
	"fmt"
	"html/template"
	"sync"

	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"

	"log"
	"net/http"
)

type tplStore struct {
	tpl *template.Template
	mux *sync.Mutex
}

func (s *tplStore) getTpl() template.Template {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.tpl == nil {
		s.tpl = template.Must(template.ParseFiles("./web/template/index.html"))
	}
	return *s.tpl
}

func IndexHandlerFunc(store *repository.Store) func(w http.ResponseWriter, r *http.Request) {
	cache := &tplStore{
		mux: &sync.Mutex{},
	}
	return func(w http.ResponseWriter, r *http.Request) {
		tpl := cache.getTpl()
		err := tpl.Execute(w, store.GetAllData())
		if err != nil {
			log.Printf("IndexHandlerFunc tpl.Execute: %v", err)
		}
		_, err = fmt.Fprint(w, "Fail load template!!!")
		if err != nil {
			log.Printf("fmt.Fprintln: %v", err)
		}
	}
}
