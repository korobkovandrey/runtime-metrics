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

func TestUpdateHandler(t *testing.T) {
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
		name   string
		target string
		want   want
	}{
		{
			name:   `empty`,
			target: `/`,
			want: want{
				code:        400,
				response:    "Type is required.\n",
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name:   `one type`,
			target: `/` + gaugeType,
			want: want{
				code:        404,
				response:    "404 page not found\n",
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name:   `without value`,
			target: `/fail_type/name`,
			want: want{
				code:        400,
				response:    "Value is required.\n",
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name:   `not exists type`,
			target: `/fail_type/name/10`,
			want: want{
				code:        400,
				response:    "bad request: \"fail_type\" type is not valid\n",
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name:   `string value gauge`,
			target: `/` + gaugeType + `/name/fail_value`,
			want: want{
				code:        400,
				response:    "bad request: invalid number\n",
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name:   `string value counter`,
			target: `/` + counterType + `/name/fail_value`,
			want: want{
				code:        400,
				response:    "bad request: invalid number\n",
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name:   `gauge ok`,
			target: `/` + gaugeType + `/name/10.12344`,
			want: want{
				code:        200,
				response:    "",
				contentType: ``,
			},
		},
		{
			name:   `counter fail float`,
			target: `/` + counterType + `/name/10.12344`,
			want: want{
				code:        400,
				response:    "bad request: invalid number\n",
				contentType: `text/plain; charset=utf-8`,
			},
		},
		{
			name:   `counter 1`,
			target: `/` + counterType + `/name/1`,
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
			request := httptest.NewRequest(http.MethodPost, test.target, http.NoBody)
			w := httptest.NewRecorder()
			UpdateHandler(s)(w, request)

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
