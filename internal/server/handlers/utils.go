package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/mlogger"
)

func requestCtxWithLogMessage(r *http.Request, msg string) {
	*r = *r.WithContext(context.WithValue(r.Context(), mlogger.LogMessageKey, msg))
}

func requestCtxWithLogMessageFromError(r *http.Request, err error) {
	requestCtxWithLogMessage(r, err.Error())
}

func responseMarshaled(data any, w http.ResponseWriter, r *http.Request) {
	response, err := json.Marshal(data)
	if err == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(response)
	}
	if err != nil {
		requestCtxWithLogMessageFromError(r, fmt.Errorf("failed response: %w", err))
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}
