// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/andig/ulm/api (interfaces: Charger,ChargeController,Meter)

// Package mock_api is a generated GoMock package.
package mock_api

import (
	api "github.com/andig/ulm/api"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockCharger is a mock of Charger interface
type MockCharger struct {
	ctrl     *gomock.Controller
	recorder *MockChargerMockRecorder
}

// MockChargerMockRecorder is the mock recorder for MockCharger
type MockChargerMockRecorder struct {
	mock *MockCharger
}

// NewMockCharger creates a new mock instance
func NewMockCharger(ctrl *gomock.Controller) *MockCharger {
	mock := &MockCharger{ctrl: ctrl}
	mock.recorder = &MockChargerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockCharger) EXPECT() *MockChargerMockRecorder {
	return m.recorder
}

// ActualCurrent mocks base method
func (m *MockCharger) ActualCurrent() (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ActualCurrent")
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ActualCurrent indicates an expected call of ActualCurrent
func (mr *MockChargerMockRecorder) ActualCurrent() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ActualCurrent", reflect.TypeOf((*MockCharger)(nil).ActualCurrent))
}

// Enable mocks base method
func (m *MockCharger) Enable(arg0 bool) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Enable", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// Enable indicates an expected call of Enable
func (mr *MockChargerMockRecorder) Enable(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Enable", reflect.TypeOf((*MockCharger)(nil).Enable), arg0)
}

// Enabled mocks base method
func (m *MockCharger) Enabled() (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Enabled")
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Enabled indicates an expected call of Enabled
func (mr *MockChargerMockRecorder) Enabled() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Enabled", reflect.TypeOf((*MockCharger)(nil).Enabled))
}

// Status mocks base method
func (m *MockCharger) Status() (api.ChargeStatus, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Status")
	ret0, _ := ret[0].(api.ChargeStatus)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Status indicates an expected call of Status
func (mr *MockChargerMockRecorder) Status() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Status", reflect.TypeOf((*MockCharger)(nil).Status))
}

// MockChargeController is a mock of ChargeController interface
type MockChargeController struct {
	ctrl     *gomock.Controller
	recorder *MockChargeControllerMockRecorder
}

// MockChargeControllerMockRecorder is the mock recorder for MockChargeController
type MockChargeControllerMockRecorder struct {
	mock *MockChargeController
}

// NewMockChargeController creates a new mock instance
func NewMockChargeController(ctrl *gomock.Controller) *MockChargeController {
	mock := &MockChargeController{ctrl: ctrl}
	mock.recorder = &MockChargeControllerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockChargeController) EXPECT() *MockChargeControllerMockRecorder {
	return m.recorder
}

// MaxCurrent mocks base method
func (m *MockChargeController) MaxCurrent(arg0 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MaxCurrent", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// MaxCurrent indicates an expected call of MaxCurrent
func (mr *MockChargeControllerMockRecorder) MaxCurrent(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MaxCurrent", reflect.TypeOf((*MockChargeController)(nil).MaxCurrent), arg0)
}

// MockMeter is a mock of Meter interface
type MockMeter struct {
	ctrl     *gomock.Controller
	recorder *MockMeterMockRecorder
}

// MockMeterMockRecorder is the mock recorder for MockMeter
type MockMeterMockRecorder struct {
	mock *MockMeter
}

// NewMockMeter creates a new mock instance
func NewMockMeter(ctrl *gomock.Controller) *MockMeter {
	mock := &MockMeter{ctrl: ctrl}
	mock.recorder = &MockMeterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMeter) EXPECT() *MockMeterMockRecorder {
	return m.recorder
}

// CurrentPower mocks base method
func (m *MockMeter) CurrentPower() (float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CurrentPower")
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CurrentPower indicates an expected call of CurrentPower
func (mr *MockMeterMockRecorder) CurrentPower() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CurrentPower", reflect.TypeOf((*MockMeter)(nil).CurrentPower))
}
