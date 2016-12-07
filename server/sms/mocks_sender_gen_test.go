// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/smancke/guble/server/sms (interfaces: Sender)

package sms

import (
	gomock "github.com/golang/mock/gomock"
	protocol "github.com/smancke/guble/protocol"
)

// Mock of Sender interface
type MockSender struct {
	ctrl     *gomock.Controller
	recorder *_MockSenderRecorder
}

// Recorder for MockSender (not exported)
type _MockSenderRecorder struct {
	mock *MockSender
}

func NewMockSender(ctrl *gomock.Controller) *MockSender {
	mock := &MockSender{ctrl: ctrl}
	mock.recorder = &_MockSenderRecorder{mock}
	return mock
}

func (_m *MockSender) EXPECT() *_MockSenderRecorder {
	return _m.recorder
}

func (_m *MockSender) Send(_param0 *protocol.Message) error {
	ret := _m.ctrl.Call(_m, "Send", _param0)
	ret0, _ := ret[0].(error)
	return ret0
}

func (_mr *_MockSenderRecorder) Send(arg0 interface{}) *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Send", arg0)
}