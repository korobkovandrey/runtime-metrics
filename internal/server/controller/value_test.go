package controller

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueHandlerFunc(t *testing.T) {
	const (
		gaugeType   = "gauge"
		counterType = "counter"
	)
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name       string
		pathValues map[string]string
		want       want
	}{
		{
			name:       "empty",
			pathValues: map[string]string{},
			want: want{
				code:        400,
				response:    "Type is required.\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "one type",
			pathValues: map[string]string{
				"type": gaugeType,
			},
			want: want{
				code:        404,
				response:    "404 page not found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "not exists name",
			pathValues: map[string]string{
				"type": gaugeType,
				"name": "fail_name",
			},
			want: want{
				code:        404,
				response:    "404 page not found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "gauge ok",
			pathValues: map[string]string{
				"type": gaugeType,
				"name": "name",
			},
			want: want{
				code:        200,
				response:    "10.1",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "counter ok",
			pathValues: map[string]string{
				"type": counterType,
				"name": "name",
			},
			want: want{
				code:        200,
				response:    "10",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}
	s := repository.NewStoreMemStorage()
	valueFloat64 := 10.1
	valueInt64 := int64(10)
	require.NoError(t, s.UpdateMetric(&model.Metric{
		MType: gaugeType,
		ID:    "name",
		Value: &valueFloat64,
	}))
	require.NoError(t, s.UpdateMetric(&model.Metric{
		MType: counterType,
		ID:    "name",
		Delta: &valueInt64,
	}))
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			target := ""
			for _, v := range test.pathValues {
				target += "/" + v
			}
			if target == "" {
				target = "/"
			}
			request := httptest.NewRequest(http.MethodPost, target, http.NoBody)
			for i, v := range test.pathValues {
				request.SetPathValue(i, v)
			}
			w := httptest.NewRecorder()
			ValueHandlerFunc(s)(w, request)

			res := w.Result()
			if res != nil {
				defer func() {
					err := res.Body.Close()
					require.NoError(t, err)
				}()
			}
			assert.Equal(t, test.want.code, res.StatusCode)
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			assert.Equal(t, test.want.response, string(resBody))
		})
	}
}
