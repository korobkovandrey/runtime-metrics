package controller

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValueJSONHandlerFunc(t *testing.T) {
	type want struct {
		code        int
		contentType string
		body        string
		json        string
	}
	tests := []struct {
		name    string
		body    string
		metrics []*model.Metric
		want    want
	}{
		{
			name: "empty",
			body: "",
			want: want{
				code:        400,
				body:        "Type is required.\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "without type",
			body: "{}",
			want: want{
				code:        400,
				body:        "Type is required.\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "without id",
			body: `{"type": "gauge"}`,
			want: want{
				code:        400,
				body:        "ID is required.\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "unknown gauge",
			body: `{"type": "gauge", "id": "unknown_id1"}`,
			want: want{
				code:        404,
				body:        "404 page not found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "unknown counter",
			body: `{"type": "counter", "id": "unknown_id1"}`,
			want: want{
				code:        404,
				body:        "404 page not found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:    "gauge",
			body:    `{"type": "gauge", "id": "test1"}`,
			metrics: []*model.Metric{model.NewMetricGauge("test1", 10)},
			want: want{
				code:        200,
				json:        `{"type": "gauge", "id": "test1", "value": 10}`,
				contentType: "application/json",
			},
		},
		{
			name:    "counter",
			body:    `{"type": "counter", "id": "test1"}`,
			metrics: []*model.Metric{model.NewMetricCounter("test1", 10)},
			want: want{
				code:        200,
				json:        `{"type": "counter", "id": "test1", "delta": 10}`,
				contentType: "application/json",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, err := repository.NewStoreMemStorage(&config.Config{})
			require.NoError(t, err)
			for _, m := range test.metrics {
				require.NoError(t, s.UpdateMetric(m))
			}

			var postBody io.Reader
			if test.body == "" {
				postBody = http.NoBody
			} else {
				postBody = strings.NewReader(test.body)
			}

			request := httptest.NewRequest(http.MethodPost, "/value/", postBody)

			w := httptest.NewRecorder()
			ValueJSONHandlerFunc(s)(w, request)

			res := w.Result()
			if res != nil {
				defer func() {
					err := res.Body.Close()
					require.NoError(t, err)
				}()
			}
			assert.Equal(t, test.want.code, res.StatusCode)
			if test.want.contentType != "" {
				assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
			}
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			if test.want.body != "" {
				assert.Equal(t, test.want.body, string(resBody))
			}
			if test.want.json != "" {
				require.NotEmpty(t, resBody)
				assert.JSONEq(t, test.want.json, string(resBody))
			}
		})
	}
}
