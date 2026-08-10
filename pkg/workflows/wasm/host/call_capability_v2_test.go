package host

import (
	"context"
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

// callCapabilityV2ParamCount is the number of params the V2 call_capability
// import declares (req, reqLen, responseBuffer, maxResponseLen).
const callCapabilityV2ParamCount = 4

// callCapabilityV1ParamCount is the number of params the legacy V1
// call_capability import declares (req, reqLen).
const callCapabilityV1ParamCount = 2

// --- WAT helpers ---

// watCallCapV2 is a minimal WAT module that imports call_capability with 4
// params (i32, i32, i32, i32) → i64, plus a version_v2 import so the host
// treats it as a NoDAG module.
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
// a NoDAG module.
const watCallCapV1 = `
	(module
	  (import "env" "version_v2" (func $version_v2))
	  (import "env" "call_capability" (func $call_capability (param i32 i32) (result i64)))
	  (memory (export "memory") 1)
	  (func (export "_start")
	    call $version_v2)
	)`

const watCallCapV2NoVersion = `
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

const watCallCapV1NoVersion = `
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

func instantiateCallCapModule(t *testing.T, wat string, hostFn interface{}) (*wasmtime.Store, *wasmtime.Instance, *wasmtime.Memory) {
	t.Helper()
	wasmBytes, err := wasmtime.Wat2Wasm(wat)
	require.NoError(t, err)

	engine := wasmtime.NewEngine()
	mod, err := wasmtime.NewModule(engine, wasmBytes)
	require.NoError(t, err)

	store := wasmtime.NewStore(engine)
	linker := wasmtime.NewLinker(engine)
	require.NoError(t, linker.FuncWrap("env", "call_capability", hostFn))

	inst, err := linker.Instantiate(store, mod)
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

// --- V2 host function tests ---

func TestCreateCallCapFnV2_CallCapAsyncErrorWritesToResponseBuffer(t *testing.T) {
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

	fn := createCallCapFnV2(lggr, exec)

	store, inst, mem := instantiateCallCapModule(t, watCallCapV2NoVersion, fn)

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
	assert.Contains(t, errStr, "callCapAsync",
		"V2 should write detailed error to response buffer")

	// Verify the error was logged.
	require.Len(t, logs.AllUntimed(), 1)
	assert.Equal(t, zapcore.ErrorLevel, logs.AllUntimed()[0].Level)
}

func TestCreateCallCapFnV2_SuccessReturnsZero(t *testing.T) {
	lggr := logger.Test(t)
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

	fn := createCallCapFnV2(lggr, exec)

	store, inst, mem := instantiateCallCapModule(t, watCallCapV2NoVersion, fn)

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

func TestCreateCallCapFnV2_ProtoUnmarshalErrorWritesToResponseBuffer(t *testing.T) {
	lggr, logs := logger.TestObserved(t, zapcore.ErrorLevel)
	mockExecHelper := mocks.NewMockExecutionHelper(t)

	exec := &execution[*sdkpb.ExecutionResult]{
		capabilityResponses: map[int32]<-chan *sdkpb.CapabilityResponse{},
		usedCallbackIDs:     map[string]bool{},
		pendingCallsLimiter: limits.GlobalResourcePoolLimiter(cresettings.Default.PerWorkflow.CapabilityConcurrencyLimit.DefaultValue),
		ctx:                 t.Context(),
		executor:            mockExecHelper,
	}

	fn := createCallCapFnV2(lggr, exec)

	store, inst, mem := instantiateCallCapModule(t, watCallCapV2NoVersion, fn)

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

// --- V1 host function tests (verify unchanged behavior) ---

func TestCreateCallCapFnV1_CallCapAsyncErrorReturnsBareMinusOne(t *testing.T) {
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

	fn := createCallCapFnV1(lggr, exec)

	store, inst, mem := instantiateCallCapModule(t, watCallCapV1NoVersion, fn)

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
	assert.Equal(t, int64(-1), resultI64,
		"V1 should return bare -1 on callCapAsync error (existing behavior)")

	// Verify error was logged.
	require.Len(t, logs.AllUntimed(), 1)
	assert.Equal(t, zapcore.ErrorLevel, logs.AllUntimed()[0].Level)
}

func TestCreateCallCapFnV1_SuccessReturnsZero(t *testing.T) {
	lggr := logger.Test(t)
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

	fn := createCallCapFnV1(lggr, exec)

	store, inst, mem := instantiateCallCapModule(t, watCallCapV1NoVersion, fn)

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

// Ensure context import is used.
var _ = context.TODO
