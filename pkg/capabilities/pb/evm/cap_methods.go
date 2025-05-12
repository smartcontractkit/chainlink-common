package evm

type CapabilityMethod string

const (
	MethodGetTransactionFee      CapabilityMethod = "GetTransactionFee"
	MethodCallContract           CapabilityMethod = "CallContract"
	MethodGetLogs                CapabilityMethod = "GetLogs"
	MethodBalanceAt              CapabilityMethod = "BalanceAt"
	MethodEstimateGas            CapabilityMethod = "EstimateGas"
	MethodGetTransactionByHash   CapabilityMethod = "GetTransactionByHash"
	MethodGetTransactionReceipt  CapabilityMethod = "GetTransactionReceipt"
	MethodLatestAndFinalizedHead CapabilityMethod = "LatestAndFinalizedHead"
	MethodQueryLogsFromCache     CapabilityMethod = "QueryLogsFromCache"
	MethodRegisterLogTracking    CapabilityMethod = "RegisterLogTracking"
	MethodUnregisterLogTracking  CapabilityMethod = "UnregisterLogTracking"
	MethodGetTransactionStatus   CapabilityMethod = "GetTransactionStatus"
)
