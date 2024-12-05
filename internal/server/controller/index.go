package controller

import (
	"fmt"
	"io"
	"sync"

	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"

	"html/template"
	"log"
	"net/http"
)

type tplIndexPage struct {
	tpl *template.Template
	mux *sync.Mutex
}

func (s *tplIndexPage) execute(w io.Writer, data any) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	if s.tpl == nil {
		s.tpl = template.Must(template.ParseFiles("./web/template/index.html"))
	}
	err := s.tpl.Execute(w, data)
	if err != nil {
		return fmt.Errorf("execute: %w", err)
	}
	return nil
}

func IndexHandlerFunc(store *repository.Store) func(w http.ResponseWriter, r *http.Request) {
	tpl := tplIndexPage{
		mux: &sync.Mutex{},
	}
	return func(w http.ResponseWriter, r *http.Request) {
		err := tpl.execute(w, store.GetAllData())
		if err != nil {
			log.Printf("IndexHandlerFunc tpl: %v", err)
			_, err = fmt.Fprint(w, "Fail load template!!!")
			if err != nil {
				log.Printf("IndexHandlerFunc fmt.Fprintln: %v", err)
			}
		}
	}
}
