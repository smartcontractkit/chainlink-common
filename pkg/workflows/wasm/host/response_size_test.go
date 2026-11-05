package host

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/bytecodealliance/wasmtime-go/v50"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/cresettings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/limits"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/host/mocks"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

const (
	// validResponse is three unknown varint fields (field 1000), a valid 9 byte ExecutionResult.
	validResponse = `\c0\3e\01\c0\3e\01\c0\3e\01`
	// invalidResponse is 9 bytes that fail proto.Unmarshal.
	invalidResponse = `\ff\ff\ff\ff\ff\ff\ff\ff\ff`
	responseLen     = 9
)

// sendResponseGuestWAT returns a minimal guest, in WebAssembly text format, that sends length bytes of data via send_response and exits.
func sendResponseGuestWAT(data string, length int) string {
	return fmt.Sprintf(`
	(module
	  ;; host functions the guest calls
	  (import "env" "version_v2" (func $version_v2))
	  (import "env" "send_response" (func $send_response (param i32 i32) (result i32)))
	  (import "wasi_snapshot_preview1" "proc_exit" (func $proc_exit (param i32)))
	  ;; one 64 KiB memory page, exported so the host can read it
	  (memory (export "memory") 1)
	  ;; the response bytes, placed at address 0
	  (data (i32.const 0) "%s")
	  (func (export "_start")
	    ;; declare the v2 ABI, as the SDK does at startup
	    call $version_v2
	    ;; send_response(ptr=0, len=length), ignoring its return code
	    (drop (call $send_response (i32.const 0) (i32.const %d)))
	    ;; exit cleanly
	    (call $proc_exit (i32.const 0)))
	)`, data, length)
}

type errGateLimiter struct{}

func (errGateLimiter) Close() error                         { return nil }
func (errGateLimiter) Limit(context.Context) (bool, error)  { return false, errors.New("unreadable") }
func (errGateLimiter) IsOpen(context.Context) (bool, error) { return false, errors.New("unreadable") }
func (errGateLimiter) AllowErr(context.Context) error       { return errors.New("unreadable") }

func newResponseSizeModule(t *testing.T, data string, cfg *ModuleConfig) (*module, *fakeModuleMetrics) {
	t.Helper()
	binary, err := wasmtime.Wat2Wasm(sendResponseGuestWAT(data, responseLen))
	require.NoError(t, err)
	m, err := NewModule(t.Context(), cfg, binary)
	require.NoError(t, err)
	t.Cleanup(m.Close)
	metrics := &fakeModuleMetrics{}
	m.metrics = metrics
	return m, metrics
}

func executeResponseSizeModule(t *testing.T, m *module) (*sdkpb.ExecutionResult, error) {
	t.Helper()
	helper := mocks.NewMockExecutionHelper(t)
	helper.EXPECT().GetWorkflowExecutionID().Return("execution-id")
	helper.EXPECT().GetNodeTime().Return(time.Now()).Maybe()
	return m.Execute(t.Context(), &sdkpb.ExecuteRequest{Request: &sdkpb.ExecuteRequest_Trigger{}}, helper)
}

func responseSizeCfg(t *testing.T, lggr logger.Logger, limit config.Size, enforce limits.GateLimiter) *ModuleConfig {
	cfg := defaultNoDAGModCfg(t)
	cfg.Logger = lggr
	if limit > 0 {
		cfg.MaxResponseSizeLimiter = limits.NewUpperBoundLimiter(limit)
	}
	cfg.EnforceResponseSizeLimiter = enforce
	return cfg
}

func TestSendResponse_ResponseSize(t *testing.T) {
	t.Run("enforcement setting defaults to off", func(t *testing.T) {
		assert.False(t, cresettings.Default.PerWorkflow.ExecutionResponseLimitEnforced.DefaultValue)

		m, _ := newResponseSizeModule(t, validResponse, defaultNoDAGModCfg(t))
		open, err := m.defaultLimiters.enforceResponseSize.IsOpen(t.Context())
		require.NoError(t, err)
		assert.False(t, open)
	})

	t.Run("under limit records size", func(t *testing.T) {
		m, metrics := newResponseSizeModule(t, validResponse, responseSizeCfg(t, logger.Test(t), 0, limits.NewGateLimiter(true)))

		result, err := executeResponseSizeModule(t, m)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, []int64{responseLen}, metrics.responseSizes)
		assert.Empty(t, metrics.responsesOverLimit)
	})

	t.Run("over limit not enforced keeps response and counts it", func(t *testing.T) {
		m, metrics := newResponseSizeModule(t, validResponse, responseSizeCfg(t, logger.Test(t), 4, nil))

		result, err := executeResponseSizeModule(t, m)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, []int64{responseLen}, metrics.responseSizes)
		assert.Equal(t, map[bool]int{false: 1}, metrics.responsesOverLimit)
	})

	t.Run("over limit enforced rejects without unmarshalling", func(t *testing.T) {
		lggr, logs := logger.TestObserved(t, zapcore.DebugLevel)
		m, metrics := newResponseSizeModule(t, invalidResponse, responseSizeCfg(t, lggr, 4, limits.NewGateLimiter(true)))

		result, err := executeResponseSizeModule(t, m)
		require.EqualError(t, err, "workflow response of 9 bytes exceeds limit of 4 bytes")
		assert.Nil(t, result)
		assert.Equal(t, map[bool]int{true: 1}, metrics.responsesOverLimit)
		for _, entry := range logs.AllUntimed() {
			assert.False(t, strings.Contains(entry.Message, "error calling sendResponse"), entry.Message)
		}
	})

	t.Run("unreadable setting is not enforced and logged each time", func(t *testing.T) {
		lggr, logs := logger.TestObserved(t, zapcore.ErrorLevel)
		m, metrics := newResponseSizeModule(t, validResponse, responseSizeCfg(t, lggr, 4, errGateLimiter{}))

		for range 2 {
			result, err := executeResponseSizeModule(t, m)
			require.NoError(t, err)
			require.NotNil(t, result)
		}
		assert.Equal(t, map[bool]int{false: 2}, metrics.responsesOverLimit)
		assert.Equal(t, 2, logs.FilterMessageSnippet("response size enforcement setting").Len())
	})
}
