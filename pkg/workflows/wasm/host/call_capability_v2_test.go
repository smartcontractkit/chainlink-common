package host

import (
	"context"
	"math"
	"testing"

	"github.com/bytecodealliance/wasmtime-go/v47"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/cresettings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/limits"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/host/mocks"

	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

// callCapabilityV1ParamCount is the number of params the legacy V1
// call_capability import declares (req, reqLen).
const callCapabilityV1ParamCount = 2

// --- WAT modules ---

// watCallCapV2 is a minimal WAT module that imports call_capability with 4
// params (i32, i32, i32, i32) → i64, plus a version_v2 import so the host
// treats it as a NoDAG module. Used for signature detection tests via NewModule.
const watCallCapV2 = `
	(module
	  (import "env" "version_v2" (func $version_v2))
	  (import "env" "call_capability" (func $call_capability (param i32 i32 i32 i32) (result i64)))
	  (memory (export "memory") 1)
	  (func (export "_start")
	    call $version_v2)
	)`

// watCallCapV1 is a minimal WAT module that imports call_capability with 2
// params (i32, i32) → i64, plus a version_v2 import so the host treats it as
// a NoDAG module. Used for signature detection tests via NewModule.
const watCallCapV1 = `
	(module
	  (import "env" "version_v2" (func $version_v2))
	  (import "env" "call_capability" (func $call_capability (param i32 i32) (result i64)))
	  (memory (export "memory") 1)
	  (func (export "_start")
	    call $version_v2)
	)`

// watCallCapV2Test is a WAT module with a 4-param call_capability import and
// an exported call_cap wrapper that forwards all 4 params. Used for host
// function behavior tests — the WAT controls whether V1 or V2 is created
// via the import signature, and createCallCapFn dispatches accordingly.
const watCallCapV2Test = `
	(module
	  (import "env" "call_capability" (func $call_capability (param i32 i32 i32 i32) (result i64)))
	  (memory (export "memory") 1)
	  (func (export "_start"))
	  (func (export "call_cap") (param i32 i32 i32 i32) (result i64)
	    local.get 0
	    local.get 1
	    local.get 2
	    local.get 3
	    call $call_capability)
	)`

// watCallCapV1Test is a WAT module with a 2-param call_capability import and
// an exported call_cap wrapper that forwards both params.
const watCallCapV1Test = `
	(module
	  (import "env" "call_capability" (func $call_capability (param i32 i32) (result i64)))
	  (memory (export "memory") 1)
	  (func (export "_start"))
	  (func (export "call_cap") (param i32 i32) (result i64)
	    local.get 0
	    local.get 1
	    call $call_capability)
	)`

// --- Memory helpers ---

func memWrite(t *testing.T, mem *wasmtime.Memory, store *wasmtime.Store, offset int32, data []byte) {
	t.Helper()
	raw := mem.UnsafeData(store)
	copy(raw[offset:], data)
}

func memRead(t *testing.T, mem *wasmtime.Memory, store *wasmtime.Store, offset, n int32) []byte {
	t.Helper()
	raw := mem.UnsafeData(store)
	out := make([]byte, n)
	copy(out, raw[offset:offset+n])
	return out
}

// instantiateCallCapModule compiles a WAT module, uses createCallCapFn to
// select the V1 or V2 host function based on the module's call_capability
// import param count, and instantiates the module with that function linked.
// The WAT module's import signature controls which function is created.
// The provided logger is passed to createCallCapFn so test observers can
// capture log output.
func instantiateCallCapModule(t *testing.T, wat string, exec *execution[*sdkpb.ExecutionResult], lggr logger.Logger) (*wasmtime.Store, *wasmtime.Instance, *wasmtime.Memory) {
	t.Helper()
	wasmBytes, err := wasmtime.Wat2Wasm(wat)
	require.NoError(t, err)

	// Use NewModule to detect the call_capability param count, matching
	// the production code path. The WAT must include a version_v2 import
	// so NewModule treats it as a NoDAG module.
	mc := defaultNoDAGModCfg(t)
	mc.Logger = lggr
	m, err := NewModule(t.Context(), mc, wasmBytes)
	require.NoError(t, err)

	hostFn := createCallCapFn(lggr, exec, m.callCapParams)

	store := wasmtime.NewStore(m.engine)
	// NewModule enables epoch interruption; set a generous deadline so the
	// test doesn't hit an interrupt when calling the exported function.
	store.SetEpochDeadline(math.MaxUint64)
	linker := wasmtime.NewLinker(m.engine)
	require.NoError(t, linker.FuncWrap("env", "call_capability", hostFn))

	inst, err := linker.Instantiate(store, m.module)
	require.NoError(t, err)

	mem := inst.GetExport(store, "memory").Memory()
	return store, inst, mem
}

// --- Signature detection tests ---

func TestNewModule_DetectsCallCapabilityParamCount_V2(t *testing.T) {
	wasmBytes, err := wasmtime.Wat2Wasm(watCallCapV2)
	require.NoError(t, err)

	mc := defaultNoDAGModCfg(t)
	m, err := NewModule(t.Context(), mc, wasmBytes)
	require.NoError(t, err)
	require.False(t, m.IsLegacyDAG(), "expected NoDAG module")

	assert.Equal(t, callCapabilityV2ParamCount, m.callCapParams,
		"module should detect 4-param call_capability import")
}

func TestNewModule_DetectsCallCapabilityParamCount_V1(t *testing.T) {
	wasmBytes, err := wasmtime.Wat2Wasm(watCallCapV1)
	require.NoError(t, err)

	mc := defaultNoDAGModCfg(t)
	m, err := NewModule(t.Context(), mc, wasmBytes)
	require.NoError(t, err)
	require.False(t, m.IsLegacyDAG(), "expected NoDAG module")

	assert.Equal(t, callCapabilityV1ParamCount, m.callCapParams,
		"module should detect 2-param call_capability import")
}

func TestNewModule_DetectsCallCapabilityParamCount_NoImport(t *testing.T) {
	wat := `
	(module
	  (import "env" "version_v2" (func $version_v2))
	  (memory (export "memory") 1)
	  (func (export "_start")
	    call $version_v2)
	)`
	wasmBytes, err := wasmtime.Wat2Wasm(wat)
	require.NoError(t, err)

	mc := defaultNoDAGModCfg(t)
	m, err := NewModule(t.Context(), mc, wasmBytes)
	require.NoError(t, err)

	assert.Equal(t, 0, m.callCapParams,
		"module without call_capability import should have 0 params")
}

// --- Host function tests (V2: 4-param import → response buffer) ---

func TestCallCapability_V2_CallCapAsyncErrorWritesToResponseBuffer(t *testing.T) {
	lggr, logs := logger.TestObserved(t, zapcore.ErrorLevel)

	// Create a limiter with capacity 0 so callCapAsync always fails.
	zeroLimiter := limits.GlobalResourcePoolLimiter(0)
	mockExecHelper := mocks.NewMockExecutionHelper(t)

	// Use an already-cancelled context so the zero-capacity limiter returns
	// immediately instead of blocking.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	exec := &execution[*sdkpb.ExecutionResult]{
		capabilityResponses: map[int32]<-chan *sdkpb.CapabilityResponse{},
		usedCallbackIDs:     map[string]bool{},
		pendingCallsLimiter: zeroLimiter,
		ctx:                 ctx,
		executor:            mockExecHelper,
	}

	// instantiateCallCapModule uses createCallCapFn internally, dispatching
	// to V2 because watCallCapV2Test declares a 4-param import.
	store, inst, mem := instantiateCallCapModule(t, watCallCapV2Test, exec, lggr)

	// Marshal a valid CapabilityRequest so wasmRead and proto.Unmarshal succeed,
	// forcing the failure to come from callCapAsync.
	req := &sdkpb.CapabilityRequest{
		Id:         "test-cap@1.0.0",
		CallbackId: 1,
	}
	reqBytes, err := proto.Marshal(req)
	require.NoError(t, err)

	reqOffset := int32(0)
	memWrite(t, mem, store, reqOffset, reqBytes)

	respOffset := int32(256)
	respSize := int32(256)

	callCap := inst.GetExport(store, "call_cap").Func()
	result, err := callCap.Call(store, reqOffset, int32(len(reqBytes)), respOffset, respSize)
	require.NoError(t, err)

	resultI64 := result.(int64)

	// Should return a negative value (error).
	assert.Less(t, resultI64, int64(0),
		"V2 should return negative on callCapAsync error")

	// Read the error string from the response buffer.
	bytesWritten := int(-resultI64)
	respData := memRead(t, mem, store, respOffset, int32(bytesWritten))

	errStr := string(respData)
	// The error chain is: callCapAsync wraps the limiter error which wraps
	// context.Canceled and ErrorResourceLimited. The full string is:
	// "error calling callCapAsync: context error (context canceled) after
	// waiting <duration> for limit: resource limited: cannot use 1, already
	// using 0/0"
	assert.Contains(t, errStr, "error calling callCapAsync",
		"V2 should wrap the error with callCapAsync context")
	assert.Contains(t, errStr, "context canceled",
		"V2 should surface the context cancellation error")
	assert.Contains(t, errStr, "resource limited",
		"V2 should surface the resource limit error")

	// Verify the error was logged.
	require.Len(t, logs.AllUntimed(), 1)
	assert.Equal(t, zapcore.ErrorLevel, logs.AllUntimed()[0].Level)
}

func TestCallCapability_V2_SuccessReturnsZero(t *testing.T) {
	mockExecHelper := mocks.NewMockExecutionHelper(t)
	// callCapAsync runs CallCapability in a goroutine; use .Maybe() since the
	// test only checks the synchronous return value, not the async result.
	mockExecHelper.EXPECT().
		CallCapability(mock.Anything, mock.Anything).
		Return(&sdkpb.CapabilityResponse{}, nil).Maybe()

	exec := &execution[*sdkpb.ExecutionResult]{
		capabilityResponses: map[int32]<-chan *sdkpb.CapabilityResponse{},
		usedCallbackIDs:     map[string]bool{},
		pendingCallsLimiter: limits.GlobalResourcePoolLimiter(cresettings.Default.PerWorkflow.CapabilityConcurrencyLimit.DefaultValue),
		ctx:                 t.Context(),
		executor:            mockExecHelper,
	}

	store, inst, mem := instantiateCallCapModule(t, watCallCapV2Test, exec, logger.Test(t))

	req := &sdkpb.CapabilityRequest{
		Id:         "test-cap@1.0.0",
		CallbackId: 42,
	}
	reqBytes, err := proto.Marshal(req)
	require.NoError(t, err)

	reqOffset := int32(0)
	memWrite(t, mem, store, reqOffset, reqBytes)

	// Pre-fill response buffer with a sentinel to verify it's not overwritten.
	respOffset := int32(256)
	sentinel := []byte("SENTINEL_DATA_THAT_SHOULD_NOT_BE_OVERWRITTEN")
	memWrite(t, mem, store, respOffset, sentinel)

	callCap := inst.GetExport(store, "call_cap").Func()
	result, err := callCap.Call(store, reqOffset, int32(len(reqBytes)), respOffset, int32(256))
	require.NoError(t, err)

	resultI64 := result.(int64)
	assert.Equal(t, int64(0), resultI64,
		"V2 should return 0 on success")

	// Verify response buffer was not modified.
	respData := memRead(t, mem, store, respOffset, int32(len(sentinel)))
	assert.Equal(t, sentinel, respData,
		"V2 should not write to response buffer on success")
}

func TestCallCapability_V2_ProtoUnmarshalErrorWritesToResponseBuffer(t *testing.T) {
	lggr, logs := logger.TestObserved(t, zapcore.ErrorLevel)
	mockExecHelper := mocks.NewMockExecutionHelper(t)

	exec := &execution[*sdkpb.ExecutionResult]{
		capabilityResponses: map[int32]<-chan *sdkpb.CapabilityResponse{},
		usedCallbackIDs:     map[string]bool{},
		pendingCallsLimiter: limits.GlobalResourcePoolLimiter(cresettings.Default.PerWorkflow.CapabilityConcurrencyLimit.DefaultValue),
		ctx:                 t.Context(),
		executor:            mockExecHelper,
	}

	store, inst, mem := instantiateCallCapModule(t, watCallCapV2Test, exec, lggr)

	// Write invalid proto data to the request buffer.
	invalidProto := []byte("this is not a valid protobuf message")
	reqOffset := int32(0)
	memWrite(t, mem, store, reqOffset, invalidProto)

	// Put known data in the request buffer area after the invalid proto,
	// to verify it's NOT overwritten (V1 bug was writing to request buffer).
	reqSentinelOffset := int32(64)
	reqSentinel := []byte("REQ_SENTINEL")
	memWrite(t, mem, store, reqSentinelOffset, reqSentinel)

	respOffset := int32(256)
	respSize := int32(256)

	callCap := inst.GetExport(store, "call_cap").Func()
	result, err := callCap.Call(store, reqOffset, int32(len(invalidProto)), respOffset, respSize)
	require.NoError(t, err)

	resultI64 := result.(int64)
	assert.Less(t, resultI64, int64(0),
		"V2 should return negative on proto unmarshal error")

	// Read error from response buffer.
	bytesWritten := int(-resultI64)
	respData := memRead(t, mem, store, respOffset, int32(bytesWritten))
	assert.Contains(t, string(respData), "proto unmarshal",
		"V2 should write unmarshal error to response buffer")

	// Verify request buffer sentinel was NOT overwritten.
	reqSentinelData := memRead(t, mem, store, reqSentinelOffset, int32(len(reqSentinel)))
	assert.Equal(t, reqSentinel, reqSentinelData,
		"V2 should NOT write to request buffer")

	// Verify error was logged.
	require.Len(t, logs.AllUntimed(), 1)
	assert.Equal(t, zapcore.ErrorLevel, logs.AllUntimed()[0].Level)
}

// --- Host function tests (V1: 2-param import → request buffer) ---

// TestCallCapability_V1_CallCapAsyncErrorWritesToRequestBuffer verifies that
// V1 (legacy 2-param) writes errors to the request buffer — the same buffer
// is used for both request and response. This matches the existing behavior
// for wasmRead and proto.Unmarshal errors, and now also applies to callCapAsync
// errors (previously bare -1). The error detail is present but in the wrong
// buffer; V2 fixes this by using a dedicated response buffer.
func TestCallCapability_V1_CallCapAsyncErrorWritesToRequestBuffer(t *testing.T) {
	lggr, logs := logger.TestObserved(t, zapcore.ErrorLevel)

	zeroLimiter := limits.GlobalResourcePoolLimiter(0)
	mockExecHelper := mocks.NewMockExecutionHelper(t)

	// Use an already-cancelled context so the zero-capacity limiter returns
	// immediately instead of blocking.
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	exec := &execution[*sdkpb.ExecutionResult]{
		capabilityResponses: map[int32]<-chan *sdkpb.CapabilityResponse{},
		usedCallbackIDs:     map[string]bool{},
		pendingCallsLimiter: zeroLimiter,
		ctx:                 ctx,
		executor:            mockExecHelper,
	}

	// instantiateCallCapModule uses createCallCapFn internally, dispatching
	// to V1 because watCallCapV1Test declares a 2-param import.
	store, inst, mem := instantiateCallCapModule(t, watCallCapV1Test, exec, lggr)

	req := &sdkpb.CapabilityRequest{
		Id:         "test-cap@1.0.0",
		CallbackId: 1,
	}
	reqBytes, err := proto.Marshal(req)
	require.NoError(t, err)

	reqOffset := int32(0)
	memWrite(t, mem, store, reqOffset, reqBytes)

	callCap := inst.GetExport(store, "call_cap").Func()
	result, err := callCap.Call(store, reqOffset, int32(len(reqBytes)))
	require.NoError(t, err)

	resultI64 := result.(int64)
	// V1 writes the error to the request buffer (same buffer for both) and
	// returns a negative value. This is the known limitation that V2 fixes
	// by using a dedicated response buffer.
	assert.Less(t, resultI64, int64(0),
		"V1 should return negative on callCapAsync error")

	// The error string is written to the request buffer (offset 0), but
	// truncated to fit the request buffer size. Check for the prefix that
	// survives truncation.
	bytesWritten := int(-resultI64)
	reqData := memRead(t, mem, store, reqOffset, int32(bytesWritten))
	assert.Contains(t, string(reqData), "error calling",
		"V1 writes error to request buffer (same buffer for both)")

	// Verify error was logged.
	require.Len(t, logs.AllUntimed(), 1)
	assert.Equal(t, zapcore.ErrorLevel, logs.AllUntimed()[0].Level)
}

func TestCallCapability_V1_SuccessReturnsZero(t *testing.T) {
	mockExecHelper := mocks.NewMockExecutionHelper(t)
	// callCapAsync runs CallCapability in a goroutine; use .Maybe() since the
	// test only checks the synchronous return value, not the async result.
	mockExecHelper.EXPECT().
		CallCapability(mock.Anything, mock.Anything).
		Return(&sdkpb.CapabilityResponse{}, nil).Maybe()

	exec := &execution[*sdkpb.ExecutionResult]{
		capabilityResponses: map[int32]<-chan *sdkpb.CapabilityResponse{},
		usedCallbackIDs:     map[string]bool{},
		pendingCallsLimiter: limits.GlobalResourcePoolLimiter(cresettings.Default.PerWorkflow.CapabilityConcurrencyLimit.DefaultValue),
		ctx:                 t.Context(),
		executor:            mockExecHelper,
	}

	store, inst, mem := instantiateCallCapModule(t, watCallCapV1Test, exec, logger.Test(t))

	req := &sdkpb.CapabilityRequest{
		Id:         "test-cap@1.0.0",
		CallbackId: 42,
	}
	reqBytes, err := proto.Marshal(req)
	require.NoError(t, err)

	reqOffset := int32(0)
	memWrite(t, mem, store, reqOffset, reqBytes)

	callCap := inst.GetExport(store, "call_cap").Func()
	result, err := callCap.Call(store, reqOffset, int32(len(reqBytes)))
	require.NoError(t, err)

	resultI64 := result.(int64)
	assert.Equal(t, int64(0), resultI64,
		"V1 should return 0 on success")
}

// --- Dynamic linking tests ---

func TestLinkNoDAG_RegistersV2For4ParamImport(t *testing.T) {
	wasmBytes, err := wasmtime.Wat2Wasm(watCallCapV2)
	require.NoError(t, err)

	mc := defaultNoDAGModCfg(t)
	m, err := NewModule(t.Context(), mc, wasmBytes)
	require.NoError(t, err)
	require.Equal(t, callCapabilityV2ParamCount, m.callCapParams)

	// The module should link successfully with the V2 host function.
	// If the wrong function signature is registered, wasmtime will reject the import.
	mockExecHelper := mocks.NewMockExecutionHelper(t)

	store := wasmtime.NewStore(m.engine)
	exec := &execution[*sdkpb.ExecutionResult]{
		module:              m,
		ctx:                 t.Context(),
		executor:            mockExecHelper,
		capabilityResponses: map[int32]<-chan *sdkpb.CapabilityResponse{},
		usedCallbackIDs:     map[string]bool{},
		pendingCallsLimiter: limits.GlobalResourcePoolLimiter(cresettings.Default.PerWorkflow.CapabilityConcurrencyLimit.DefaultValue),
	}

	_, err = m.linkV2(t.Context(), m, store, exec)
	require.NoError(t, err, "V2 module should link successfully with V2 host function")
}

func TestLinkNoDAG_RegistersV1For2ParamImport(t *testing.T) {
	wasmBytes, err := wasmtime.Wat2Wasm(watCallCapV1)
	require.NoError(t, err)

	mc := defaultNoDAGModCfg(t)
	m, err := NewModule(t.Context(), mc, wasmBytes)
	require.NoError(t, err)
	require.Equal(t, callCapabilityV1ParamCount, m.callCapParams)

	// The module should link successfully with the V1 host function.
	mockExecHelper := mocks.NewMockExecutionHelper(t)

	store := wasmtime.NewStore(m.engine)
	exec := &execution[*sdkpb.ExecutionResult]{
		module:              m,
		ctx:                 t.Context(),
		executor:            mockExecHelper,
		capabilityResponses: map[int32]<-chan *sdkpb.CapabilityResponse{},
		usedCallbackIDs:     map[string]bool{},
		pendingCallsLimiter: limits.GlobalResourcePoolLimiter(cresettings.Default.PerWorkflow.CapabilityConcurrencyLimit.DefaultValue),
	}

	_, err = m.linkV2(t.Context(), m, store, exec)
	require.NoError(t, err, "V1 module should link successfully with V1 host function")
}
