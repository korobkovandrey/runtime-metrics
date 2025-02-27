package controller

import (
	"fmt"

	"html/template"
	"net/http"
)

func (c *Controller) indexFunc() (func(w http.ResponseWriter, r *http.Request), error) {
	tpl, err := template.ParseFiles("./web/template/index.html")
	if err != nil {
		return nil, fmt.Errorf("controller.indexFunc: %w", err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		data, err := c.s.FindAll()
		if err != nil {
			c.requestCtxWithLogMessageFromError(r, fmt.Errorf("controller.indexFunc: %w", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		if err = tpl.Execute(w, data); err != nil {
			c.requestCtxWithLogMessageFromError(r, fmt.Errorf("controller.indexFunc: %w", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}, nil
}
