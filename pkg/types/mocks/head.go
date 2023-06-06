// Code generated by mockery v2.28.1. DO NOT EDIT.

package mocks

import (
	big "math/big"

	types "github.com/smartcontractkit/chainlink-relay/pkg/types"
	mock "github.com/stretchr/testify/mock"
)

// Head is an autogenerated mock type for the Head type
type Head[BLOCK_HASH types.Hashable] struct {
	mock.Mock
}

// BlockHash provides a mock function with given fields:
func (_m *Head[BLOCK_HASH]) BlockHash() BLOCK_HASH {
	ret := _m.Called()

	var r0 BLOCK_HASH
	if rf, ok := ret.Get(0).(func() BLOCK_HASH); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(BLOCK_HASH)
	}

	return r0
}

// BlockNumber provides a mock function with given fields:
func (_m *Head[BLOCK_HASH]) BlockNumber() int64 {
	ret := _m.Called()

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	return r0
}

// ChainLength provides a mock function with given fields:
func (_m *Head[BLOCK_HASH]) ChainLength() uint32 {
	ret := _m.Called()

	var r0 uint32
	if rf, ok := ret.Get(0).(func() uint32); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint32)
	}

	return r0
}

// EarliestHeadInChain provides a mock function with given fields:
func (_m *Head[BLOCK_HASH]) EarliestHeadInChain() types.Head[BLOCK_HASH] {
	ret := _m.Called()

	var r0 types.Head[BLOCK_HASH]
	if rf, ok := ret.Get(0).(func() types.Head[BLOCK_HASH]); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(types.Head[BLOCK_HASH])
		}
	}

	return r0
}

// GetParent provides a mock function with given fields:
func (_m *Head[BLOCK_HASH]) GetParent() types.Head[BLOCK_HASH] {
	ret := _m.Called()

	var r0 types.Head[BLOCK_HASH]
	if rf, ok := ret.Get(0).(func() types.Head[BLOCK_HASH]); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(types.Head[BLOCK_HASH])
		}
	}

	return r0
}

// GetParentHash provides a mock function with given fields:
func (_m *Head[BLOCK_HASH]) GetParentHash() BLOCK_HASH {
	ret := _m.Called()

	var r0 BLOCK_HASH
	if rf, ok := ret.Get(0).(func() BLOCK_HASH); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(BLOCK_HASH)
	}

	return r0
}

// HashAtHeight provides a mock function with given fields: blockNum
func (_m *Head[BLOCK_HASH]) HashAtHeight(blockNum int64) BLOCK_HASH {
	ret := _m.Called(blockNum)

	var r0 BLOCK_HASH
	if rf, ok := ret.Get(0).(func(int64) BLOCK_HASH); ok {
		r0 = rf(blockNum)
	} else {
		r0 = ret.Get(0).(BLOCK_HASH)
	}

	return r0
}

// ToInt provides a mock function with given fields:
func (_m *Head[BLOCK_HASH]) ToInt() *big.Int {
	ret := _m.Called()

	var r0 *big.Int
	if rf, ok := ret.Get(0).(func() *big.Int); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*big.Int)
		}
	}

	return r0
}

type mockConstructorTestingTNewHead interface {
	mock.TestingT
	Cleanup(func())
}

// NewHead creates a new instance of Head. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewHead[BLOCK_HASH types.Hashable](t mockConstructorTestingTNewHead) *Head[BLOCK_HASH] {
	mock := &Head[BLOCK_HASH]{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
