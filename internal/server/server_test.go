package server

import (
	"strings"

	"github.com/korobkovandrey/runtime-metrics/internal/server/config"
	"github.com/korobkovandrey/runtime-metrics/internal/server/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestServer_NewHandler(t *testing.T) {
	zapLogger, err := logger.NewZapLogger()
	require.NoError(t, err)
	s := New(&config.Config{}, zapLogger)
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir("../..")
	require.NoError(t, err)
	handler, err := s.NewHandler()
	require.NoError(t, err)
	err = os.Chdir(currentDir)
	require.NoError(t, err)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	tests := []struct {
		method      string
		url         string
		wantStatus  int
		wantBodyStr string
	}{
		{"GET", "/update/", http.StatusMethodNotAllowed, ""},
		{"GET", "/update/blabla", http.StatusMethodNotAllowed, ""},
		{"GET", "/update/blabla/bla", http.StatusMethodNotAllowed, ""},
		{"GET", "/update/blabla/bla/10", http.StatusMethodNotAllowed, ""},

		{"POST", "/update/", http.StatusBadRequest, "Type is required.\n"},
		// без имени метрики
		{"POST", "/update/gauge", http.StatusNotFound, ""},
		{"POST", "/update/gauge/", http.StatusNotFound, ""},
		// без значения
		{"POST", "/update/gauge/test1", http.StatusBadRequest, "Value is required.\n"},
		// неизвестный тип
		{"POST", "/update/blabla/test2/1", http.StatusBadRequest, "Bad Request: \"blabla\" type is not valid\n"},
		// некорректное значение
		{"POST", "/update/gauge/test3/fail_value", http.StatusBadRequest, "Bad Request: invalid number\n"},
		{"POST", "/update/counter/test4/1.5", http.StatusBadRequest, "Bad Request: invalid number\n"},
		// корректное значение gauge
		{"POST", "/update/gauge/test5/1.5", http.StatusOK, ""},
		{"POST", "/update/gauge/test6/0.001", http.StatusOK, ""},
		{"POST", "/update/gauge/test7/1.23456E-6", http.StatusOK, ""},
		// корректное значение counter
		{"POST", "/update/counter/test8/1", http.StatusOK, ""},
		{"POST", "/update/counter/test8/5", http.StatusOK, ""},
		{"POST", "/update/counter/test9/100", http.StatusOK, ""},
		{"POST", "/update/counter/test9/-10", http.StatusOK, ""},

		// тесты для получения данных
		// некорректрые запросы
		{"GET", "/value/gauge/blabla", http.StatusNotFound, ""},
		{"GET", "/value/counter/blabla", http.StatusNotFound, ""},
		{"GET", "/value/blabla/blabla", http.StatusBadRequest, "Bad Request: \"blabla\" type is not valid\n"},
		// корректные запросы
		{"GET", "/value/gauge/test5", http.StatusOK, "1.5"},
		{"GET", "/value/gauge/test6", http.StatusOK, "0.001"},
		{"GET", "/value/gauge/test7", http.StatusOK, "1.23456e-06"},
		{"GET", "/value/counter/test8", http.StatusOK, "6"},
		{"GET", "/value/counter/test9", http.StatusOK, "90"},
		// index
		{"GET", "/blablabla", http.StatusNotFound, ""},
		{"GET", "/", http.StatusOK, "<!DOCTYPE html>"},
	}
	for _, v := range tests {
		gotBody, gotStatusCode, _ := testRequest(t, ts, v.method, v.url, http.NoBody)
		assert.Equal(t, v.wantStatus, gotStatusCode, v.method+" "+v.url)

		if v.method == "GET" && v.url == "/" {
			assert.Contains(t, string(gotBody), v.wantBodyStr)
		} else if v.wantBodyStr != "" {
			if v.wantBodyStr != "" {
				assert.Equal(t, v.wantBodyStr, string(gotBody), v.method+" "+v.url)
			}
		}
	}
}

func TestServer_NewHandler_update(t *testing.T) {
	zapLogger, err := logger.NewZapLogger()
	require.NoError(t, err)
	s := New(&config.Config{}, zapLogger)
	currentDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir("../..")
	require.NoError(t, err)
	handler, err := s.NewHandler()
	require.NoError(t, err)
	err = os.Chdir(currentDir)
	require.NoError(t, err)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	type want struct {
		code        int
		contentType string
		body        string
		json        string
	}
	tests := []struct {
		name string
		body string
		want want
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
			name: "without value",
			body: `{"type": "gauge", "id": "test1"}`,
			want: want{
				code:        400,
				body:        "Bad Request: value is required\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name: "gauge",
			body: `{"type": "gauge", "id": "test1", "value": 10}`,
			want: want{
				code:        200,
				json:        `{"type": "gauge", "id": "test1", "value": 10}`,
				contentType: "application/json",
			},
		},
		{
			name: "counter",
			body: `{"type": "counter", "id": "test1", "delta": 10}`,
			want: want{
				code:        200,
				json:        `{"type": "counter", "id": "test1", "delta": 10}`,
				contentType: "application/json",
			},
		},
	}
	for _, test := range tests {
		var postBody io.Reader
		if test.body == "" {
			postBody = http.NoBody
		} else {
			postBody = strings.NewReader(test.body)
		}
		gotBody, gotStatusCode, contentType := testRequest(t, ts, http.MethodPost, "/update/", postBody)
		assert.Equal(t, test.want.code, gotStatusCode)
		if test.want.contentType != "" {
			assert.Equal(t, test.want.contentType, contentType)
		}
		if test.want.body != "" {
			assert.Equal(t, test.want.body, string(gotBody))
		}
		if test.want.json != "" {
			require.NotEmpty(t, gotBody)
			assert.JSONEq(t, test.want.json, string(gotBody))
		}
	}
}

func testRequest(
	t *testing.T, ts *httptest.Server,
	method, path string, postBody io.Reader) (body []byte, statusCode int, contentType string) {
	t.Helper()
	req, err := http.NewRequest(method, ts.URL+path, postBody)
	require.NoError(t, err)
	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()
	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	statusCode = resp.StatusCode
	contentType = resp.Header.Get("Content-Type")
	return
}
