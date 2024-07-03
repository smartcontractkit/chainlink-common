// Code generated by mockery v2.43.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// RemoteCodec is an autogenerated mock type for the RemoteCodec type
type RemoteCodec struct {
	mock.Mock
}

// CreateType provides a mock function with given fields: itemType, forEncoding
func (_m *RemoteCodec) CreateType(itemType string, forEncoding bool) (interface{}, error) {
	ret := _m.Called(itemType, forEncoding)

	if len(ret) == 0 {
		panic("no return value specified for CreateType")
	}

	var r0 interface{}
	var r1 error
	if rf, ok := ret.Get(0).(func(string, bool) (interface{}, error)); ok {
		return rf(itemType, forEncoding)
	}
	if rf, ok := ret.Get(0).(func(string, bool) interface{}); ok {
		r0 = rf(itemType, forEncoding)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
	}

	if rf, ok := ret.Get(1).(func(string, bool) error); ok {
		r1 = rf(itemType, forEncoding)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Decode provides a mock function with given fields: ctx, raw, into, itemType
func (_m *RemoteCodec) Decode(ctx context.Context, raw []byte, into interface{}, itemType string) error {
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

// Encode provides a mock function with given fields: ctx, item, itemType
func (_m *RemoteCodec) Encode(ctx context.Context, item interface{}, itemType string) ([]byte, error) {
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

// GetMaxDecodingSize provides a mock function with given fields: ctx, n, itemType
func (_m *RemoteCodec) GetMaxDecodingSize(ctx context.Context, n int, itemType string) (int, error) {
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

// GetMaxEncodingSize provides a mock function with given fields: ctx, n, itemType
func (_m *RemoteCodec) GetMaxEncodingSize(ctx context.Context, n int, itemType string) (int, error) {
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

// NewRemoteCodec creates a new instance of RemoteCodec. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRemoteCodec(t interface {
	mock.TestingT
	Cleanup(func())
}) *RemoteCodec {
	mock := &RemoteCodec{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}