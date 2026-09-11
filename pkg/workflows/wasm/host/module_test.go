package host

import (
	"bytes"
	"context"
	"math"
	"runtime/pprof"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/bytecodealliance/wasmtime-go/v47"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basictrigger"
	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/host/mocks"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/cresettings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/limits"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/matches"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func countLimiterUpdaterGoroutines(t *testing.T) int {
	t.Helper()

	var profile bytes.Buffer
	require.NoError(t, pprof.Lookup("goroutine").WriteTo(&profile, 2))
	count := 0
	for _, goroutine := range strings.Split(profile.String(), "\n\n") {
		if strings.Contains(goroutine, "pkg/settings/limits.(*updater") && strings.Contains(goroutine, ").updateLoop(") {
			count++
		}
	}
	return count
}

// This test inspects the process-wide goroutine profile and must remain serial.
func TestNewModuleClosesDefaultLimiters(t *testing.T) {
	t.Run("constructor error", func(t *testing.T) {
		before := countLimiterUpdaterGoroutines(t)
		cfg := &ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
		}

		_, err := NewModule(t.Context(), cfg, []byte("invalid wasm"))
		require.Error(t, err)
		require.Equal(t, before, countLimiterUpdaterGoroutines(t))
		require.Nil(t, cfg.PendingCallsLimiter)
		require.Nil(t, cfg.MemoryLimiter)
	})

	t.Run("module close", func(t *testing.T) {
		binary, err := wasmtime.Wat2Wasm(`(module (memory (export "memory") 1))`)
		require.NoError(t, err)

		before := countLimiterUpdaterGoroutines(t)
		mod, err := NewModule(t.Context(), &ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
		}, binary)
		require.NoError(t, err)
		require.Equal(t, before, countLimiterUpdaterGoroutines(t))
		free, err := limiterOrDefault(mod.cfg.PendingCallsLimiter, mod.defaultLimiters.pendingCalls).Wait(contexts.WithCRE(t.Context(), contexts.CRE{Workflow: "workflow-id"}), 1)
		require.NoError(t, err)
		free()
		require.Eventually(t, func() bool {
			return before+1 == countLimiterUpdaterGoroutines(t)
		}, time.Second, time.Millisecond)
		mod.Start()
		mod.Close()
		require.Eventually(t, func() bool {
			return before == countLimiterUpdaterGoroutines(t)
		}, time.Second, time.Millisecond)
	})

	t.Run("reused config", func(t *testing.T) {
		binary, err := wasmtime.Wat2Wasm(`(module (memory (export "memory") 1))`)
		require.NoError(t, err)
		cfg := &ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
		}
		before := countLimiterUpdaterGoroutines(t)

		for range 2 {
			mod, err := NewModule(t.Context(), cfg, binary)
			require.NoError(t, err)
			mod.Close()
			require.Nil(t, cfg.PendingCallsLimiter)
			require.Nil(t, cfg.MaxResponseSizeLimiter)
		}
		require.Equal(t, before, countLimiterUpdaterGoroutines(t))
	})

	t.Run("simultaneous modules sharing config", func(t *testing.T) {
		binary, err := wasmtime.Wat2Wasm(`(module (memory (export "memory") 1) (func (export "_start")))`)
		require.NoError(t, err)
		cfg := &ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
		}

		first, err := NewModule(t.Context(), cfg, binary)
		require.NoError(t, err)
		second, err := NewModule(t.Context(), cfg, binary)
		require.NoError(t, err)
		t.Cleanup(second.Close)

		first.Close()
		helper := mocks.NewMockExecutionHelper(t)
		helper.EXPECT().GetWorkflowExecutionID().Return("request-id")
		request := &sdkpb.ExecuteRequest{Request: &sdkpb.ExecuteRequest_Trigger{}}
		var runErr error
		var subscriptionErr error
		require.NotPanics(t, func() {
			_, runErr = second.Execute(t.Context(), request, helper)
			subscriptionErr = limiterOrDefault(second.cfg.MaxSubscriptionsLimiter, second.defaultLimiters.maxSubscriptions).Check(t.Context(), 1)
		})
		require.NoError(t, runErr)
		require.NoError(t, subscriptionErr)
	})

	t.Run("caller-provided limiter", func(t *testing.T) {
		binary, err := wasmtime.Wat2Wasm(`(module (memory (export "memory") 1))`)
		require.NoError(t, err)
		limiter := limits.NewUpperBoundLimiter(1)
		t.Cleanup(func() { require.NoError(t, limiter.Close()) })

		mod, err := NewModule(t.Context(), &ModuleConfig{
			Logger:                  logger.Test(t),
			IsUncompressed:          true,
			MaxSubscriptionsLimiter: limiter,
		}, binary)
		require.NoError(t, err)
		mod.Close()
		require.NoError(t, limiter.Check(t.Context(), 1))
	})

	t.Run("caller-replaced limiter", func(t *testing.T) {
		binary, err := wasmtime.Wat2Wasm(`(module (memory (export "memory") 1))`)
		require.NoError(t, err)
		cfg := &ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
		}
		mod, err := NewModule(t.Context(), cfg, binary)
		require.NoError(t, err)

		limiter := &closeTrackingGateLimiter{}
		cfg.EnableUserMetricsLimiter = limiter
		require.NoError(t, limiterOrDefault(mod.cfg.EnableUserMetricsLimiter, mod.defaultLimiters.enableUserMetrics).AllowErr(t.Context()))

		mod.Close()
		require.Same(t, limiter, cfg.EnableUserMetricsLimiter)
		require.Zero(t, limiter.closeCalls)
	})
}

type closeTrackingGateLimiter struct {
	closeCalls int
}

func (l *closeTrackingGateLimiter) Close() error {
	l.closeCalls++
	return nil
}

func (*closeTrackingGateLimiter) Limit(context.Context) (bool, error) {
	return true, nil
}

func (*closeTrackingGateLimiter) Open(context.Context) (bool, error) {
	return true, nil
}

func (*closeTrackingGateLimiter) AllowErr(context.Context) error {
	return nil
}

func Test_read(t *testing.T) {
	t.Run("successfully read from slice", func(t *testing.T) {
		memory := []byte("hello, world")
		got, err := read(memory, 0, int32(len(memory)))
		assert.NoError(t, err)
		assert.Equal(t, []byte("hello, world"), got)
	})

	t.Run("fail to read because out of bounds request", func(t *testing.T) {
		memory := []byte("hello, world")
		_, err := read(memory, 0, int32(len(memory)+1))
		assert.Error(t, err)
	})

	t.Run("fails to read because of invalid pointer or length", func(t *testing.T) {
		memory := []byte("hello, world")
		_, err := read(memory, 0, -1)
		assert.Error(t, err)

		_, err = read(memory, -1, 1)
		assert.Error(t, err)
	})

	t.Run("validate that memory is read only once copied", func(t *testing.T) {
		memory := []byte("hello, world")
		copied, err := read(memory, 0, int32(len(memory)))
		assert.NoError(t, err)

		// mutate copy
		copied[0] = 'H'
		assert.Equal(t, []byte("Hello, world"), copied)

		// original memory is unchanged
		assert.Equal(t, []byte("hello, world"), memory)
	})
}

func Test_write(t *testing.T) {
	t.Run("successfully write to slice", func(t *testing.T) {
		giveSrc := []byte("hello, world")
		memory := make([]byte, 12)
		n := write(memory, giveSrc, 0, int32(len(giveSrc)))
		assert.Equal(t, n, int64(len(giveSrc)))
		assert.Equal(t, []byte("hello, world"), memory[:len(giveSrc)])
	})

	t.Run("cannot write to slice because memory too small", func(t *testing.T) {
		giveSrc := []byte("hello, world")
		memory := make([]byte, len(giveSrc)-1)
		n := write(memory, giveSrc, 0, int32(len(giveSrc)))
		assert.Equal(t, int64(-1), n)
	})

	t.Run("fails to write to invalid access", func(t *testing.T) {
		giveSrc := []byte("hello, world")
		memory := make([]byte, len(giveSrc))
		n := write(memory, giveSrc, 0, -1)
		assert.Equal(t, int64(-1), n)

		n = write(memory, giveSrc, -1, 1)
		assert.Equal(t, int64(-1), n)
	})

	t.Run("truncated write due to size being smaller than len", func(t *testing.T) {
		giveSrc := []byte("hello, world")
		memory := make([]byte, 12)
		n := write(memory, giveSrc, 0, int32(len(giveSrc)-2))
		assert.Equal(t, int64(-1), n)
	})

	t.Run("unwanted data when size exceeds written data only writes the data", func(t *testing.T) {
		giveSrc := []byte("hello, world")
		memory := make([]byte, 20)
		n := write(memory, giveSrc, 0, 20)
		// TODO verify this won't break anything...
		assert.Equal(t, int64(12), n)
	})

	t.Run("near-MaxInt32 ptr does not overflow the bounds check", func(t *testing.T) {
		giveSrc := []byte("12345678")
		memory := make([]byte, 12)
		// ptr+maxSize would wrap negative with signed int32 arithmetic.
		n := write(memory, giveSrc, math.MaxInt32-4, int32(len(giveSrc)))
		assert.Equal(t, int64(-1), n)
	})

	t.Run("negative maxSize is rejected", func(t *testing.T) {
		memory := make([]byte, 12)
		n := write(memory, nil, 0, math.MinInt32)
		assert.Equal(t, int64(-1), n)
	})

	t.Run("zero-length write at end of memory is allowed", func(t *testing.T) {
		memory := make([]byte, 12)
		n := write(memory, nil, int32(len(memory)), 0)
		assert.Equal(t, int64(0), n)
	})
}

func Test_SdkLabeler(t *testing.T) {
	t.Run("defaults to no-op when nil", func(t *testing.T) {
		// ModuleConfig with nil SdkLabeler should not panic when creating a module
		binary := createTestBinary(stdioBinaryCmd, stdioBinaryLocation, true, t)
		mc := &ModuleConfig{
			Logger:         logger.Test(t),
			IsUncompressed: true,
		}
		_, err := NewModule(t.Context(), mc, binary)
		require.NoError(t, err)
		require.NotNil(t, mc.SdkLabeler, "SdkLabeler should be set to no-op")
	})

	t.Run("is called with v2ImportName after discovery", func(t *testing.T) {
		binary := createTestBinary(nodagRandomBinaryCmd, nodagRandomBinaryLocation, true, t)
		var capturedName string
		mc := defaultNoDAGModCfg(t)
		mc.SdkLabeler = func(name string) {
			capturedName = name
		}
		_, err := NewModule(t.Context(), mc, binary)
		require.NoError(t, err)
		require.NotEmpty(t, capturedName, "SdkLabeler should have been called with v2 import name")
		require.True(t, strings.HasPrefix(capturedName, v2ImportPrefix), "captured name should have v2 prefix")
	})
}

// CallAwaitRace validates that every call can be awaited.
func Test_CallAwaitRace(t *testing.T) {
	ctx := t.Context()
	mockExecHelper := mocks.NewMockExecutionHelper(t)
	mockExecHelper.EXPECT().
		CallCapability(matches.AnyContext, mock.Anything).
		Return(&sdkpb.CapabilityResponse{}, nil)

	m := &module{}

	var wg sync.WaitGroup
	var wantAttempts = 100

	exec := &execution[*sdkpb.ExecutionResult]{
		module:              m,
		capabilityResponses: map[int32]<-chan *sdkpb.CapabilityResponse{},
		usedCallbackIDs:     map[string]bool{},
		pendingCallsLimiter: limits.GlobalResourcePoolLimiter(cresettings.Default.PerWorkflow.CapabilityConcurrencyLimit.DefaultValue),
		ctx:                 t.Context(),
		executor:            mockExecHelper,
	}

	wg.Add(wantAttempts)
	for on := range wantAttempts {
		go func() {
			defer wg.Done()
			// call
			err := exec.callCapAsync(ctx, &sdkpb.CapabilityRequest{
				Id:         "test-cap-request",
				CallbackId: int32(on),
			})
			require.NoError(t, err)

			// await with id
			_, err = exec.awaitCapabilities(ctx, &sdkpb.AwaitCapabilitiesRequest{
				Ids: []int32{int32(on)},
			})
			require.NoError(t, err)
		}()
	}

	wg.Wait()
}

type fakeModuleMetrics struct {
	hostFnPanicRecoveredCount int
}

func (f *fakeModuleMetrics) IncHostFnPanicRecovered() {
	f.hostFnPanicRecoveredCount++
}

func Test_Integration_Execute_RecoversHostFunctionPanic(t *testing.T) {
	lggr, logs := logger.TestObserved(t, zapcore.ErrorLevel)
	cfg := defaultNoDAGModCfg(t)
	cfg.Logger = lggr

	// The logging module calls env.log, whose host callback reaches EmitUserLog
	// synchronously on the goroutine running the guest. Using a real module keeps
	// the whole link and execute path in the test instead of a stub linker.
	m := makeTestModuleByName(t, testPath, "logging", cfg, true)

	m.Start()
	defer m.Close()

	metrics := &fakeModuleMetrics{}
	m.metrics = metrics

	mockExecutionHelper := mocks.NewMockExecutionHelper(t)
	mockExecutionHelper.EXPECT().GetWorkflowExecutionID().Return("test-execution-id")
	mockExecutionHelper.EXPECT().GetNodeTime().RunAndReturn(func() time.Time {
		return time.Now()
	}).Maybe()
	mockExecutionHelper.EXPECT().GetDONTime().RunAndReturn(func() (time.Time, error) {
		return time.Now(), nil
	}).Maybe()
	mockExecutionHelper.EXPECT().EmitUserLog(mock.Anything).RunAndReturn(func(string) error {
		panic("boom")
	}).Once()

	trigger := &basictrigger.Outputs{CoolOutput: anyTestTriggerValue}
	_, err := m.Execute(t.Context(), triggerExecuteRequest(t, 0, trigger), mockExecutionHelper)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
	assert.Equal(t, 1, metrics.hostFnPanicRecoveredCount)

	require.Len(t, logs.AllUntimed(), 1)
	assert.Equal(t, zapcore.ErrorLevel, logs.AllUntimed()[0].Level)
	assert.Equal(t, "panic during wasm execution", logs.AllUntimed()[0].Message)
}
