// Code generated by mockery v2.43.2. DO NOT EDIT.

package mocks

import (
	context "context"

	core "github.com/smartcontractkit/chainlink-common/pkg/types/core"
	mock "github.com/stretchr/testify/mock"

	types "github.com/smartcontractkit/chainlink-common/pkg/types"
)

// Relayer is an autogenerated mock type for the Relayer type
type Relayer struct {
	mock.Mock
}

type Relayer_Expecter struct {
	mock *mock.Mock
}

func (_m *Relayer) EXPECT() *Relayer_Expecter {
	return &Relayer_Expecter{mock: &_m.Mock}
}

// Close provides a mock function with given fields:
func (_m *Relayer) Close() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Close")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Relayer_Close_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Close'
type Relayer_Close_Call struct {
	*mock.Call
}

// Close is a helper method to define mock.On call
func (_e *Relayer_Expecter) Close() *Relayer_Close_Call {
	return &Relayer_Close_Call{Call: _e.mock.On("Close")}
}

func (_c *Relayer_Close_Call) Run(run func()) *Relayer_Close_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Relayer_Close_Call) Return(_a0 error) *Relayer_Close_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Relayer_Close_Call) RunAndReturn(run func() error) *Relayer_Close_Call {
	_c.Call.Return(run)
	return _c
}

// HealthReport provides a mock function with given fields:
func (_m *Relayer) HealthReport() map[string]error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for HealthReport")
	}

	var r0 map[string]error
	if rf, ok := ret.Get(0).(func() map[string]error); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]error)
		}
	}

	return r0
}

// Relayer_HealthReport_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HealthReport'
type Relayer_HealthReport_Call struct {
	*mock.Call
}

// HealthReport is a helper method to define mock.On call
func (_e *Relayer_Expecter) HealthReport() *Relayer_HealthReport_Call {
	return &Relayer_HealthReport_Call{Call: _e.mock.On("HealthReport")}
}

func (_c *Relayer_HealthReport_Call) Run(run func()) *Relayer_HealthReport_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Relayer_HealthReport_Call) Return(_a0 map[string]error) *Relayer_HealthReport_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Relayer_HealthReport_Call) RunAndReturn(run func() map[string]error) *Relayer_HealthReport_Call {
	_c.Call.Return(run)
	return _c
}

// Name provides a mock function with given fields:
func (_m *Relayer) Name() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Name")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Relayer_Name_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Name'
type Relayer_Name_Call struct {
	*mock.Call
}

// Name is a helper method to define mock.On call
func (_e *Relayer_Expecter) Name() *Relayer_Name_Call {
	return &Relayer_Name_Call{Call: _e.mock.On("Name")}
}

func (_c *Relayer_Name_Call) Run(run func()) *Relayer_Name_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Relayer_Name_Call) Return(_a0 string) *Relayer_Name_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Relayer_Name_Call) RunAndReturn(run func() string) *Relayer_Name_Call {
	_c.Call.Return(run)
	return _c
}

// NewChainWriter provides a mock function with given fields: _a0, chainWriterConfig
func (_m *Relayer) NewChainWriter(_a0 context.Context, chainWriterConfig []byte) (types.ChainWriter, error) {
	ret := _m.Called(_a0, chainWriterConfig)

	if len(ret) == 0 {
		panic("no return value specified for NewChainWriter")
	}

	var r0 types.ChainWriter
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []byte) (types.ChainWriter, error)); ok {
		return rf(_a0, chainWriterConfig)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []byte) types.ChainWriter); ok {
		r0 = rf(_a0, chainWriterConfig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(types.ChainWriter)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []byte) error); ok {
		r1 = rf(_a0, chainWriterConfig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Relayer_NewChainWriter_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NewChainWriter'
type Relayer_NewChainWriter_Call struct {
	*mock.Call
}

// NewChainWriter is a helper method to define mock.On call
//   - _a0 context.Context
//   - chainWriterConfig []byte
func (_e *Relayer_Expecter) NewChainWriter(_a0 interface{}, chainWriterConfig interface{}) *Relayer_NewChainWriter_Call {
	return &Relayer_NewChainWriter_Call{Call: _e.mock.On("NewChainWriter", _a0, chainWriterConfig)}
}

func (_c *Relayer_NewChainWriter_Call) Run(run func(_a0 context.Context, chainWriterConfig []byte)) *Relayer_NewChainWriter_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]byte))
	})
	return _c
}

func (_c *Relayer_NewChainWriter_Call) Return(_a0 types.ChainWriter, _a1 error) *Relayer_NewChainWriter_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Relayer_NewChainWriter_Call) RunAndReturn(run func(context.Context, []byte) (types.ChainWriter, error)) *Relayer_NewChainWriter_Call {
	_c.Call.Return(run)
	return _c
}

// NewContractReader provides a mock function with given fields: _a0, contractReaderConfig
func (_m *Relayer) NewContractReader(_a0 context.Context, contractReaderConfig []byte) (types.ContractReader, error) {
	ret := _m.Called(_a0, contractReaderConfig)

	if len(ret) == 0 {
		panic("no return value specified for NewContractReader")
	}

	var r0 types.ContractReader
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []byte) (types.ContractReader, error)); ok {
		return rf(_a0, contractReaderConfig)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []byte) types.ContractReader); ok {
		r0 = rf(_a0, contractReaderConfig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(types.ContractReader)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []byte) error); ok {
		r1 = rf(_a0, contractReaderConfig)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Relayer_NewContractReader_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NewContractReader'
type Relayer_NewContractReader_Call struct {
	*mock.Call
}

// NewContractReader is a helper method to define mock.On call
//   - _a0 context.Context
//   - contractReaderConfig []byte
func (_e *Relayer_Expecter) NewContractReader(_a0 interface{}, contractReaderConfig interface{}) *Relayer_NewContractReader_Call {
	return &Relayer_NewContractReader_Call{Call: _e.mock.On("NewContractReader", _a0, contractReaderConfig)}
}

func (_c *Relayer_NewContractReader_Call) Run(run func(_a0 context.Context, contractReaderConfig []byte)) *Relayer_NewContractReader_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]byte))
	})
	return _c
}

func (_c *Relayer_NewContractReader_Call) Return(_a0 types.ContractReader, _a1 error) *Relayer_NewContractReader_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Relayer_NewContractReader_Call) RunAndReturn(run func(context.Context, []byte) (types.ContractReader, error)) *Relayer_NewContractReader_Call {
	_c.Call.Return(run)
	return _c
}

// NewPluginProvider provides a mock function with given fields: _a0, _a1, _a2
func (_m *Relayer) NewPluginProvider(_a0 context.Context, _a1 core.RelayArgs, _a2 core.PluginArgs) (types.PluginProvider, error) {
	ret := _m.Called(_a0, _a1, _a2)

	if len(ret) == 0 {
		panic("no return value specified for NewPluginProvider")
	}

	var r0 types.PluginProvider
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, core.RelayArgs, core.PluginArgs) (types.PluginProvider, error)); ok {
		return rf(_a0, _a1, _a2)
	}
	if rf, ok := ret.Get(0).(func(context.Context, core.RelayArgs, core.PluginArgs) types.PluginProvider); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(types.PluginProvider)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, core.RelayArgs, core.PluginArgs) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Relayer_NewPluginProvider_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'NewPluginProvider'
type Relayer_NewPluginProvider_Call struct {
	*mock.Call
}

// NewPluginProvider is a helper method to define mock.On call
//   - _a0 context.Context
//   - _a1 core.RelayArgs
//   - _a2 core.PluginArgs
func (_e *Relayer_Expecter) NewPluginProvider(_a0 interface{}, _a1 interface{}, _a2 interface{}) *Relayer_NewPluginProvider_Call {
	return &Relayer_NewPluginProvider_Call{Call: _e.mock.On("NewPluginProvider", _a0, _a1, _a2)}
}

func (_c *Relayer_NewPluginProvider_Call) Run(run func(_a0 context.Context, _a1 core.RelayArgs, _a2 core.PluginArgs)) *Relayer_NewPluginProvider_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(core.RelayArgs), args[2].(core.PluginArgs))
	})
	return _c
}

func (_c *Relayer_NewPluginProvider_Call) Return(_a0 types.PluginProvider, _a1 error) *Relayer_NewPluginProvider_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Relayer_NewPluginProvider_Call) RunAndReturn(run func(context.Context, core.RelayArgs, core.PluginArgs) (types.PluginProvider, error)) *Relayer_NewPluginProvider_Call {
	_c.Call.Return(run)
	return _c
}

// Ready provides a mock function with given fields:
func (_m *Relayer) Ready() error {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Ready")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Relayer_Ready_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Ready'
type Relayer_Ready_Call struct {
	*mock.Call
}

// Ready is a helper method to define mock.On call
func (_e *Relayer_Expecter) Ready() *Relayer_Ready_Call {
	return &Relayer_Ready_Call{Call: _e.mock.On("Ready")}
}

func (_c *Relayer_Ready_Call) Run(run func()) *Relayer_Ready_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *Relayer_Ready_Call) Return(_a0 error) *Relayer_Ready_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Relayer_Ready_Call) RunAndReturn(run func() error) *Relayer_Ready_Call {
	_c.Call.Return(run)
	return _c
}

// Start provides a mock function with given fields: _a0
func (_m *Relayer) Start(_a0 context.Context) error {
	ret := _m.Called(_a0)

	if len(ret) == 0 {
		panic("no return value specified for Start")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Relayer_Start_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Start'
type Relayer_Start_Call struct {
	*mock.Call
}

// Start is a helper method to define mock.On call
//   - _a0 context.Context
func (_e *Relayer_Expecter) Start(_a0 interface{}) *Relayer_Start_Call {
	return &Relayer_Start_Call{Call: _e.mock.On("Start", _a0)}
}

func (_c *Relayer_Start_Call) Run(run func(_a0 context.Context)) *Relayer_Start_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *Relayer_Start_Call) Return(_a0 error) *Relayer_Start_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Relayer_Start_Call) RunAndReturn(run func(context.Context) error) *Relayer_Start_Call {
	_c.Call.Return(run)
	return _c
}

// NewRelayer creates a new instance of Relayer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRelayer(t interface {
	mock.TestingT
	Cleanup(func())
}) *Relayer {
	mock := &Relayer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
