// Code generated by mockery v2.43.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// Encoder is an autogenerated mock type for the Encoder type
type Encoder struct {
	mock.Mock
}

// Encode provides a mock function with given fields: ctx, item, itemType
func (_m *Encoder) Encode(ctx context.Context, item interface{}, itemType string) ([]byte, error) {
	ret := _m.Called(ctx, item, itemType)

	if len(ret) == 0 {
		panic("no return value specified for Encode")
	}

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, string) ([]byte, error)); ok {
		return rf(ctx, item, itemType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, string) []byte); ok {
		r0 = rf(ctx, item, itemType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, interface{}, string) error); ok {
		r1 = rf(ctx, item, itemType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMaxEncodingSize provides a mock function with given fields: ctx, n, itemType
func (_m *Encoder) GetMaxEncodingSize(ctx context.Context, n int, itemType string) (int, error) {
	ret := _m.Called(ctx, n, itemType)

	if len(ret) == 0 {
		panic("no return value specified for GetMaxEncodingSize")
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

// NewEncoder creates a new instance of Encoder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEncoder(t interface {
	mock.TestingT
	Cleanup(func())
}) *Encoder {
	mock := &Encoder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
