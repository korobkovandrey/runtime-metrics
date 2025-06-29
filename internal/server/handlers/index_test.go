package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/handlers/mocks"
	"github.com/korobkovandrey/runtime-metrics/internal/server/middleware/mlogger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type errWriter struct {
	*httptest.ResponseRecorder
	err error
}

func (ew *errWriter) Write(buf []byte) (int, error) {
	if ew.err != nil {
		return 0, ew.err
	}
	return ew.Write(buf)
}

func TestNewIndexHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		serviceError    error
		name            string
		serviceResponse []*model.Metric
		containsStrings []string
		wantCode        int
	}{
		{
			name: "success",
			serviceResponse: []*model.Metric{
				model.NewMetricGauge("Alloc", 123.4),
				model.NewMetricCounter("PollCount", 10),
				model.NewMetricGauge("RandomValue", 12.55),
			},
			wantCode: http.StatusOK,
			containsStrings: []string{
				"Alloc", "123.4",
				"PollCount", "10",
				"RandomValue", "12.55",
			},
		},
		{
			name:         "error",
			serviceError: errors.New("error"),
			wantCode:     http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mocks.NewMockAllFinder(ctrl)
			s.EXPECT().FindAll(gomock.Any()).Return(tt.serviceResponse, tt.serviceError)

			currentDir, err := os.Getwd()
			require.NoError(t, err)
			t.Chdir("../../..")
			handler, err := NewIndexHandler(s)
			require.NoError(t, err)
			t.Chdir(currentDir)

			r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			w := httptest.NewRecorder()
			handler(w, r)

			require.Equal(t, tt.wantCode, w.Code)
			body := w.Body.String()
			for _, str := range tt.containsStrings {
				require.Contains(t, body, str)
			}
		})
	}

	t.Run("fail template path", func(t *testing.T) {
		_, err := NewIndexHandler(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse template")
	})

	t.Run("fail template execute", func(t *testing.T) {
		s := mocks.NewMockAllFinder(ctrl)
		s.EXPECT().FindAll(gomock.Any()).Return(nil, nil)
		currentDir, err := os.Getwd()
		require.NoError(t, err)
		t.Chdir("../../..")
		handler, err := NewIndexHandler(s)
		require.NoError(t, err)
		t.Chdir(currentDir)
		r := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
		w := &errWriter{ResponseRecorder: httptest.NewRecorder(), err: errors.New("error")}
		handler(w, r)
		m, _ := r.Context().Value(mlogger.LogMessageKey).(string)
		assert.Equal(t, "failed to execute template: error", m)
		fmt.Println(r.Context().Value(mlogger.LogMessageKey), w.Body.String())
	})
}
