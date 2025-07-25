// Code generated by mockery v2.53.3. DO NOT EDIT.

package mocks

import (
	context "context"

	evm "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	mock "github.com/stretchr/testify/mock"

	primitives "github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"

	query "github.com/smartcontractkit/chainlink-common/pkg/types/query"

	types "github.com/smartcontractkit/chainlink-common/pkg/types"
)

// EVMService is an autogenerated mock type for the EVMService type
type EVMService struct {
	mock.Mock
}

type EVMService_Expecter struct {
	mock *mock.Mock
}

func (_m *EVMService) EXPECT() *EVMService_Expecter {
	return &EVMService_Expecter{mock: &_m.Mock}
}

// BalanceAt provides a mock function with given fields: ctx, request
func (_m *EVMService) BalanceAt(ctx context.Context, request evm.BalanceAtRequest) (*evm.BalanceAtReply, error) {
	ret := _m.Called(ctx, request)

	if len(ret) == 0 {
		panic("no return value specified for BalanceAt")
	}

	var r0 *evm.BalanceAtReply
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, evm.BalanceAtRequest) (*evm.BalanceAtReply, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, evm.BalanceAtRequest) *evm.BalanceAtReply); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*evm.BalanceAtReply)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, evm.BalanceAtRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EVMService_BalanceAt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'BalanceAt'
type EVMService_BalanceAt_Call struct {
	*mock.Call
}

// BalanceAt is a helper method to define mock.On call
//   - ctx context.Context
//   - request evm.BalanceAtRequest
func (_e *EVMService_Expecter) BalanceAt(ctx interface{}, request interface{}) *EVMService_BalanceAt_Call {
	return &EVMService_BalanceAt_Call{Call: _e.mock.On("BalanceAt", ctx, request)}
}

func (_c *EVMService_BalanceAt_Call) Run(run func(ctx context.Context, request evm.BalanceAtRequest)) *EVMService_BalanceAt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(evm.BalanceAtRequest))
	})
	return _c
}

func (_c *EVMService_BalanceAt_Call) Return(_a0 *evm.BalanceAtReply, _a1 error) *EVMService_BalanceAt_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EVMService_BalanceAt_Call) RunAndReturn(run func(context.Context, evm.BalanceAtRequest) (*evm.BalanceAtReply, error)) *EVMService_BalanceAt_Call {
	_c.Call.Return(run)
	return _c
}

// CalculateTransactionFee provides a mock function with given fields: ctx, receiptGasInfo
func (_m *EVMService) CalculateTransactionFee(ctx context.Context, receiptGasInfo evm.ReceiptGasInfo) (*evm.TransactionFee, error) {
	ret := _m.Called(ctx, receiptGasInfo)

	if len(ret) == 0 {
		panic("no return value specified for CalculateTransactionFee")
	}

	var r0 *evm.TransactionFee
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, evm.ReceiptGasInfo) (*evm.TransactionFee, error)); ok {
		return rf(ctx, receiptGasInfo)
	}
	if rf, ok := ret.Get(0).(func(context.Context, evm.ReceiptGasInfo) *evm.TransactionFee); ok {
		r0 = rf(ctx, receiptGasInfo)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*evm.TransactionFee)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, evm.ReceiptGasInfo) error); ok {
		r1 = rf(ctx, receiptGasInfo)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EVMService_CalculateTransactionFee_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CalculateTransactionFee'
type EVMService_CalculateTransactionFee_Call struct {
	*mock.Call
}

// CalculateTransactionFee is a helper method to define mock.On call
//   - ctx context.Context
//   - receiptGasInfo evm.ReceiptGasInfo
func (_e *EVMService_Expecter) CalculateTransactionFee(ctx interface{}, receiptGasInfo interface{}) *EVMService_CalculateTransactionFee_Call {
	return &EVMService_CalculateTransactionFee_Call{Call: _e.mock.On("CalculateTransactionFee", ctx, receiptGasInfo)}
}

func (_c *EVMService_CalculateTransactionFee_Call) Run(run func(ctx context.Context, receiptGasInfo evm.ReceiptGasInfo)) *EVMService_CalculateTransactionFee_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(evm.ReceiptGasInfo))
	})
	return _c
}

func (_c *EVMService_CalculateTransactionFee_Call) Return(_a0 *evm.TransactionFee, _a1 error) *EVMService_CalculateTransactionFee_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EVMService_CalculateTransactionFee_Call) RunAndReturn(run func(context.Context, evm.ReceiptGasInfo) (*evm.TransactionFee, error)) *EVMService_CalculateTransactionFee_Call {
	_c.Call.Return(run)
	return _c
}

// CallContract provides a mock function with given fields: ctx, request
func (_m *EVMService) CallContract(ctx context.Context, request evm.CallContractRequest) (*evm.CallContractReply, error) {
	ret := _m.Called(ctx, request)

	if len(ret) == 0 {
		panic("no return value specified for CallContract")
	}

	var r0 *evm.CallContractReply
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, evm.CallContractRequest) (*evm.CallContractReply, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, evm.CallContractRequest) *evm.CallContractReply); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*evm.CallContractReply)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, evm.CallContractRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EVMService_CallContract_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CallContract'
type EVMService_CallContract_Call struct {
	*mock.Call
}

// CallContract is a helper method to define mock.On call
//   - ctx context.Context
//   - request evm.CallContractRequest
func (_e *EVMService_Expecter) CallContract(ctx interface{}, request interface{}) *EVMService_CallContract_Call {
	return &EVMService_CallContract_Call{Call: _e.mock.On("CallContract", ctx, request)}
}

func (_c *EVMService_CallContract_Call) Run(run func(ctx context.Context, request evm.CallContractRequest)) *EVMService_CallContract_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(evm.CallContractRequest))
	})
	return _c
}

func (_c *EVMService_CallContract_Call) Return(_a0 *evm.CallContractReply, _a1 error) *EVMService_CallContract_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EVMService_CallContract_Call) RunAndReturn(run func(context.Context, evm.CallContractRequest) (*evm.CallContractReply, error)) *EVMService_CallContract_Call {
	_c.Call.Return(run)
	return _c
}

// EstimateGas provides a mock function with given fields: ctx, call
func (_m *EVMService) EstimateGas(ctx context.Context, call *evm.CallMsg) (uint64, error) {
	ret := _m.Called(ctx, call)

	if len(ret) == 0 {
		panic("no return value specified for EstimateGas")
	}

	var r0 uint64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *evm.CallMsg) (uint64, error)); ok {
		return rf(ctx, call)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *evm.CallMsg) uint64); ok {
		r0 = rf(ctx, call)
	} else {
		r0 = ret.Get(0).(uint64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *evm.CallMsg) error); ok {
		r1 = rf(ctx, call)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EVMService_EstimateGas_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'EstimateGas'
type EVMService_EstimateGas_Call struct {
	*mock.Call
}

// EstimateGas is a helper method to define mock.On call
//   - ctx context.Context
//   - call *evm.CallMsg
func (_e *EVMService_Expecter) EstimateGas(ctx interface{}, call interface{}) *EVMService_EstimateGas_Call {
	return &EVMService_EstimateGas_Call{Call: _e.mock.On("EstimateGas", ctx, call)}
}

func (_c *EVMService_EstimateGas_Call) Run(run func(ctx context.Context, call *evm.CallMsg)) *EVMService_EstimateGas_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*evm.CallMsg))
	})
	return _c
}

func (_c *EVMService_EstimateGas_Call) Return(_a0 uint64, _a1 error) *EVMService_EstimateGas_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EVMService_EstimateGas_Call) RunAndReturn(run func(context.Context, *evm.CallMsg) (uint64, error)) *EVMService_EstimateGas_Call {
	_c.Call.Return(run)
	return _c
}

// FilterLogs provides a mock function with given fields: ctx, request
func (_m *EVMService) FilterLogs(ctx context.Context, request evm.FilterLogsRequest) (*evm.FilterLogsReply, error) {
	ret := _m.Called(ctx, request)

	if len(ret) == 0 {
		panic("no return value specified for FilterLogs")
	}

	var r0 *evm.FilterLogsReply
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, evm.FilterLogsRequest) (*evm.FilterLogsReply, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, evm.FilterLogsRequest) *evm.FilterLogsReply); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*evm.FilterLogsReply)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, evm.FilterLogsRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EVMService_FilterLogs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FilterLogs'
type EVMService_FilterLogs_Call struct {
	*mock.Call
}

// FilterLogs is a helper method to define mock.On call
//   - ctx context.Context
//   - request evm.FilterLogsRequest
func (_e *EVMService_Expecter) FilterLogs(ctx interface{}, request interface{}) *EVMService_FilterLogs_Call {
	return &EVMService_FilterLogs_Call{Call: _e.mock.On("FilterLogs", ctx, request)}
}

func (_c *EVMService_FilterLogs_Call) Run(run func(ctx context.Context, request evm.FilterLogsRequest)) *EVMService_FilterLogs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(evm.FilterLogsRequest))
	})
	return _c
}

func (_c *EVMService_FilterLogs_Call) Return(_a0 *evm.FilterLogsReply, _a1 error) *EVMService_FilterLogs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EVMService_FilterLogs_Call) RunAndReturn(run func(context.Context, evm.FilterLogsRequest) (*evm.FilterLogsReply, error)) *EVMService_FilterLogs_Call {
	_c.Call.Return(run)
	return _c
}

// GetFiltersNames provides a mock function with given fields: ctx
func (_m *EVMService) GetFiltersNames(ctx context.Context) ([]string, error) {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for GetFiltersNames")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]string, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []string); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EVMService_GetFiltersNames_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetFiltersNames'
type EVMService_GetFiltersNames_Call struct {
	*mock.Call
}

// GetFiltersNames is a helper method to define mock.On call
//   - ctx context.Context
func (_e *EVMService_Expecter) GetFiltersNames(ctx interface{}) *EVMService_GetFiltersNames_Call {
	return &EVMService_GetFiltersNames_Call{Call: _e.mock.On("GetFiltersNames", ctx)}
}

func (_c *EVMService_GetFiltersNames_Call) Run(run func(ctx context.Context)) *EVMService_GetFiltersNames_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context))
	})
	return _c
}

func (_c *EVMService_GetFiltersNames_Call) Return(_a0 []string, _a1 error) *EVMService_GetFiltersNames_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EVMService_GetFiltersNames_Call) RunAndReturn(run func(context.Context) ([]string, error)) *EVMService_GetFiltersNames_Call {
	_c.Call.Return(run)
	return _c
}

// GetForwarderForEOA provides a mock function with given fields: ctx, eoa, ocr2AggregatorID, pluginType
func (_m *EVMService) GetForwarderForEOA(ctx context.Context, eoa [20]byte, ocr2AggregatorID [20]byte, pluginType string) ([20]byte, error) {
	ret := _m.Called(ctx, eoa, ocr2AggregatorID, pluginType)

	if len(ret) == 0 {
		panic("no return value specified for GetForwarderForEOA")
	}

	var r0 [20]byte
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, [20]byte, [20]byte, string) ([20]byte, error)); ok {
		return rf(ctx, eoa, ocr2AggregatorID, pluginType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, [20]byte, [20]byte, string) [20]byte); ok {
		r0 = rf(ctx, eoa, ocr2AggregatorID, pluginType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([20]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, [20]byte, [20]byte, string) error); ok {
		r1 = rf(ctx, eoa, ocr2AggregatorID, pluginType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EVMService_GetForwarderForEOA_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetForwarderForEOA'
type EVMService_GetForwarderForEOA_Call struct {
	*mock.Call
}

// GetForwarderForEOA is a helper method to define mock.On call
//   - ctx context.Context
//   - eoa [20]byte
//   - ocr2AggregatorID [20]byte
//   - pluginType string
func (_e *EVMService_Expecter) GetForwarderForEOA(ctx interface{}, eoa interface{}, ocr2AggregatorID interface{}, pluginType interface{}) *EVMService_GetForwarderForEOA_Call {
	return &EVMService_GetForwarderForEOA_Call{Call: _e.mock.On("GetForwarderForEOA", ctx, eoa, ocr2AggregatorID, pluginType)}
}

func (_c *EVMService_GetForwarderForEOA_Call) Run(run func(ctx context.Context, eoa [20]byte, ocr2AggregatorID [20]byte, pluginType string)) *EVMService_GetForwarderForEOA_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([20]byte), args[2].([20]byte), args[3].(string))
	})
	return _c
}

func (_c *EVMService_GetForwarderForEOA_Call) Return(forwarder [20]byte, err error) *EVMService_GetForwarderForEOA_Call {
	_c.Call.Return(forwarder, err)
	return _c
}

func (_c *EVMService_GetForwarderForEOA_Call) RunAndReturn(run func(context.Context, [20]byte, [20]byte, string) ([20]byte, error)) *EVMService_GetForwarderForEOA_Call {
	_c.Call.Return(run)
	return _c
}

// GetTransactionByHash provides a mock function with given fields: ctx, hash
func (_m *EVMService) GetTransactionByHash(ctx context.Context, hash [32]byte) (*evm.Transaction, error) {
	ret := _m.Called(ctx, hash)

	if len(ret) == 0 {
		panic("no return value specified for GetTransactionByHash")
	}

	var r0 *evm.Transaction
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, [32]byte) (*evm.Transaction, error)); ok {
		return rf(ctx, hash)
	}
	if rf, ok := ret.Get(0).(func(context.Context, [32]byte) *evm.Transaction); ok {
		r0 = rf(ctx, hash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*evm.Transaction)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, [32]byte) error); ok {
		r1 = rf(ctx, hash)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EVMService_GetTransactionByHash_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTransactionByHash'
type EVMService_GetTransactionByHash_Call struct {
	*mock.Call
}

// GetTransactionByHash is a helper method to define mock.On call
//   - ctx context.Context
//   - hash [32]byte
func (_e *EVMService_Expecter) GetTransactionByHash(ctx interface{}, hash interface{}) *EVMService_GetTransactionByHash_Call {
	return &EVMService_GetTransactionByHash_Call{Call: _e.mock.On("GetTransactionByHash", ctx, hash)}
}

func (_c *EVMService_GetTransactionByHash_Call) Run(run func(ctx context.Context, hash [32]byte)) *EVMService_GetTransactionByHash_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([32]byte))
	})
	return _c
}

func (_c *EVMService_GetTransactionByHash_Call) Return(_a0 *evm.Transaction, _a1 error) *EVMService_GetTransactionByHash_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EVMService_GetTransactionByHash_Call) RunAndReturn(run func(context.Context, [32]byte) (*evm.Transaction, error)) *EVMService_GetTransactionByHash_Call {
	_c.Call.Return(run)
	return _c
}

// GetTransactionFee provides a mock function with given fields: ctx, transactionID
func (_m *EVMService) GetTransactionFee(ctx context.Context, transactionID string) (*evm.TransactionFee, error) {
	ret := _m.Called(ctx, transactionID)

	if len(ret) == 0 {
		panic("no return value specified for GetTransactionFee")
	}

	var r0 *evm.TransactionFee
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*evm.TransactionFee, error)); ok {
		return rf(ctx, transactionID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *evm.TransactionFee); ok {
		r0 = rf(ctx, transactionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*evm.TransactionFee)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, transactionID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EVMService_GetTransactionFee_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTransactionFee'
type EVMService_GetTransactionFee_Call struct {
	*mock.Call
}

// GetTransactionFee is a helper method to define mock.On call
//   - ctx context.Context
//   - transactionID string
func (_e *EVMService_Expecter) GetTransactionFee(ctx interface{}, transactionID interface{}) *EVMService_GetTransactionFee_Call {
	return &EVMService_GetTransactionFee_Call{Call: _e.mock.On("GetTransactionFee", ctx, transactionID)}
}

func (_c *EVMService_GetTransactionFee_Call) Run(run func(ctx context.Context, transactionID string)) *EVMService_GetTransactionFee_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *EVMService_GetTransactionFee_Call) Return(_a0 *evm.TransactionFee, _a1 error) *EVMService_GetTransactionFee_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EVMService_GetTransactionFee_Call) RunAndReturn(run func(context.Context, string) (*evm.TransactionFee, error)) *EVMService_GetTransactionFee_Call {
	_c.Call.Return(run)
	return _c
}

// GetTransactionReceipt provides a mock function with given fields: ctx, txHash
func (_m *EVMService) GetTransactionReceipt(ctx context.Context, txHash [32]byte) (*evm.Receipt, error) {
	ret := _m.Called(ctx, txHash)

	if len(ret) == 0 {
		panic("no return value specified for GetTransactionReceipt")
	}

	var r0 *evm.Receipt
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, [32]byte) (*evm.Receipt, error)); ok {
		return rf(ctx, txHash)
	}
	if rf, ok := ret.Get(0).(func(context.Context, [32]byte) *evm.Receipt); ok {
		r0 = rf(ctx, txHash)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*evm.Receipt)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, [32]byte) error); ok {
		r1 = rf(ctx, txHash)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EVMService_GetTransactionReceipt_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTransactionReceipt'
type EVMService_GetTransactionReceipt_Call struct {
	*mock.Call
}

// GetTransactionReceipt is a helper method to define mock.On call
//   - ctx context.Context
//   - txHash [32]byte
func (_e *EVMService_Expecter) GetTransactionReceipt(ctx interface{}, txHash interface{}) *EVMService_GetTransactionReceipt_Call {
	return &EVMService_GetTransactionReceipt_Call{Call: _e.mock.On("GetTransactionReceipt", ctx, txHash)}
}

func (_c *EVMService_GetTransactionReceipt_Call) Run(run func(ctx context.Context, txHash [32]byte)) *EVMService_GetTransactionReceipt_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([32]byte))
	})
	return _c
}

func (_c *EVMService_GetTransactionReceipt_Call) Return(_a0 *evm.Receipt, _a1 error) *EVMService_GetTransactionReceipt_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EVMService_GetTransactionReceipt_Call) RunAndReturn(run func(context.Context, [32]byte) (*evm.Receipt, error)) *EVMService_GetTransactionReceipt_Call {
	_c.Call.Return(run)
	return _c
}

// GetTransactionStatus provides a mock function with given fields: ctx, transactionID
func (_m *EVMService) GetTransactionStatus(ctx context.Context, transactionID string) (types.TransactionStatus, error) {
	ret := _m.Called(ctx, transactionID)

	if len(ret) == 0 {
		panic("no return value specified for GetTransactionStatus")
	}

	var r0 types.TransactionStatus
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (types.TransactionStatus, error)); ok {
		return rf(ctx, transactionID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) types.TransactionStatus); ok {
		r0 = rf(ctx, transactionID)
	} else {
		r0 = ret.Get(0).(types.TransactionStatus)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, transactionID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EVMService_GetTransactionStatus_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetTransactionStatus'
type EVMService_GetTransactionStatus_Call struct {
	*mock.Call
}

// GetTransactionStatus is a helper method to define mock.On call
//   - ctx context.Context
//   - transactionID string
func (_e *EVMService_Expecter) GetTransactionStatus(ctx interface{}, transactionID interface{}) *EVMService_GetTransactionStatus_Call {
	return &EVMService_GetTransactionStatus_Call{Call: _e.mock.On("GetTransactionStatus", ctx, transactionID)}
}

func (_c *EVMService_GetTransactionStatus_Call) Run(run func(ctx context.Context, transactionID string)) *EVMService_GetTransactionStatus_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *EVMService_GetTransactionStatus_Call) Return(_a0 types.TransactionStatus, _a1 error) *EVMService_GetTransactionStatus_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EVMService_GetTransactionStatus_Call) RunAndReturn(run func(context.Context, string) (types.TransactionStatus, error)) *EVMService_GetTransactionStatus_Call {
	_c.Call.Return(run)
	return _c
}

// HeaderByNumber provides a mock function with given fields: ctx, request
func (_m *EVMService) HeaderByNumber(ctx context.Context, request evm.HeaderByNumberRequest) (*evm.HeaderByNumberReply, error) {
	ret := _m.Called(ctx, request)

	if len(ret) == 0 {
		panic("no return value specified for HeaderByNumber")
	}

	var r0 *evm.HeaderByNumberReply
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, evm.HeaderByNumberRequest) (*evm.HeaderByNumberReply, error)); ok {
		return rf(ctx, request)
	}
	if rf, ok := ret.Get(0).(func(context.Context, evm.HeaderByNumberRequest) *evm.HeaderByNumberReply); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*evm.HeaderByNumberReply)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, evm.HeaderByNumberRequest) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EVMService_HeaderByNumber_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HeaderByNumber'
type EVMService_HeaderByNumber_Call struct {
	*mock.Call
}

// HeaderByNumber is a helper method to define mock.On call
//   - ctx context.Context
//   - request evm.HeaderByNumberRequest
func (_e *EVMService_Expecter) HeaderByNumber(ctx interface{}, request interface{}) *EVMService_HeaderByNumber_Call {
	return &EVMService_HeaderByNumber_Call{Call: _e.mock.On("HeaderByNumber", ctx, request)}
}

func (_c *EVMService_HeaderByNumber_Call) Run(run func(ctx context.Context, request evm.HeaderByNumberRequest)) *EVMService_HeaderByNumber_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(evm.HeaderByNumberRequest))
	})
	return _c
}

func (_c *EVMService_HeaderByNumber_Call) Return(_a0 *evm.HeaderByNumberReply, _a1 error) *EVMService_HeaderByNumber_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EVMService_HeaderByNumber_Call) RunAndReturn(run func(context.Context, evm.HeaderByNumberRequest) (*evm.HeaderByNumberReply, error)) *EVMService_HeaderByNumber_Call {
	_c.Call.Return(run)
	return _c
}

// QueryTrackedLogs provides a mock function with given fields: ctx, filterQuery, limitAndSort, confidenceLevel
func (_m *EVMService) QueryTrackedLogs(ctx context.Context, filterQuery []query.Expression, limitAndSort query.LimitAndSort, confidenceLevel primitives.ConfidenceLevel) ([]*evm.Log, error) {
	ret := _m.Called(ctx, filterQuery, limitAndSort, confidenceLevel)

	if len(ret) == 0 {
		panic("no return value specified for QueryTrackedLogs")
	}

	var r0 []*evm.Log
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []query.Expression, query.LimitAndSort, primitives.ConfidenceLevel) ([]*evm.Log, error)); ok {
		return rf(ctx, filterQuery, limitAndSort, confidenceLevel)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []query.Expression, query.LimitAndSort, primitives.ConfidenceLevel) []*evm.Log); ok {
		r0 = rf(ctx, filterQuery, limitAndSort, confidenceLevel)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*evm.Log)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []query.Expression, query.LimitAndSort, primitives.ConfidenceLevel) error); ok {
		r1 = rf(ctx, filterQuery, limitAndSort, confidenceLevel)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EVMService_QueryTrackedLogs_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'QueryTrackedLogs'
type EVMService_QueryTrackedLogs_Call struct {
	*mock.Call
}

// QueryTrackedLogs is a helper method to define mock.On call
//   - ctx context.Context
//   - filterQuery []query.Expression
//   - limitAndSort query.LimitAndSort
//   - confidenceLevel primitives.ConfidenceLevel
func (_e *EVMService_Expecter) QueryTrackedLogs(ctx interface{}, filterQuery interface{}, limitAndSort interface{}, confidenceLevel interface{}) *EVMService_QueryTrackedLogs_Call {
	return &EVMService_QueryTrackedLogs_Call{Call: _e.mock.On("QueryTrackedLogs", ctx, filterQuery, limitAndSort, confidenceLevel)}
}

func (_c *EVMService_QueryTrackedLogs_Call) Run(run func(ctx context.Context, filterQuery []query.Expression, limitAndSort query.LimitAndSort, confidenceLevel primitives.ConfidenceLevel)) *EVMService_QueryTrackedLogs_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].([]query.Expression), args[2].(query.LimitAndSort), args[3].(primitives.ConfidenceLevel))
	})
	return _c
}

func (_c *EVMService_QueryTrackedLogs_Call) Return(_a0 []*evm.Log, _a1 error) *EVMService_QueryTrackedLogs_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EVMService_QueryTrackedLogs_Call) RunAndReturn(run func(context.Context, []query.Expression, query.LimitAndSort, primitives.ConfidenceLevel) ([]*evm.Log, error)) *EVMService_QueryTrackedLogs_Call {
	_c.Call.Return(run)
	return _c
}

// RegisterLogTracking provides a mock function with given fields: ctx, filter
func (_m *EVMService) RegisterLogTracking(ctx context.Context, filter evm.LPFilterQuery) error {
	ret := _m.Called(ctx, filter)

	if len(ret) == 0 {
		panic("no return value specified for RegisterLogTracking")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, evm.LPFilterQuery) error); ok {
		r0 = rf(ctx, filter)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EVMService_RegisterLogTracking_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'RegisterLogTracking'
type EVMService_RegisterLogTracking_Call struct {
	*mock.Call
}

// RegisterLogTracking is a helper method to define mock.On call
//   - ctx context.Context
//   - filter evm.LPFilterQuery
func (_e *EVMService_Expecter) RegisterLogTracking(ctx interface{}, filter interface{}) *EVMService_RegisterLogTracking_Call {
	return &EVMService_RegisterLogTracking_Call{Call: _e.mock.On("RegisterLogTracking", ctx, filter)}
}

func (_c *EVMService_RegisterLogTracking_Call) Run(run func(ctx context.Context, filter evm.LPFilterQuery)) *EVMService_RegisterLogTracking_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(evm.LPFilterQuery))
	})
	return _c
}

func (_c *EVMService_RegisterLogTracking_Call) Return(_a0 error) *EVMService_RegisterLogTracking_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *EVMService_RegisterLogTracking_Call) RunAndReturn(run func(context.Context, evm.LPFilterQuery) error) *EVMService_RegisterLogTracking_Call {
	_c.Call.Return(run)
	return _c
}

// SubmitTransaction provides a mock function with given fields: ctx, txRequest
func (_m *EVMService) SubmitTransaction(ctx context.Context, txRequest evm.SubmitTransactionRequest) (*evm.TransactionResult, error) {
	ret := _m.Called(ctx, txRequest)

	if len(ret) == 0 {
		panic("no return value specified for SubmitTransaction")
	}

	var r0 *evm.TransactionResult
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, evm.SubmitTransactionRequest) (*evm.TransactionResult, error)); ok {
		return rf(ctx, txRequest)
	}
	if rf, ok := ret.Get(0).(func(context.Context, evm.SubmitTransactionRequest) *evm.TransactionResult); ok {
		r0 = rf(ctx, txRequest)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*evm.TransactionResult)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, evm.SubmitTransactionRequest) error); ok {
		r1 = rf(ctx, txRequest)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EVMService_SubmitTransaction_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SubmitTransaction'
type EVMService_SubmitTransaction_Call struct {
	*mock.Call
}

// SubmitTransaction is a helper method to define mock.On call
//   - ctx context.Context
//   - txRequest evm.SubmitTransactionRequest
func (_e *EVMService_Expecter) SubmitTransaction(ctx interface{}, txRequest interface{}) *EVMService_SubmitTransaction_Call {
	return &EVMService_SubmitTransaction_Call{Call: _e.mock.On("SubmitTransaction", ctx, txRequest)}
}

func (_c *EVMService_SubmitTransaction_Call) Run(run func(ctx context.Context, txRequest evm.SubmitTransactionRequest)) *EVMService_SubmitTransaction_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(evm.SubmitTransactionRequest))
	})
	return _c
}

func (_c *EVMService_SubmitTransaction_Call) Return(_a0 *evm.TransactionResult, _a1 error) *EVMService_SubmitTransaction_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EVMService_SubmitTransaction_Call) RunAndReturn(run func(context.Context, evm.SubmitTransactionRequest) (*evm.TransactionResult, error)) *EVMService_SubmitTransaction_Call {
	_c.Call.Return(run)
	return _c
}

// UnregisterLogTracking provides a mock function with given fields: ctx, filterName
func (_m *EVMService) UnregisterLogTracking(ctx context.Context, filterName string) error {
	ret := _m.Called(ctx, filterName)

	if len(ret) == 0 {
		panic("no return value specified for UnregisterLogTracking")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, filterName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EVMService_UnregisterLogTracking_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UnregisterLogTracking'
type EVMService_UnregisterLogTracking_Call struct {
	*mock.Call
}

// UnregisterLogTracking is a helper method to define mock.On call
//   - ctx context.Context
//   - filterName string
func (_e *EVMService_Expecter) UnregisterLogTracking(ctx interface{}, filterName interface{}) *EVMService_UnregisterLogTracking_Call {
	return &EVMService_UnregisterLogTracking_Call{Call: _e.mock.On("UnregisterLogTracking", ctx, filterName)}
}

func (_c *EVMService_UnregisterLogTracking_Call) Run(run func(ctx context.Context, filterName string)) *EVMService_UnregisterLogTracking_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *EVMService_UnregisterLogTracking_Call) Return(_a0 error) *EVMService_UnregisterLogTracking_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *EVMService_UnregisterLogTracking_Call) RunAndReturn(run func(context.Context, string) error) *EVMService_UnregisterLogTracking_Call {
	_c.Call.Return(run)
	return _c
}

// NewEVMService creates a new instance of EVMService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEVMService(t interface {
	mock.TestingT
	Cleanup(func())
}) *EVMService {
	mock := &EVMService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
