package compress

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Мок для http.ResponseWriter
type mockResponseWriter struct {
	mock.Mock
}

func (m *mockResponseWriter) Header() http.Header {
	args := m.Called()
	return args.Get(0).(http.Header)
}

func (m *mockResponseWriter) Write(p []byte) (int, error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.Called(statusCode)
}

func TestNewCompressWriter(t *testing.T) {
	w := httptest.NewRecorder()
	cw := NewCompressWriter(w)
	assert.NotNil(t, cw)
	assert.Equal(t, w, cw.w)
	assert.False(t, cw.Compressible)
	assert.Nil(t, cw.zw)
}

func TestWriter_Header(t *testing.T) {
	w := httptest.NewRecorder()
	cw := NewCompressWriter(w)
	header := cw.Header()
	assert.Equal(t, w.Header(), header)
}

func TestWriter_Write_Compressible(t *testing.T) {
	w := httptest.NewRecorder()
	cw := NewCompressWriter(w)
	cw.Compressible = true

	data := []byte("test data")
	n, err := cw.Write(data)
	assert.NoError(t, err)
	assert.NotNil(t, cw.zw)
	assert.Equal(t, len(data), n)

	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err = gw.Write(data)
	require.NoError(t, err)
	assert.Equal(t, buf.Bytes(), w.Body.Bytes())
	require.NoError(t, gw.Close())
}

func TestWriter_Write_NonCompressible(t *testing.T) {
	w := httptest.NewRecorder()
	cw := NewCompressWriter(w)
	cw.Compressible = false

	data := []byte("test data")
	n, err := cw.Write(data)
	assert.NoError(t, err)
	assert.Nil(t, cw.zw)
	assert.Equal(t, len(data), n)
	assert.Equal(t, data, w.Body.Bytes())
}

func TestWriter_Write_Compressible_Error(t *testing.T) {
	mw := &mockResponseWriter{}
	mw.On("Write", mock.Anything).Return(0, errors.New("write error"))

	cw := NewCompressWriter(mw)
	cw.Compressible = true

	data := []byte("test data")
	n, err := cw.Write(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "compress[Writer].Write: write error")
	assert.Equal(t, 0, n)
	mw.AssertExpectations(t)
}

func TestWriter_WriteHeader_Compressible(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		statusCode  int
		compress    bool
	}{
		{"HTML_OK", "text/html", http.StatusOK, true},
		{"JSON_OK", "application/json", http.StatusCreated, true},
		{"HTML_BadRequest", "text/html", http.StatusBadRequest, false},
		{"PlainText_OK", "text/plain", http.StatusOK, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			cw := NewCompressWriter(w)
			w.Header().Set("Content-Type", tt.contentType)
			w.Header().Set("Content-Length", "100")

			cw.WriteHeader(tt.statusCode)

			assert.Equal(t, tt.compress, cw.Compressible)
			if tt.compress {
				assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
				assert.Empty(t, w.Header().Get("Content-Length"))
			} else {
				assert.Empty(t, w.Header().Get("Content-Encoding"))
				assert.Equal(t, "100", w.Header().Get("Content-Length"))
			}
			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestWriter_Close(t *testing.T) {
	w := httptest.NewRecorder()
	cw := NewCompressWriter(w)
	cw.Compressible = true

	_, _ = cw.Write([]byte("test"))
	err := cw.Close()
	assert.NoError(t, err)
	assert.NotNil(t, cw.zw)

	zr, err := gzip.NewReader(w.Body)
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, zr.Close())
	}()
	data, err := io.ReadAll(zr)
	assert.NoError(t, err)
	assert.Equal(t, "test", string(data))
}

func TestWriter_Close_NoGzip(t *testing.T) {
	w := httptest.NewRecorder()
	cw := NewCompressWriter(w)
	err := cw.Close()
	assert.NoError(t, err)
}

func TestNewCompressReader(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write([]byte("test data"))
	require.NoError(t, err)
	require.NoError(t, gw.Close())

	cr, err := NewCompressReader(io.NopCloser(&buf))
	assert.NoError(t, err)
	assert.NotNil(t, cr)
	assert.NotNil(t, cr.zr)
}

func TestNewCompressReader_Error(t *testing.T) {
	cr, err := NewCompressReader(&errorReader{err: errors.New("invalid input")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "NewCompressReader: invalid input")
	assert.Nil(t, cr)
}

func TestReader_Read(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write([]byte("test data"))
	require.NoError(t, err)
	require.NoError(t, gw.Close())

	cr, err := NewCompressReader(io.NopCloser(&buf))
	assert.NoError(t, err)

	p := make([]byte, 100)
	n, err := cr.Read(p)
	if err != nil {
		assert.ErrorIs(t, err, io.EOF)
	}
	assert.Equal(t, 9, n)
	assert.Equal(t, "test data", string(p[:n]))

	n, err = cr.Read(p)
	assert.Equal(t, 0, n)
	assert.ErrorIs(t, err, io.EOF)
}

func TestReader_Read_Error(t *testing.T) {
	buf := bytes.NewReader([]byte("invalid gzip data"))
	_, err := NewCompressReader(io.NopCloser(buf))
	assert.Error(t, err)
}

func TestReader_Close(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, err := gw.Write([]byte("test data"))
	require.NoError(t, err)
	require.NoError(t, gw.Close())

	cr, err := NewCompressReader(io.NopCloser(&buf))
	assert.NoError(t, err)

	err = cr.Close()
	assert.NoError(t, err)
}

type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (int, error) {
	return 0, r.err
}

func (r *errorReader) Close() error {
	return r.err
}
