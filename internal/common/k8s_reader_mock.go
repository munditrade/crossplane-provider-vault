package common

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// MockK8sReader is a mock of K8sReader interface.
type MockK8sReader struct {
	ctrl     *gomock.Controller
	recorder *MockK8sReaderMockRecorder
}

// MockK8sReaderMockRecorder is the mock recorder for MockK8sReader.
type MockK8sReaderMockRecorder struct {
	mock *MockK8sReader
}

// NewMockK8sReader creates a new mock instance.
func NewMockK8sReader(ctrl *gomock.Controller) *MockK8sReader {
	mock := &MockK8sReader{ctrl: ctrl}
	mock.recorder = &MockK8sReaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockK8sReader) EXPECT() *MockK8sReaderMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockK8sReader) Get(ctx context.Context, key client.ObjectKey, obj client.Object) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, key, obj)
	ret0, _ := ret[0].(error)
	return ret0
}

// Get indicates an expected call of Get.
func (mr *MockK8sReaderMockRecorder) Get(ctx, key, obj interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockK8sReader)(nil).Get), ctx, key, obj)
}

// List mocks base method.
func (m *MockK8sReader) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{ctx, list}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "List", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// List indicates an expected call of List.
func (mr *MockK8sReaderMockRecorder) List(ctx, list interface{}, opts ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{ctx, list}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockK8sReader)(nil).List), varargs...)
}
