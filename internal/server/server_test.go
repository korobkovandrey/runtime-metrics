package server

import (
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
	require.NoError(t, logger.Initialize())
	s := New(&config.Config{})
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
		{"POST", "/update/blabla/test2/1", http.StatusBadRequest, "bad request: \"blabla\" type is not valid\n"},
		// некорректное значение
		{"POST", "/update/gauge/test3/fail_value", http.StatusBadRequest, "bad request: invalid number\n"},
		{"POST", "/update/counter/test4/1.5", http.StatusBadRequest, "bad request: invalid number\n"},
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
		{"GET", "/value/blabla/blabla", http.StatusBadRequest, "bad request: \"blabla\" type is not valid\n"},
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
		gotBody, gotStatusCode := testRequest(t, ts, v.method, v.url)
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

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (body []byte, statusCode int) {
	t.Helper()
	req, err := http.NewRequest(method, ts.URL+path, http.NoBody)
	require.NoError(t, err)
	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, resp.Body.Close())
	}()
	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	statusCode = resp.StatusCode
	return
}
