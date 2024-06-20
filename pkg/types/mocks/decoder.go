// Code generated by mockery v2.43.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// Decoder is an autogenerated mock type for the Decoder type
type Decoder struct {
	mock.Mock
}

// Decode provides a mock function with given fields: ctx, raw, into, itemType
func (_m *Decoder) Decode(ctx context.Context, raw []byte, into interface{}, itemType string) error {
	ret := _m.Called(ctx, raw, into, itemType)

	if len(ret) == 0 {
		panic("no return value specified for Decode")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []byte, interface{}, string) error); ok {
		r0 = rf(ctx, raw, into, itemType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetMaxDecodingSize provides a mock function with given fields: ctx, n, itemType
func (_m *Decoder) GetMaxDecodingSize(ctx context.Context, n int, itemType string) (int, error) {
	ret := _m.Called(ctx, n, itemType)

	if len(ret) == 0 {
		panic("no return value specified for GetMaxDecodingSize")
	}

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int, string) (int, error)); ok {
		return rf(ctx, n, itemType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int, string) int); ok {
		r0 = rf(ctx, n, itemType)
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func(context.Context, int, string) error); ok {
		r1 = rf(ctx, n, itemType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewDecoder creates a new instance of Decoder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDecoder(t interface {
	mock.TestingT
	Cleanup(func())
}) *Decoder {
	mock := &Decoder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
