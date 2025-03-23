// Code generated by MockGen. DO NOT EDIT.
// Source: updatejson.go
//
// Generated by this command:
//
//	mockgen -source=updatejson.go -destination=mock_updater.go -package=handlers
//

// Package handlers is a generated GoMock package.
package handlers

import (
	context "context"
	reflect "reflect"

	model "github.com/korobkovandrey/runtime-metrics/internal/model"
	gomock "go.uber.org/mock/gomock"
)

// MockUpdater is a mock of Updater interface.
type MockUpdater struct {
	ctrl     *gomock.Controller
	recorder *MockUpdaterMockRecorder
	isgomock struct{}
}

// MockUpdaterMockRecorder is the mock recorder for MockUpdater.
type MockUpdaterMockRecorder struct {
	mock *MockUpdater
}

// NewMockUpdater creates a new mock instance.
func NewMockUpdater(ctrl *gomock.Controller) *MockUpdater {
	mock := &MockUpdater{ctrl: ctrl}
	mock.recorder = &MockUpdaterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUpdater) EXPECT() *MockUpdaterMockRecorder {
	return m.recorder
}

// Update mocks base method.
func (m *MockUpdater) Update(ctx context.Context, mr *model.MetricRequest) (*model.Metric, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, mr)
	ret0, _ := ret[0].(*model.Metric)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update.
func (mr_2 *MockUpdaterMockRecorder) Update(ctx, mr any) *gomock.Call {
	mr_2.mock.ctrl.T.Helper()
	return mr_2.mock.ctrl.RecordCallWithMethodType(mr_2.mock, "Update", reflect.TypeOf((*MockUpdater)(nil).Update), ctx, mr)
}
