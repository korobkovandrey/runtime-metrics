package server

import (
	"errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type ServerTest interface {
	Run() error
}

func TestNew(t *testing.T) {
	s := New(Config{})
	assert.IsType(t, s, &Server{})
	assert.Implements(t, (*ServerTest)(nil), s)
}

func TestServer_NewHandler(t *testing.T) {
	s := New(Config{
		UpdatePath: `/update`,
	})
	ts := httptest.NewServer(s.NewHandler())
	defer ts.Close()

	tests := []struct {
		method      string
		url         string
		wantStatus  int
		wantBodyStr string
	}{
		{`GET`, `/update/`, http.StatusMethodNotAllowed, ``},
		{`GET`, `/update/blabla`, http.StatusMethodNotAllowed, ``},
		{`GET`, `/update/blabla/bla`, http.StatusMethodNotAllowed, ``},
		{`GET`, `/update/blabla/bla/10`, http.StatusMethodNotAllowed, ``},

		{`POST`, `/update/`, http.StatusBadRequest, "Type is required.\n"},
		// без имени метрики
		{`POST`, `/update/gauge`, http.StatusNotFound, ``},
		{`POST`, `/update/gauge/`, http.StatusNotFound, ``},
		// без значения
		{`POST`, `/update/gauge/test1`, http.StatusBadRequest, "Value is required.\n"},
		// неизвестный тип
		{`POST`, `/update/blabla/test2/1`, http.StatusBadRequest, "bad request: \"blabla\" type is not valid\n"},
		// некорректное значение
		{`POST`, `/update/gauge/test3/fail_value`, http.StatusBadRequest, "bad request: invalid number\n"},
		{`POST`, `/update/counter/test4/1.5`, http.StatusBadRequest, "bad request: invalid number\n"},
		// корректное значение gauge
		{`POST`, `/update/gauge/test5/1.5`, http.StatusOK, ``},
		{`POST`, `/update/gauge/test6/0.001`, http.StatusOK, ``},
		{`POST`, `/update/gauge/test7/1.23456E-6`, http.StatusOK, ``},
		// корректное значение counter
		{`POST`, `/update/counter/test8/1`, http.StatusOK, ``},
		{`POST`, `/update/counter/test9/100`, http.StatusOK, ``},
		{`POST`, `/update/counter/test10/-10`, http.StatusOK, ``},
	}
	for _, v := range tests {
		gotBody, gotStatusCode := testRequest(t, ts, v.method, v.url)
		assert.Equal(t, v.wantStatus, gotStatusCode, v.method+` `+v.url)
		if v.wantBodyStr != `` {
			assert.Equal(t, v.wantBodyStr, string(gotBody), v.method+` `+v.url)
		}
	}
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string) (body []byte, statusCode int) {
	t.Helper()
	req, err := http.NewRequest(method, ts.URL+path, http.NoBody)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	if resp != nil {
		defer func() {
			err = errors.Join(err, resp.Body.Close())
			require.NoError(t, err)
		}()
	}

	body, err = io.ReadAll(resp.Body)
	statusCode = resp.StatusCode
	return
}
