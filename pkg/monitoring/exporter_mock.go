// Code generated by mockery v2.53.3. DO NOT EDIT.

package monitoring

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// ExporterMock is an autogenerated mock type for the Exporter type
type ExporterMock struct {
	mock.Mock
}

type ExporterMock_Expecter struct {
	mock *mock.Mock
}

func (_m *ExporterMock) EXPECT() *ExporterMock_Expecter {
	return &ExporterMock_Expecter{mock: &_m.Mock}
}

// Cleanup provides a mock function with given fields: ctx
func (_m *ExporterMock) Cleanup(ctx context.Context) {
	_m.Called(ctx)
}

// ExporterMock_Cleanup_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Cleanup'
type ExporterMock_Cleanup_Call struct {
	*mock.Call
}

// Cleanup is a helper method to define mock.On call
//   - ctx context.Context
func (_e *ExporterMock_Expecter) Cleanup(ctx interface{}) *ExporterMock_Cleanup_Call {
	return &ExporterMock_Cleanup_Call{Call: _e.mock.On("Cleanup", ctx)}
}

func (_c *ExporterMock_Cleanup_Call) Run(run func(ctx context.Context)) *ExporterMock_Cleanup_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *ExporterMock_Cleanup_Call) Return() *ExporterMock_Cleanup_Call {
	_c.Call.Return()
	return _c
}

func (_c *ExporterMock_Cleanup_Call) RunAndReturn(run func(context.Context)) *ExporterMock_Cleanup_Call {
	_c.Run(run)
	return _c
}

// Export provides a mock function with given fields: ctx, data
func (_m *ExporterMock) Export(ctx context.Context, data interface{}) {
	_m.Called(ctx, data)
}

// ExporterMock_Export_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Export'
type ExporterMock_Export_Call struct {
	*mock.Call
}

// Export is a helper method to define mock.On call
//   - ctx context.Context
//   - data interface{}
func (_e *ExporterMock_Expecter) Export(ctx interface{}, data interface{}) *ExporterMock_Export_Call {
	return &ExporterMock_Export_Call{Call: _e.mock.On("Export", ctx, data)}
}

func (_c *ExporterMock_Export_Call) Run(run func(ctx context.Context, data interface{})) *ExporterMock_Export_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(interface{}))
	})
	return _c
}

func (_c *ExporterMock_Export_Call) Return() *ExporterMock_Export_Call {
	_c.Call.Return()
	return _c
}

func (_c *ExporterMock_Export_Call) RunAndReturn(run func(context.Context, interface{})) *ExporterMock_Export_Call {
	_c.Run(run)
	return _c
}

// NewExporterMock creates a new instance of ExporterMock. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewExporterMock(t interface {
	mock.TestingT
	Cleanup(func())
}) *ExporterMock {
	mock := &ExporterMock{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
