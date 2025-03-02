// Code generated by MockGen. DO NOT EDIT.
// Source: service.go
//
// Generated by this command:
//
//	mockgen -source=service.go -destination=../mocks/service.go -package=mocks
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	model "github.com/korobkovandrey/runtime-metrics/internal/model"
	gomock "go.uber.org/mock/gomock"
)

// MockRepository is a mock of Repository interface.
type MockRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRepositoryMockRecorder
	isgomock struct{}
}

// MockRepositoryMockRecorder is the mock recorder for MockRepository.
type MockRepositoryMockRecorder struct {
	mock *MockRepository
}

// NewMockRepository creates a new mock instance.
func NewMockRepository(ctrl *gomock.Controller) *MockRepository {
	mock := &MockRepository{ctrl: ctrl}
	mock.recorder = &MockRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRepository) EXPECT() *MockRepositoryMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockRepository) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockRepositoryMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockRepository)(nil).Close))
}

// Create mocks base method.
func (m *MockRepository) Create(mr *model.MetricRequest) (*model.Metric, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", mr)
	ret0, _ := ret[0].(*model.Metric)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr_2 *MockRepositoryMockRecorder) Create(mr any) *gomock.Call {
	mr_2.mock.ctrl.T.Helper()
	return mr_2.mock.ctrl.RecordCallWithMethodType(mr_2.mock, "Create", reflect.TypeOf((*MockRepository)(nil).Create), mr)
}

// Find mocks base method.
func (m *MockRepository) Find(mr *model.MetricRequest) (*model.Metric, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Find", mr)
	ret0, _ := ret[0].(*model.Metric)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Find indicates an expected call of Find.
func (mr_2 *MockRepositoryMockRecorder) Find(mr any) *gomock.Call {
	mr_2.mock.ctrl.T.Helper()
	return mr_2.mock.ctrl.RecordCallWithMethodType(mr_2.mock, "Find", reflect.TypeOf((*MockRepository)(nil).Find), mr)
}

// FindAll mocks base method.
func (m *MockRepository) FindAll() ([]*model.Metric, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindAll")
	ret0, _ := ret[0].([]*model.Metric)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindAll indicates an expected call of FindAll.
func (mr *MockRepositoryMockRecorder) FindAll() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindAll", reflect.TypeOf((*MockRepository)(nil).FindAll))
}

// Update mocks base method.
func (m *MockRepository) Update(mr *model.MetricRequest) (*model.Metric, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", mr)
	ret0, _ := ret[0].(*model.Metric)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update.
func (mr_2 *MockRepositoryMockRecorder) Update(mr any) *gomock.Call {
	mr_2.mock.ctrl.T.Helper()
	return mr_2.mock.ctrl.RecordCallWithMethodType(mr_2.mock, "Update", reflect.TypeOf((*MockRepository)(nil).Update), mr)
}
