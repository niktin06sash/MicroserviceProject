// Code generated by MockGen. DO NOT EDIT.
// Source: service.go

// Package mock_service is a generated GoMock package.
package mock_service

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	model "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/model"
	repository "github.com/niktin06sash/MicroserviceProject/SessionManagement_service/internal/repository"
)

// MockSessionRepos is a mock of SessionRepos interface.
type MockSessionRepos struct {
	ctrl     *gomock.Controller
	recorder *MockSessionReposMockRecorder
}

// MockSessionReposMockRecorder is the mock recorder for MockSessionRepos.
type MockSessionReposMockRecorder struct {
	mock *MockSessionRepos
}

// NewMockSessionRepos creates a new mock instance.
func NewMockSessionRepos(ctrl *gomock.Controller) *MockSessionRepos {
	mock := &MockSessionRepos{ctrl: ctrl}
	mock.recorder = &MockSessionReposMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSessionRepos) EXPECT() *MockSessionReposMockRecorder {
	return m.recorder
}

// DeleteSession mocks base method.
func (m *MockSessionRepos) DeleteSession(ctx context.Context, sessionID string) *repository.RepositoryResponse {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteSession", ctx, sessionID)
	ret0, _ := ret[0].(*repository.RepositoryResponse)
	return ret0
}

// DeleteSession indicates an expected call of DeleteSession.
func (mr *MockSessionReposMockRecorder) DeleteSession(ctx, sessionID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteSession", reflect.TypeOf((*MockSessionRepos)(nil).DeleteSession), ctx, sessionID)
}

// GetSession mocks base method.
func (m *MockSessionRepos) GetSession(ctx context.Context, sessionID, flag string) *repository.RepositoryResponse {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSession", ctx, sessionID, flag)
	ret0, _ := ret[0].(*repository.RepositoryResponse)
	return ret0
}

// GetSession indicates an expected call of GetSession.
func (mr *MockSessionReposMockRecorder) GetSession(ctx, sessionID, flag interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSession", reflect.TypeOf((*MockSessionRepos)(nil).GetSession), ctx, sessionID, flag)
}

// SetSession mocks base method.
func (m *MockSessionRepos) SetSession(ctx context.Context, session *model.Session) *repository.RepositoryResponse {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetSession", ctx, session)
	ret0, _ := ret[0].(*repository.RepositoryResponse)
	return ret0
}

// SetSession indicates an expected call of SetSession.
func (mr *MockSessionReposMockRecorder) SetSession(ctx, session interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetSession", reflect.TypeOf((*MockSessionRepos)(nil).SetSession), ctx, session)
}

// MockLogProducer is a mock of LogProducer interface.
type MockLogProducer struct {
	ctrl     *gomock.Controller
	recorder *MockLogProducerMockRecorder
}

// MockLogProducerMockRecorder is the mock recorder for MockLogProducer.
type MockLogProducerMockRecorder struct {
	mock *MockLogProducer
}

// NewMockLogProducer creates a new mock instance.
func NewMockLogProducer(ctrl *gomock.Controller) *MockLogProducer {
	mock := &MockLogProducer{ctrl: ctrl}
	mock.recorder = &MockLogProducerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLogProducer) EXPECT() *MockLogProducerMockRecorder {
	return m.recorder
}

// NewSessionLog mocks base method.
func (m *MockLogProducer) NewSessionLog(level, place, traceid, msg string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "NewSessionLog", level, place, traceid, msg)
}

// NewSessionLog indicates an expected call of NewSessionLog.
func (mr *MockLogProducerMockRecorder) NewSessionLog(level, place, traceid, msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewSessionLog", reflect.TypeOf((*MockLogProducer)(nil).NewSessionLog), level, place, traceid, msg)
}
