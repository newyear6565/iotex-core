// Code generated by MockGen. DO NOT EDIT.
// Source: ./api/responder.go

// Package mock_apiresponder is a generated GoMock package.
package mock_apiresponder

import (
	gomock "github.com/golang/mock/gomock"
	block "github.com/iotexproject/iotex-core/blockchain/block"
	reflect "reflect"
)

// MockResponder is a mock of Responder interface
type MockResponder struct {
	ctrl     *gomock.Controller
	recorder *MockResponderMockRecorder
}

// MockResponderMockRecorder is the mock recorder for MockResponder
type MockResponderMockRecorder struct {
	mock *MockResponder
}

// NewMockResponder creates a new mock instance
func NewMockResponder(ctrl *gomock.Controller) *MockResponder {
	mock := &MockResponder{ctrl: ctrl}
	mock.recorder = &MockResponderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockResponder) EXPECT() *MockResponderMockRecorder {
	return m.recorder
}

// Respond mocks base method
func (m *MockResponder) Respond(arg0 *block.Block) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Respond", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Respond indicates an expected call of Respond
func (mr *MockResponderMockRecorder) Respond(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Respond", reflect.TypeOf((*MockResponder)(nil).Respond), arg0)
}

// Exit mocks base method
func (m *MockResponder) Exit() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Exit")
}

// Exit indicates an expected call of Exit
func (mr *MockResponderMockRecorder) Exit() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Exit", reflect.TypeOf((*MockResponder)(nil).Exit))
}
