// Code generated by mockery v2.38.0. DO NOT EDIT.

package mocks

import (
	mercury "github.com/smartcontractkit/chainlink-common/pkg/capabilities/mercury"
	mock "github.com/stretchr/testify/mock"

	values "github.com/smartcontractkit/chainlink-common/pkg/values"
)

// MercuryCodec is an autogenerated mock type for the MercuryCodec type
type MercuryCodec struct {
	mock.Mock
}

// Unwrap provides a mock function with given fields: raw
func (_m *MercuryCodec) Unwrap(raw values.Value) (mercury.ReportSet, error) {
	ret := _m.Called(raw)

	if len(ret) == 0 {
		panic("no return value specified for Unwrap")
	}

	var r0 mercury.ReportSet
	var r1 error
	if rf, ok := ret.Get(0).(func(values.Value) (mercury.ReportSet, error)); ok {
		return rf(raw)
	}
	if rf, ok := ret.Get(0).(func(values.Value) mercury.ReportSet); ok {
		r0 = rf(raw)
	} else {
		r0 = ret.Get(0).(mercury.ReportSet)
	}

	if rf, ok := ret.Get(1).(func(values.Value) error); ok {
		r1 = rf(raw)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Wrap provides a mock function with given fields: reportSet
func (_m *MercuryCodec) Wrap(reportSet mercury.ReportSet) (values.Value, error) {
	ret := _m.Called(reportSet)

	if len(ret) == 0 {
		panic("no return value specified for Wrap")
	}

	var r0 values.Value
	var r1 error
	if rf, ok := ret.Get(0).(func(mercury.ReportSet) (values.Value, error)); ok {
		return rf(reportSet)
	}
	if rf, ok := ret.Get(0).(func(mercury.ReportSet) values.Value); ok {
		r0 = rf(reportSet)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(values.Value)
		}
	}

	if rf, ok := ret.Get(1).(func(mercury.ReportSet) error); ok {
		r1 = rf(reportSet)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewMercuryCodec creates a new instance of MercuryCodec. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMercuryCodec(t interface {
	mock.TestingT
	Cleanup(func())
}) *MercuryCodec {
	mock := &MercuryCodec{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
