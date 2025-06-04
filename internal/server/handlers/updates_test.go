package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/korobkovandrey/runtime-metrics/internal/model"
	"github.com/korobkovandrey/runtime-metrics/internal/server/handlers/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestNewUpdatesHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		mockSetup       func(*mocks.MockBatchUpdater)
		name            string
		json            string
		wantContentType string
		wantJSON        string
		containsStrings []string
		wantCode        int
	}{
		{
			name: "valid",
			json: `[{"type":"gauge","id":"test","value":66.34},
					{"type":"counter","id":"test","delta":10}]`,
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
			wantCode:        http.StatusOK,
			wantContentType: "application/json",
			wantJSON:        `[{"type":"gauge","id":"test","value":66.34},{"type":"counter","id":"test","delta":10}]`,
		},
		{
			name: "invalid type",
			json: `[{"type":"invalid","id":"test","value":66.34},
					{"type":"counter","id":"test","delta":10}]`,
			mockSetup: func(s *mocks.MockBatchUpdater) {
				s.EXPECT().UpdateBatch(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			wantCode:        http.StatusBadRequest,
			containsStrings: []string{"Bad Request", "type is not valid"},
		},
		{
			name: "missing value",
			json: `[{"type":"gauge","id":"test","value":1.23},
					{"type":"counter","id":"test"}]`,
			mockSetup: func(s *mocks.MockBatchUpdater) {
				s.EXPECT().UpdateBatch(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			wantCode:        http.StatusBadRequest,
			containsStrings: []string{"Bad Request", "value is not valid"},
		},
		{
			name: "invalid json",
			json: `invalid`,
			mockSetup: func(s *mocks.MockBatchUpdater) {
				s.EXPECT().UpdateBatch(gomock.Any(), gomock.Any()).MaxTimes(0)
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name: "service error not found",
			json: `[{"type":"gauge","id":"test","value":1.23}]`,
			mockSetup: func(s *mocks.MockBatchUpdater) {
				s.EXPECT().UpdateBatch(gomock.Any(), gomock.Any()).Return(nil, model.ErrMetricNotFound)
			},
			wantCode: http.StatusNotFound,
		},
		{
			name: " service error",
			json: `[{"type":"gauge","id":"test","value":1.23}]`,
			mockSetup: func(mockService *mocks.MockBatchUpdater) {
				mockService.EXPECT().
					UpdateBatch(gomock.Any(), gomock.Any()).Return(nil, errors.New("service error"))
			},
			wantCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mocks.NewMockBatchUpdater(ctrl)
			tt.mockSetup(s)
			r := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewBufferString(tt.json))
			w := httptest.NewRecorder()
			NewUpdatesHandler(s)(w, r)
			require.Equal(t, tt.wantCode, w.Code)
			body := w.Body.String()
			if tt.wantJSON != "" {
				require.JSONEq(t, tt.wantJSON, body)
			}
			for _, str := range tt.containsStrings {
				require.Contains(t, body, str)
			}
		})
	}
}
