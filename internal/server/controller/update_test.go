package controller

import (
	"github.com/korobkovandrey/runtime-metrics/internal/server/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"io"
	"net/http"
	"net/http/httptest"

	"testing"
)

func TestUpdateHandlerFunc(t *testing.T) {
	const (
		gaugeType   = `gauge`
		counterType = `counter`
	)
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name       string
		target     string
		pathValues map[string]string
		want       want
	}{
		{
			name:       `empty`,
			pathValues: map[string]string{},
			want: want{
				code:        400,
				response:    "Type is required.\n",
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name: `one type`,
			pathValues: map[string]string{
				`type`: gaugeType,
			},
			want: want{
				code:        404,
				response:    "404 page not found\n",
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name: `without value`,
			pathValues: map[string]string{
				`type`: `fail_type`,
				`name`: `name`,
			},
			want: want{
				code:        400,
				response:    "Value is required.\n",
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name: `not exists type`,
			pathValues: map[string]string{
				`type`:  `fail_type`,
				`name`:  `name`,
				`value`: `10`,
			},
			want: want{
				code:        400,
				response:    "bad request: \"fail_type\" type is not valid\n",
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name: `string value gauge`,
			pathValues: map[string]string{
				`type`:  gaugeType,
				`name`:  `name`,
				`value`: `fail_value`,
			},
			want: want{
				code:        400,
				response:    "bad request: invalid number\n",
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name: `string value counter`,
			pathValues: map[string]string{
				`type`:  counterType,
				`name`:  `name`,
				`value`: `fail_value`,
			},
			want: want{
				code:        400,
				response:    "bad request: invalid number\n",
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name: `gauge ok`,
			pathValues: map[string]string{
				`type`:  gaugeType,
				`name`:  `name`,
				`value`: `10.12344`,
			},
			want: want{
				code:        200,
				response:    "",
				contentType: ``,
			},
		},
		{
			name: `counter fail float`,
			pathValues: map[string]string{
				`type`:  counterType,
				`name`:  `name`,
				`value`: `10.12344`,
			},
			want: want{
				code:        400,
				response:    "bad request: invalid number\n",
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name: `counter 1`,
			pathValues: map[string]string{
				`type`:  counterType,
				`name`:  `name`,
				`value`: `1`,
			},
			want: want{
				code:        200,
				response:    "",
				contentType: ``,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s := repository.NewStoreMemStorage()
			target := ``
			for _, v := range test.pathValues {
				target += `/` + v
			}
			if target == `` {
				target = `/`
			}
			request := httptest.NewRequest(http.MethodPost, target, http.NoBody)
			for i, v := range test.pathValues {
				request.SetPathValue(i, v)
			}
			w := httptest.NewRecorder()
			UpdateHandlerFunc(s)(w, request)

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
