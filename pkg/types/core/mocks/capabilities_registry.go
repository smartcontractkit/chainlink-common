// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	context "context"

	capabilities "github.com/smartcontractkit/chainlink-common/pkg/capabilities"

	mock "github.com/stretchr/testify/mock"
)

// CapabilitiesRegistry is an autogenerated mock type for the CapabilitiesRegistry type
type CapabilitiesRegistry struct {
	mock.Mock
}

type CapabilitiesRegistry_Expecter struct {
	mock *mock.Mock
}

func (_m *CapabilitiesRegistry) EXPECT() *CapabilitiesRegistry_Expecter {
	return &CapabilitiesRegistry_Expecter{mock: &_m.Mock}
}

// Add provides a mock function with given fields: ctx, c
func (_m *CapabilitiesRegistry) Add(ctx context.Context, c capabilities.BaseCapability) error {
	ret := _m.Called(ctx, c)

	if len(ret) == 0 {
		panic("no return value specified for Add")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, capabilities.BaseCapability) error); ok {
		r0 = rf(ctx, c)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CapabilitiesRegistry_Add_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Add'
type CapabilitiesRegistry_Add_Call struct {
	*mock.Call
}

// Add is a helper method to define mock.On call
//   - ctx context.Context
//   - c capabilities.BaseCapability
func (_e *CapabilitiesRegistry_Expecter) Add(ctx interface{}, c interface{}) *CapabilitiesRegistry_Add_Call {
	return &CapabilitiesRegistry_Add_Call{Call: _e.mock.On("Add", ctx, c)}
}

func (_c *CapabilitiesRegistry_Add_Call) Run(run func(ctx context.Context, c capabilities.BaseCapability)) *CapabilitiesRegistry_Add_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(capabilities.BaseCapability))
	})
	return _c
}

func (_c *CapabilitiesRegistry_Add_Call) Return(_a0 error) *CapabilitiesRegistry_Add_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *CapabilitiesRegistry_Add_Call) RunAndReturn(run func(context.Context, capabilities.BaseCapability) error) *CapabilitiesRegistry_Add_Call {
	_c.Call.Return(run)
	return _c
}

// ConfigForCapability provides a mock function with given fields: ctx, capabilityID, donID
func (_m *CapabilitiesRegistry) ConfigForCapability(ctx context.Context, capabilityID string, donID uint32) (capabilities.CapabilityConfiguration, error) {
	ret := _m.Called(ctx, capabilityID, donID)

	if len(ret) == 0 {
		panic("no return value specified for ConfigForCapability")
	}

	var r0 capabilities.CapabilityConfiguration
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, uint32) (capabilities.CapabilityConfiguration, error)); ok {
		return rf(ctx, capabilityID, donID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, uint32) capabilities.CapabilityConfiguration); ok {
		r0 = rf(ctx, capabilityID, donID)
	} else {
		r0 = ret.Get(0).(capabilities.CapabilityConfiguration)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, uint32) error); ok {
		r1 = rf(ctx, capabilityID, donID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CapabilitiesRegistry_ConfigForCapability_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ConfigForCapability'
type CapabilitiesRegistry_ConfigForCapability_Call struct {
	*mock.Call
}

// ConfigForCapability is a helper method to define mock.On call
//   - ctx context.Context
//   - capabilityID string
//   - donID uint32
func (_e *CapabilitiesRegistry_Expecter) ConfigForCapability(ctx interface{}, capabilityID interface{}, donID interface{}) *CapabilitiesRegistry_ConfigForCapability_Call {
	return &CapabilitiesRegistry_ConfigForCapability_Call{Call: _e.mock.On("ConfigForCapability", ctx, capabilityID, donID)}
}

func (_c *CapabilitiesRegistry_ConfigForCapability_Call) Run(run func(ctx context.Context, capabilityID string, donID uint32)) *CapabilitiesRegistry_ConfigForCapability_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(uint32))
	})
	return _c
}

func (_c *CapabilitiesRegistry_ConfigForCapability_Call) Return(_a0 capabilities.CapabilityConfiguration, _a1 error) *CapabilitiesRegistry_ConfigForCapability_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CapabilitiesRegistry_ConfigForCapability_Call) RunAndReturn(run func(context.Context, string, uint32) (capabilities.CapabilityConfiguration, error)) *CapabilitiesRegistry_ConfigForCapability_Call {
	_c.Call.Return(run)
	return _c
}

// Get provides a mock function with given fields: ctx, ID
func (_m *CapabilitiesRegistry) Get(ctx context.Context, ID string) (capabilities.BaseCapability, error) {
	ret := _m.Called(ctx, ID)

	if len(ret) == 0 {
		panic("no return value specified for Get")
	}

	var r0 capabilities.BaseCapability
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (capabilities.BaseCapability, error)); ok {
		return rf(ctx, ID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) capabilities.BaseCapability); ok {
		r0 = rf(ctx, ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(capabilities.BaseCapability)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CapabilitiesRegistry_Get_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Get'
type CapabilitiesRegistry_Get_Call struct {
	*mock.Call
}

// Get is a helper method to define mock.On call
//   - ctx context.Context
//   - ID string
func (_e *CapabilitiesRegistry_Expecter) Get(ctx interface{}, ID interface{}) *CapabilitiesRegistry_Get_Call {
	return &CapabilitiesRegistry_Get_Call{Call: _e.mock.On("Get", ctx, ID)}
}

func (_c *CapabilitiesRegistry_Get_Call) Run(run func(ctx context.Context, ID string)) *CapabilitiesRegistry_Get_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CapabilitiesRegistry_Get_Call) Return(_a0 capabilities.BaseCapability, _a1 error) *CapabilitiesRegistry_Get_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CapabilitiesRegistry_Get_Call) RunAndReturn(run func(context.Context, string) (capabilities.BaseCapability, error)) *CapabilitiesRegistry_Get_Call {
	_c.Call.Return(run)
	return _c
}

// GetAction provides a mock function with given fields: ctx, ID
func (_m *CapabilitiesRegistry) GetAction(ctx context.Context, ID string) (capabilities.ActionCapability, error) {
	ret := _m.Called(ctx, ID)

	if len(ret) == 0 {
		panic("no return value specified for GetAction")
	}

	var r0 capabilities.ActionCapability
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (capabilities.ActionCapability, error)); ok {
		return rf(ctx, ID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) capabilities.ActionCapability); ok {
		r0 = rf(ctx, ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(capabilities.ActionCapability)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CapabilitiesRegistry_GetAction_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAction'
type CapabilitiesRegistry_GetAction_Call struct {
	*mock.Call
}

// GetAction is a helper method to define mock.On call
//   - ctx context.Context
//   - ID string
func (_e *CapabilitiesRegistry_Expecter) GetAction(ctx interface{}, ID interface{}) *CapabilitiesRegistry_GetAction_Call {
	return &CapabilitiesRegistry_GetAction_Call{Call: _e.mock.On("GetAction", ctx, ID)}
}

func (_c *CapabilitiesRegistry_GetAction_Call) Run(run func(ctx context.Context, ID string)) *CapabilitiesRegistry_GetAction_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CapabilitiesRegistry_GetAction_Call) Return(_a0 capabilities.ActionCapability, _a1 error) *CapabilitiesRegistry_GetAction_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CapabilitiesRegistry_GetAction_Call) RunAndReturn(run func(context.Context, string) (capabilities.ActionCapability, error)) *CapabilitiesRegistry_GetAction_Call {
	_c.Call.Return(run)
	return _c
}

// GetConsensus provides a mock function with given fields: ctx, ID
func (_m *CapabilitiesRegistry) GetConsensus(ctx context.Context, ID string) (capabilities.ConsensusCapability, error) {
	ret := _m.Called(ctx, ID)

	if len(ret) == 0 {
		panic("no return value specified for GetConsensus")
	}

	var r0 capabilities.ConsensusCapability
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (capabilities.ConsensusCapability, error)); ok {
		return rf(ctx, ID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) capabilities.ConsensusCapability); ok {
		r0 = rf(ctx, ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(capabilities.ConsensusCapability)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CapabilitiesRegistry_GetConsensus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetConsensus'
type CapabilitiesRegistry_GetConsensus_Call struct {
	*mock.Call
}

// GetConsensus is a helper method to define mock.On call
//   - ctx context.Context
//   - ID string
func (_e *CapabilitiesRegistry_Expecter) GetConsensus(ctx interface{}, ID interface{}) *CapabilitiesRegistry_GetConsensus_Call {
	return &CapabilitiesRegistry_GetConsensus_Call{Call: _e.mock.On("GetConsensus", ctx, ID)}
}

func (_c *CapabilitiesRegistry_GetConsensus_Call) Run(run func(ctx context.Context, ID string)) *CapabilitiesRegistry_GetConsensus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CapabilitiesRegistry_GetConsensus_Call) Return(_a0 capabilities.ConsensusCapability, _a1 error) *CapabilitiesRegistry_GetConsensus_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CapabilitiesRegistry_GetConsensus_Call) RunAndReturn(run func(context.Context, string) (capabilities.ConsensusCapability, error)) *CapabilitiesRegistry_GetConsensus_Call {
	_c.Call.Return(run)
	return _c
}

// GetTarget provides a mock function with given fields: ctx, ID
func (_m *CapabilitiesRegistry) GetTarget(ctx context.Context, ID string) (capabilities.TargetCapability, error) {
	ret := _m.Called(ctx, ID)

	if len(ret) == 0 {
		panic("no return value specified for GetTarget")
	}

	var r0 capabilities.TargetCapability
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (capabilities.TargetCapability, error)); ok {
		return rf(ctx, ID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) capabilities.TargetCapability); ok {
		r0 = rf(ctx, ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(capabilities.TargetCapability)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CapabilitiesRegistry_GetTarget_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTarget'
type CapabilitiesRegistry_GetTarget_Call struct {
	*mock.Call
}

// GetTarget is a helper method to define mock.On call
//   - ctx context.Context
//   - ID string
func (_e *CapabilitiesRegistry_Expecter) GetTarget(ctx interface{}, ID interface{}) *CapabilitiesRegistry_GetTarget_Call {
	return &CapabilitiesRegistry_GetTarget_Call{Call: _e.mock.On("GetTarget", ctx, ID)}
}

func (_c *CapabilitiesRegistry_GetTarget_Call) Run(run func(ctx context.Context, ID string)) *CapabilitiesRegistry_GetTarget_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CapabilitiesRegistry_GetTarget_Call) Return(_a0 capabilities.TargetCapability, _a1 error) *CapabilitiesRegistry_GetTarget_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CapabilitiesRegistry_GetTarget_Call) RunAndReturn(run func(context.Context, string) (capabilities.TargetCapability, error)) *CapabilitiesRegistry_GetTarget_Call {
	_c.Call.Return(run)
	return _c
}

// GetTrigger provides a mock function with given fields: ctx, ID
func (_m *CapabilitiesRegistry) GetTrigger(ctx context.Context, ID string) (capabilities.TriggerCapability, error) {
	ret := _m.Called(ctx, ID)

	if len(ret) == 0 {
		panic("no return value specified for GetTrigger")
	}

	var r0 capabilities.TriggerCapability
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (capabilities.TriggerCapability, error)); ok {
		return rf(ctx, ID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) capabilities.TriggerCapability); ok {
		r0 = rf(ctx, ID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(capabilities.TriggerCapability)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, ID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CapabilitiesRegistry_GetTrigger_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTrigger'
type CapabilitiesRegistry_GetTrigger_Call struct {
	*mock.Call
}

// GetTrigger is a helper method to define mock.On call
//   - ctx context.Context
//   - ID string
func (_e *CapabilitiesRegistry_Expecter) GetTrigger(ctx interface{}, ID interface{}) *CapabilitiesRegistry_GetTrigger_Call {
	return &CapabilitiesRegistry_GetTrigger_Call{Call: _e.mock.On("GetTrigger", ctx, ID)}
}

func (_c *CapabilitiesRegistry_GetTrigger_Call) Run(run func(ctx context.Context, ID string)) *CapabilitiesRegistry_GetTrigger_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CapabilitiesRegistry_GetTrigger_Call) Return(_a0 capabilities.TriggerCapability, _a1 error) *CapabilitiesRegistry_GetTrigger_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CapabilitiesRegistry_GetTrigger_Call) RunAndReturn(run func(context.Context, string) (capabilities.TriggerCapability, error)) *CapabilitiesRegistry_GetTrigger_Call {
	_c.Call.Return(run)
	return _c
}

// List provides a mock function with given fields: ctx
func (_m *CapabilitiesRegistry) List(ctx context.Context) ([]capabilities.BaseCapability, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var r0 []capabilities.BaseCapability
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]capabilities.BaseCapability, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []capabilities.BaseCapability); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]capabilities.BaseCapability)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CapabilitiesRegistry_List_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'List'
type CapabilitiesRegistry_List_Call struct {
	*mock.Call
}

// List is a helper method to define mock.On call
//   - ctx context.Context
func (_e *CapabilitiesRegistry_Expecter) List(ctx interface{}) *CapabilitiesRegistry_List_Call {
	return &CapabilitiesRegistry_List_Call{Call: _e.mock.On("List", ctx)}
}

func (_c *CapabilitiesRegistry_List_Call) Run(run func(ctx context.Context)) *CapabilitiesRegistry_List_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *CapabilitiesRegistry_List_Call) Return(_a0 []capabilities.BaseCapability, _a1 error) *CapabilitiesRegistry_List_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CapabilitiesRegistry_List_Call) RunAndReturn(run func(context.Context) ([]capabilities.BaseCapability, error)) *CapabilitiesRegistry_List_Call {
	_c.Call.Return(run)
	return _c
}

// LocalNode provides a mock function with given fields: ctx
func (_m *CapabilitiesRegistry) LocalNode(ctx context.Context) (capabilities.Node, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for LocalNode")
	}

	var r0 capabilities.Node
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) (capabilities.Node, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) capabilities.Node); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Get(0).(capabilities.Node)
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CapabilitiesRegistry_LocalNode_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'LocalNode'
type CapabilitiesRegistry_LocalNode_Call struct {
	*mock.Call
}

// LocalNode is a helper method to define mock.On call
//   - ctx context.Context
func (_e *CapabilitiesRegistry_Expecter) LocalNode(ctx interface{}) *CapabilitiesRegistry_LocalNode_Call {
	return &CapabilitiesRegistry_LocalNode_Call{Call: _e.mock.On("LocalNode", ctx)}
}

func (_c *CapabilitiesRegistry_LocalNode_Call) Run(run func(ctx context.Context)) *CapabilitiesRegistry_LocalNode_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *CapabilitiesRegistry_LocalNode_Call) Return(_a0 capabilities.Node, _a1 error) *CapabilitiesRegistry_LocalNode_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *CapabilitiesRegistry_LocalNode_Call) RunAndReturn(run func(context.Context) (capabilities.Node, error)) *CapabilitiesRegistry_LocalNode_Call {
	_c.Call.Return(run)
	return _c
}

// Remove provides a mock function with given fields: ctx, ID
func (_m *CapabilitiesRegistry) Remove(ctx context.Context, ID string) error {
	ret := _m.Called(ctx, ID)

	if len(ret) == 0 {
		panic("no return value specified for Remove")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, ID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CapabilitiesRegistry_Remove_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Remove'
type CapabilitiesRegistry_Remove_Call struct {
	*mock.Call
}

// Remove is a helper method to define mock.On call
//   - ctx context.Context
//   - ID string
func (_e *CapabilitiesRegistry_Expecter) Remove(ctx interface{}, ID interface{}) *CapabilitiesRegistry_Remove_Call {
	return &CapabilitiesRegistry_Remove_Call{Call: _e.mock.On("Remove", ctx, ID)}
}

func (_c *CapabilitiesRegistry_Remove_Call) Run(run func(ctx context.Context, ID string)) *CapabilitiesRegistry_Remove_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *CapabilitiesRegistry_Remove_Call) Return(_a0 error) *CapabilitiesRegistry_Remove_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *CapabilitiesRegistry_Remove_Call) RunAndReturn(run func(context.Context, string) error) *CapabilitiesRegistry_Remove_Call {
	_c.Call.Return(run)
	return _c
}

// NewCapabilitiesRegistry creates a new instance of CapabilitiesRegistry. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCapabilitiesRegistry(t interface {
	mock.TestingT
	Cleanup(func())
}) *CapabilitiesRegistry {
	mock := &CapabilitiesRegistry{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
