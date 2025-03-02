package mcompress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/pkg/logging"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestGzipCompressed(t *testing.T) {
	l, err := logging.NewZapLogger(zap.InfoLevel)
	require.NoError(t, err)

	type send struct {
		acceptGzip   bool
		encodingGzip bool
		text         []byte
	}
	type response struct {
		headers map[string]string
		text    []byte
	}

	tests := []struct {
		name     string
		send     send
		response response
		wantGzip bool
	}{
		{
			name: "send gzip, accept gzip, response gzip html",
			send: send{
				acceptGzip:   true,
				encodingGzip: true,
				text:         []byte("send hello"),
			},
			response: response{
				headers: map[string]string{"Content-Type": "text/html"},
				text:    []byte("response hello"),
			},
			wantGzip: true,
		},
		{
			name: "send not gzip, accept gzip, response gzip html",
			send: send{
				acceptGzip:   true,
				encodingGzip: false,
				text:         []byte("send hello"),
			},
			response: response{
				headers: map[string]string{"Content-Type": "text/html"},
				text:    []byte("response hello"),
			},
			wantGzip: true,
		},
		{
			name: "send gzip, accept gzip, response not gzip text/plain",
			send: send{
				acceptGzip:   true,
				encodingGzip: true,
				text:         []byte("send hello"),
			},
			response: response{
				headers: map[string]string{"Content-Type": "text/plain"},
				text:    []byte("response hello"),
			},
			wantGzip: false,
		},
		{
			name: "send gzip, accept gzip, response not gzip without Content-Type",
			send: send{
				acceptGzip:   true,
				encodingGzip: true,
				text:         []byte("send hello"),
			},
			response: response{
				headers: map[string]string{},
				text:    []byte("response hello"),
			},
			wantGzip: false,
		},

		{
			name: "send gzip, accept gzip, response gzip application/json",
			send: send{
				acceptGzip:   true,
				encodingGzip: true,
				text:         []byte("send hello"),
			},
			response: response{
				headers: map[string]string{"Content-Type": "application/json"},
				text:    []byte(`{"response": "hello"}`),
			},
			wantGzip: true,
		},
		{
			name: "send not gzip, accept gzip, response gzip application/json",
			send: send{
				acceptGzip:   true,
				encodingGzip: false,
				text:         []byte("send hello"),
			},
			response: response{
				headers: map[string]string{"Content-Type": "application/json"},
				text:    []byte(`{"response": "hello"}`),
			},
			wantGzip: true,
		},

		{
			name: "send gzip, not accept gzip, response gzip html",
			send: send{
				acceptGzip:   false,
				encodingGzip: true,
				text:         []byte("send hello"),
			},
			response: response{
				headers: map[string]string{"Content-Type": "text/html"},
				text:    []byte("response hello"),
			},
			wantGzip: false,
		},
		{
			name: "send not gzip, not accept gzip, response gzip html",
			send: send{
				acceptGzip:   false,
				encodingGzip: false,
				text:         []byte("send hello"),
			},
			response: response{
				headers: map[string]string{"Content-Type": "text/html"},
				text:    []byte("response hello"),
			},
			wantGzip: false,
		},
		{
			name: "send gzip, not accept gzip, response not gzip text/plain",
			send: send{
				acceptGzip:   false,
				encodingGzip: true,
				text:         []byte("send hello"),
			},
			response: response{
				headers: map[string]string{"Content-Type": "text/plain"},
				text:    []byte("response hello"),
			},
			wantGzip: false,
		},
		{
			name: "send gzip, not accept gzip, response not gzip without Content-Type",
			send: send{
				acceptGzip:   false,
				encodingGzip: true,
				text:         []byte("send hello"),
			},
			response: response{
				headers: map[string]string{},
				text:    []byte("response hello"),
			},
			wantGzip: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var postBody io.Reader
			if tt.send.encodingGzip {
				buf := bytes.NewBuffer(nil)
				gz := gzip.NewWriter(buf)
				_, err = gz.Write(tt.send.text)
				require.NoError(t, err)
				require.NoError(t, gz.Close())
				postBody = buf
			} else {
				postBody = bytes.NewBuffer(tt.send.text)
			}
			r := httptest.NewRequest(http.MethodGet, "/", postBody)

			if tt.send.encodingGzip {
				r.Header.Set("Content-Encoding", "gzip")
			}
			if tt.send.acceptGzip {
				r.Header.Set("Accept-Encoding", "gzip")
			}
			w := httptest.NewRecorder()
			GzipCompressed(l)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				require.Equal(t, string(tt.send.text), string(body))

				for k, v := range tt.response.headers {
					w.Header().Set(k, v)
				}
				w.WriteHeader(http.StatusOK)
				_, err = w.Write(tt.response.text)
				require.NoError(t, err)
			})).ServeHTTP(w, r)

			require.Equal(t, http.StatusOK, w.Code)
			var body []byte
			if tt.wantGzip {
				require.Contains(t, w.Header().Get("Content-Encoding"), "gzip")
				zr, err := gzip.NewReader(w.Body)
				require.NoError(t, err)
				body, err = io.ReadAll(zr)
				require.NoError(t, err)
				require.NoError(t, zr.Close())
			} else {
				require.NotContains(t, w.Header().Get("Content-Encoding"), "gzip")
				body = w.Body.Bytes()
			}
			require.Equal(t, string(tt.response.text), string(body))
		})
	}
}
