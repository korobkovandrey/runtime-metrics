package controller

import (
	"fmt"

	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"

	"log"
	"net/http"
)

// IndexHandlerFunc @todo test!!!
func IndexHandlerFunc(store *repository.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := fmt.Fprintln(w, `<table><thead><tr><th>Type</th><th>Name</th><th>Value</th></tr><thead><tbody>`)
		if err != nil {
			log.Println(fmt.Errorf(`fmt.Fprintln: %w`, err))
		}
		for _, v := range store.GetAllData() {
			_, err = fmt.Fprintf(w, `<tr><td>%s</td><td>%s</td><td>%v</td></tr>`, v.T, v.Name, v.Value)
			if err != nil {
				log.Println(fmt.Errorf(`fmt.Fprintln: %w`, err))
			}
		}
		_, err = fmt.Fprintln(w, `</tbody></table>`)
		if err != nil {
			log.Println(fmt.Errorf(`fmt.Fprintln: %w`, err))
		}
	}
}
