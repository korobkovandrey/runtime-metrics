package controller

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
)

const strTypeIsRequired = "Type is required."

func readMetricFromRequest(w http.ResponseWriter, r *http.Request) (metric model.Metric, ok bool, err error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, strTypeIsRequired, http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		http.Error(w, strTypeIsRequired, http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(body, &metric)
	if err != nil {
		http.Error(w, strTypeIsRequired, http.StatusBadRequest)
		return
	}
	if len(metric.MType) == 0 {
		http.Error(w, strTypeIsRequired, http.StatusBadRequest)
		return
	}
	if len(metric.ID) == 0 {
		http.Error(w, "ID is required.", http.StatusBadRequest)
		return
	}
	ok = true
	return
}

func responseMetricJSON(metric *model.Metric, w http.ResponseWriter) (err error) {
	response, err := json.Marshal(metric)
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(response)
	if err != nil {
		return
	}
	return
}
