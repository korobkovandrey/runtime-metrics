package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/handlers/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type testCase struct {
	method          string
	url             string
	postBody        string
	wantCode        int
	wantContentType string
	wantJSON        string
	wantBody        string
	containsStrings []string
}

func TestHandler_setIndexRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tests := []struct {
		name      string
		mockSetup func(*mocks.MockAllFinder)
		testCase
	}{
		{
			name: "index ok",
			mockSetup: func(s *mocks.MockAllFinder) {
				s.EXPECT().
					FindAll(gomock.Any()).
					Return([]*model.Metric{
						model.NewMetricGauge("Alloc", 123.4),
						model.NewMetricCounter("PollCount", 10),
						model.NewMetricGauge("RandomValue", 12.55),
					}, nil)
			},
			testCase: testCase{
				method:          http.MethodGet,
				url:             "/",
				wantCode:        http.StatusOK,
				wantContentType: "text/html; charset=utf-8",
				containsStrings: []string{
					"Alloc", "123.4",
					"PollCount", "10",
					"RandomValue", "12.55",
				},
			},
		},
		{
			name: "index service error",
			mockSetup: func(s *mocks.MockAllFinder) {
				s.EXPECT().
					FindAll(gomock.Any()).
					Return(nil, errors.New("service error"))
			},
			testCase: testCase{
				method:   http.MethodGet,
				url:      "/",
				wantCode: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mocks.NewMockAllFinder(ctrl)
			tt.mockSetup(s)
			h := NewHandler()
			currentDir, err := os.Getwd()
			require.NoError(t, err)
			err = os.Chdir("../..")
			require.NoError(t, err)
			err = h.setIndexRoute(s)
			require.NoError(t, err)
			err = os.Chdir(currentDir)
			require.NoError(t, err)
			testHelper(t, h, tt.testCase)
		})
	}
}

func TestHandler_setPingRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tests := []struct {
		name      string
		mockSetup func(*mocks.MockPinger)
		testCase
	}{
		{
			name: "ping ok",
			mockSetup: func(s *mocks.MockPinger) {
				s.EXPECT().Ping(gomock.Any()).Return(nil).Times(1)
			},
			testCase: testCase{
				method:   http.MethodGet,
				url:      "/ping",
				wantCode: http.StatusOK,
			},
		},
		{
			name: "ping without db",
			testCase: testCase{
				method:   http.MethodGet,
				url:      "/ping",
				wantCode: http.StatusOK,
			},
		},
		{
			name: "ping error",
			mockSetup: func(s *mocks.MockPinger) {
				s.EXPECT().Ping(gomock.Any()).Return(errors.New("ping error")).Times(1)
			},
			testCase: testCase{
				method:   http.MethodGet,
				url:      "/ping",
				wantCode: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler()
			if tt.mockSetup != nil {
				s := mocks.NewMockPinger(ctrl)
				tt.mockSetup(s)
				h.setPingRoute(s)
			} else {
				h.setPingRoute(nil)
			}
			testHelper(t, h, tt.testCase)
		})
	}
}

func TestHandler_setUpdateRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tests := []struct {
		name      string
		mockSetup func(*mocks.MockUpdater)
		testCase
	}{
		{
			name: "updateURI gauge valid",
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().
					Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("test", 1)})).
					Return(model.NewMetricGauge("test", 1), nil)
			},
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/update/gauge/test/1",
				wantCode: http.StatusOK,
			},
		},
		{
			name: "updateURI counter valid",
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().
					Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricCounter("test", 1)})).
					Return(model.NewMetricCounter("test", 1), nil)
			},
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/update/counter/test/1",
				wantCode: http.StatusOK,
			},
		},
		{
			name: "updateURI service error",
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().
					Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricCounter("test", 1)})).
					Return(nil, errors.New("service error"))
			},
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/update/counter/test/1",
				wantCode: http.StatusInternalServerError,
			},
		},
		{
			name: "updateURI service error not found",
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().
					Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricCounter("test", 1)})).
					Return(nil, model.ErrMetricNotFound)
			},
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/update/counter/test/1",
				wantCode: http.StatusNotFound,
			},
		},
		{
			name: "updateURI fail value",
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			testCase: testCase{
				method:          http.MethodPost,
				url:             "/update/counter/test/fail",
				wantCode:        http.StatusBadRequest,
				containsStrings: []string{"value is not valid"},
			},
		},
		{
			name: "updateURI without value",
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			testCase: testCase{
				method:          http.MethodPost,
				url:             "/update/counter/test",
				wantCode:        http.StatusBadRequest,
				containsStrings: []string{"Value is required."},
			},
		},
		{
			name: "updateURI without name",
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/update/gauge",
				wantCode: http.StatusNotFound,
			},
		},
		{
			name: "updateJSON valid gauge",
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().
					Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("test", 12.34)})).
					Return(model.NewMetricGauge("test", 12.34), nil)
			},
			testCase: testCase{
				method:          http.MethodPost,
				url:             "/update",
				postBody:        `{"type":"gauge","id":"test","value":12.34}`,
				wantCode:        http.StatusOK,
				wantContentType: "application/json",
				wantJSON:        `{"type":"gauge","id":"test","value":12.34}`,
			},
		},
		{
			name: "updateJSON valid counter",
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().
					Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricCounter("test", 1)})).
					Return(model.NewMetricCounter("test", 1), nil)
			},
			testCase: testCase{
				method:          http.MethodPost,
				url:             "/update",
				postBody:        `{"type":"counter","id":"test","delta":1}`,
				wantCode:        http.StatusOK,
				wantContentType: "application/json",
				wantJSON:        `{"type":"counter","id":"test","delta":1}`,
			},
		},
		{
			name: "updateJSON invalid json",
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/update",
				postBody: `invalid`,
				wantCode: http.StatusBadRequest,
			},
		},
		{
			name: "updateJSON service error",
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().
					Update(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("error", 12.34)})).
					Return(nil, errors.New("service error"))
			},
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/update",
				postBody: `{"type":"gauge","id":"error","value":12.34}`,
				wantCode: http.StatusInternalServerError,
			},
		},
		{
			name: "updateJSON fail type",
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			testCase: testCase{
				method:          http.MethodPost,
				url:             "/update",
				postBody:        `{"type":"fail","id":"test"}`,
				wantCode:        http.StatusBadRequest,
				containsStrings: []string{"type is not valid"},
			},
		},
		{
			name: "updateJSON missing value",
			mockSetup: func(s *mocks.MockUpdater) {
				s.EXPECT().Update(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			testCase: testCase{
				method:          http.MethodPost,
				url:             "/update",
				postBody:        `{"type":"gauge","id":"test"}`,
				wantCode:        http.StatusBadRequest,
				containsStrings: []string{"value is not valid"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler()
			s := mocks.NewMockUpdater(ctrl)
			tt.mockSetup(s)
			h.setUpdateRoutes(s)
			testHelper(t, h, tt.testCase)
		})
	}
}

func TestHandler_setUpdatesRoute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tests := []struct {
		name      string
		mockSetup func(*mocks.MockBatchUpdater)
		testCase
	}{
		{
			name: "updatesJSON valid",
			mockSetup: func(s *mocks.MockBatchUpdater) {
				s.EXPECT().
					UpdateBatch(gomock.Any(), []*model.MetricRequest{
						{Metric: model.NewMetricGauge("test", 66.34)},
						{Metric: model.NewMetricCounter("test", 10)},
					}).
					Return([]*model.Metric{
						model.NewMetricGauge("test", 66.34),
						model.NewMetricCounter("test", 10),
					}, nil)
			},
			testCase: testCase{
				method:          http.MethodPost,
				url:             "/updates/",
				postBody:        `[{"type":"gauge","id":"test","value":66.34},{"type":"counter","id":"test","delta":10}]`,
				wantCode:        http.StatusOK,
				wantContentType: "application/json",
				wantJSON:        `[{"type":"gauge","id":"test","value":66.34},{"type":"counter","id":"test","delta":10}]`,
			},
		},
		{
			name: "updatesJSON invalid type",
			mockSetup: func(s *mocks.MockBatchUpdater) {
				s.EXPECT().UpdateBatch(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			testCase: testCase{
				method:          http.MethodPost,
				url:             "/updates/",
				postBody:        `[{"type":"invalid","id":"test","value":66.34},{"type":"counter","id":"test","delta":10}]`,
				wantCode:        http.StatusBadRequest,
				containsStrings: []string{"Bad Request", "type is not valid"},
			},
		},
		{
			name: "updatesJSON missing value",
			mockSetup: func(s *mocks.MockBatchUpdater) {
				s.EXPECT().UpdateBatch(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			testCase: testCase{
				method:          http.MethodPost,
				url:             "/updates/",
				postBody:        `[{"type":"gauge","id":"test","value":1.23},{"type":"counter","id":"test"}]`,
				wantCode:        http.StatusBadRequest,
				containsStrings: []string{"Bad Request", "value is not valid"},
			},
		},
		{
			name: "updatesJSON invalid json",
			mockSetup: func(s *mocks.MockBatchUpdater) {
				s.EXPECT().UpdateBatch(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/updates/",
				postBody: `invalid`,
				wantCode: http.StatusBadRequest,
			},
		},
		{
			name: "updatesJSON service error not found",
			mockSetup: func(s *mocks.MockBatchUpdater) {
				s.EXPECT().UpdateBatch(gomock.Any(), gomock.Any()).Return(nil, model.ErrMetricNotFound)
			},
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/updates/",
				postBody: `[{"type":"gauge","id":"test","value":1.23}]`,
				wantCode: http.StatusNotFound,
			},
		},
		{
			name: "updatesJSON service error",
			mockSetup: func(s *mocks.MockBatchUpdater) {
				s.EXPECT().UpdateBatch(gomock.Any(), gomock.Any()).Return(nil, errors.New("service error"))
			},
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/updates/",
				postBody: `[{"type":"gauge","id":"test","value":1.23}]`,
				wantCode: http.StatusInternalServerError,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler()
			s := mocks.NewMockBatchUpdater(ctrl)
			tt.mockSetup(s)
			h.setUpdatesRoute(s)
			testHelper(t, h, tt.testCase)
		})
	}
}

func TestHandler_setValueRoutes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	tests := []struct {
		name      string
		mockSetup func(*mocks.MockFinder)
		testCase
	}{
		{
			name: "valueURI gauge valid",
			mockSetup: func(s *mocks.MockFinder) {
				s.EXPECT().
					Find(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricGauge("test", 0)})).
					Return(model.NewMetricGauge("test", 10), nil)
			},
			testCase: testCase{
				method:          http.MethodGet,
				url:             "/value/gauge/test",
				wantCode:        http.StatusOK,
				wantContentType: "text/plain; charset=utf-8",
				wantBody:        "10",
			},
		},
		{
			name: "valueURI counter valid",
			mockSetup: func(s *mocks.MockFinder) {
				s.EXPECT().
					Find(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricCounter("test", 0)})).
					Return(model.NewMetricCounter("test", 111), nil)
			},
			testCase: testCase{
				method:          http.MethodGet,
				url:             "/value/counter/test",
				wantCode:        http.StatusOK,
				wantContentType: "text/plain; charset=utf-8",
				wantBody:        "111",
			},
		},
		{
			name: "valueURI service error",
			mockSetup: func(s *mocks.MockFinder) {
				s.EXPECT().
					Find(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricCounter("test", 0)})).
					Return(nil, errors.New("service error"))
			},
			testCase: testCase{
				method:   http.MethodGet,
				url:      "/value/counter/test",
				wantCode: http.StatusInternalServerError,
			},
		},
		{
			name: "valueURI not found metric",
			mockSetup: func(s *mocks.MockFinder) {
				s.EXPECT().
					Find(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: model.NewMetricCounter("test", 0)})).
					Return(nil, model.ErrMetricNotFound)
			},
			testCase: testCase{
				method:   http.MethodGet,
				url:      "/value/counter/test",
				wantCode: http.StatusNotFound,
			},
		},
		{
			name: "valueURI fail type",
			mockSetup: func(s *mocks.MockFinder) {
				s.EXPECT().Find(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			testCase: testCase{
				method:          http.MethodGet,
				url:             "/value/fail/name",
				wantCode:        http.StatusBadRequest,
				containsStrings: []string{"type is not valid"},
			},
		},
		{
			name: "valueURI without name",
			mockSetup: func(s *mocks.MockFinder) {
				s.EXPECT().Find(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			testCase: testCase{
				method:   http.MethodGet,
				url:      "/value/counter",
				wantCode: http.StatusNotFound,
			},
		},
		{
			name: "valueJSON valid gauge",
			mockSetup: func(s *mocks.MockFinder) {
				s.EXPECT().
					Find(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: &model.Metric{
						MType: model.TypeGauge,
						ID:    "test"}})).
					Return(model.NewMetricGauge("test", 12.34), nil)
			},
			testCase: testCase{
				method:          http.MethodPost,
				url:             "/value",
				postBody:        `{"type":"gauge","id":"test"}`,
				wantCode:        http.StatusOK,
				wantContentType: "application/json",
				wantJSON:        `{"type":"gauge","id":"test","value":12.34}`,
			},
		},
		{
			name: "valueJSON valid counter",
			mockSetup: func(s *mocks.MockFinder) {
				s.EXPECT().
					Find(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: &model.Metric{
						MType: model.TypeCounter,
						ID:    "test"}})).
					Return(model.NewMetricCounter("test", 111), nil)
			},
			testCase: testCase{
				method:          http.MethodPost,
				url:             "/value",
				postBody:        `{"type":"counter","id":"test"}`,
				wantCode:        http.StatusOK,
				wantContentType: "application/json",
				wantJSON:        `{"type":"counter","id":"test","delta":111}`,
			},
		},
		{
			name: "valueJSON invalid json",
			mockSetup: func(s *mocks.MockFinder) {
				s.EXPECT().Find(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/value",
				postBody: `invalid`,
				wantCode: http.StatusBadRequest,
			},
		},
		{
			name: "valueJSON service error",
			mockSetup: func(s *mocks.MockFinder) {
				s.EXPECT().
					Find(gomock.Any(), gomock.Eq(&model.MetricRequest{Metric: &model.Metric{
						MType: model.TypeGauge,
						ID:    "error"}})).
					Return(nil, errors.New("service error"))
			},
			testCase: testCase{
				method:   http.MethodPost,
				url:      "/value",
				postBody: `{"type":"gauge","id":"error"}`,
				wantCode: http.StatusInternalServerError,
			},
		},
		{
			name: "valueJSON fail type",
			mockSetup: func(s *mocks.MockFinder) {
				s.EXPECT().Find(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			testCase: testCase{
				method:          http.MethodPost,
				url:             "/value",
				postBody:        `{"type":"fail","id":"test"}`,
				wantCode:        http.StatusBadRequest,
				containsStrings: []string{"type is not valid"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler()
			s := mocks.NewMockFinder(ctrl)
			tt.mockSetup(s)
			h.setValueRoutes(s)
			testHelper(t, h, tt.testCase)
		})
	}
}

func testRequest(
	t *testing.T, ts *httptest.Server,
	method, path string, postBody io.Reader) (body []byte, statusCode int, contentType string) {
	t.Helper()
	req, err := http.NewRequestWithContext(context.TODO(), method, ts.URL+path, postBody)
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

func testHelper(t *testing.T, h http.Handler, tt testCase) {
	t.Helper()
	ts := httptest.NewServer(h)
	defer ts.Close()
	var postBody io.Reader
	if tt.postBody == "" {
		postBody = http.NoBody
	} else {
		postBody = bytes.NewBufferString(tt.postBody)
	}
	gotBody, gotCode, gotContentType := testRequest(t, ts, tt.method, tt.url, postBody)
	gotBodyString := string(gotBody)
	require.Equal(t, tt.wantCode, gotCode)
	if tt.wantContentType != "" {
		assert.Equal(t, tt.wantContentType, gotContentType)
	}
	if tt.wantJSON != "" {
		assert.JSONEq(t, tt.wantJSON, gotBodyString)
	}
	if tt.wantBody != "" {
		assert.JSONEq(t, tt.wantBody, gotBodyString)
	}
	for _, str := range tt.containsStrings {
		require.Contains(t, gotBodyString, str)
	}
}
