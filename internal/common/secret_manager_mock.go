package common

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockSecretManager is a mock of SecretManager interface.
type MockSecretManager struct {
	ctrl     *gomock.Controller
	recorder *MockSecretManagerMockRecorder
}

// MockSecretManagerMockRecorder is the mock recorder for MockSecretManager.
type MockSecretManagerMockRecorder struct {
	mock *MockSecretManager
}

// NewMockSecretManager creates a new mock instance.
func NewMockSecretManager(ctrl *gomock.Controller) *MockSecretManager {
	mock := &MockSecretManager{ctrl: ctrl}
	mock.recorder = &MockSecretManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSecretManager) EXPECT() *MockSecretManagerMockRecorder {
	return m.recorder
}

// CreateEngine mocks base method.
func (m *MockSecretManager) CreateEngine(ctx context.Context, engine, engineType string, options map[string]string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateEngine", ctx, engine, engineType, options)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateEngine indicates an expected call of CreateEngine.
func (mr *MockSecretManagerMockRecorder) CreateEngine(ctx, engine, engineType, options interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateEngine", reflect.TypeOf((*MockSecretManager)(nil).CreateEngine), ctx, engine, engineType, options)
}

// DeleteEngine mocks base method.
func (m *MockSecretManager) DeleteEngine(ctx context.Context, engine string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteEngine", ctx, engine)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteEngine indicates an expected call of DeleteEngine.
func (mr *MockSecretManagerMockRecorder) DeleteEngine(ctx, engine interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteEngine", reflect.TypeOf((*MockSecretManager)(nil).DeleteEngine), ctx, engine)
}

// ExistEngine mocks base method.
func (m *MockSecretManager) ExistEngine(ctx context.Context, engine string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExistEngine", ctx, engine)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExistEngine indicates an expected call of ExistEngine.
func (mr *MockSecretManagerMockRecorder) ExistEngine(ctx, engine interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExistEngine", reflect.TypeOf((*MockSecretManager)(nil).ExistEngine), ctx, engine)
}

// Put mocks base method.
func (m *MockSecretManager) Put(ctx context.Context, engine, secretPath string, data map[string]interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Put", ctx, engine, secretPath, data)
	ret0, _ := ret[0].(error)
	return ret0
}

// Put indicates an expected call of Put.
func (mr *MockSecretManagerMockRecorder) Put(ctx, engine, secretPath, data interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Put", reflect.TypeOf((*MockSecretManager)(nil).Put), ctx, engine, secretPath, data)
}
