package msign

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/pkg/sign"
	"github.com/stretchr/testify/require"
)

func TestSigner(t *testing.T) {
	type args struct {
		text     string
		response string
		key      string
	}
	type fields struct {
		key string
	}
	tests := []struct {
		name           string
		args           args
		fields         fields
		wantErr        bool
		wantSignHeader bool
	}{
		{
			name: "valid",
			args: args{
				text:     "send hello",
				response: "response hello",
				key:      "sing key",
			},
			fields: fields{
				key: "sing key",
			},
			wantSignHeader: true,
		},
		{
			name: "fail",
			args: args{
				text:     "send hello",
				response: "response hello",
				key:      "other sing key",
			},
			fields: fields{
				key: "sing key",
			},
			wantErr:        true,
			wantSignHeader: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.args.text))
			if len(tt.args.key) > 0 {
				r.Header.Set("HashSHA256", sign.MakeToString([]byte(tt.args.text), []byte(tt.args.key)))
			}

			w := httptest.NewRecorder()
			Signer([]byte(tt.fields.key))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := io.ReadAll(r.Body)
				if tt.wantErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				w.WriteHeader(http.StatusOK)
				_, err = w.Write([]byte(tt.args.response))
				require.NoError(t, err)
			})).ServeHTTP(w, r)
			if tt.wantSignHeader {
				require.Equal(t, w.Header().Get("HashSHA256"), sign.MakeToString([]byte(tt.args.response), []byte(tt.fields.key)))
			}
		})
	}
}
