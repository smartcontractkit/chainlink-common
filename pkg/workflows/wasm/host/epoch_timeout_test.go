package host

import (
	"context"
	"testing"
	"time"

	"github.com/bytecodealliance/wasmtime-go/v47"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/host/mocks"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func TestModuleEpochDeadline(t *testing.T) {
	// Manual ticks reproduce an interrupt before the wall-clock deadline without
	// depending on scheduler timing. This test must not run in parallel.
	engine := newEngine(logger.Test(t))
	globalEngineMu.Lock()
	previous := globalEngine
	globalEngine = engine
	globalEngineMu.Unlock()
	t.Cleanup(func() {
		globalEngineMu.Lock()
		globalEngine = previous
		globalEngineMu.Unlock()
		engine.Close()
	})
	for range 17 {
		engine.IncrementEpoch()
	}

	for _, tc := range []struct {
		name    string
		timeout time.Duration
		ticks   int
		body    string
		want    string
	}{
		{"fractional budget before deadline", 999 * time.Millisecond, 9, "", "success"},
		{"fractional budget at deadline", 999 * time.Millisecond, 10, "", "timeout"},
		{"whole ticks before deadline", time.Second, 9, "", "success"},
		{"whole ticks at deadline", time.Second, 10, "", "timeout"},
		{"sub-tick budget before deadline", 50 * time.Millisecond, 0, "", "success"},
		{"sub-tick budget at deadline", 50 * time.Millisecond, 1, "", "timeout"},
		{"unrelated trap", 50 * time.Millisecond, 0, "unreachable", "unreachable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := wasmtime.Wat2Wasm(`(module
  (import "env" "version_v2" (func))
  (memory (export "memory") 1)
  (func (export "_start") ` + tc.body + `))`)
			require.NoError(t, err)
			mod, err := NewModule(t.Context(), &ModuleConfig{
				Logger: logger.Test(t), IsUncompressed: true, Timeout: &tc.timeout,
			}, raw)
			require.NoError(t, err)
			defer mod.Close()
			helper := mocks.NewMockExecutionHelper(t)
			helper.EXPECT().GetWorkflowExecutionID().RunAndReturn(func() string {
				// The store's relative deadline is armed before this callback.
				for range tc.ticks {
					engine.IncrementEpoch()
				}
				return "id"
			})
			result, err := mod.Execute(t.Context(), &sdkpb.ExecuteRequest{}, helper)
			require.Nil(t, result)
			switch tc.want {
			case "success":
				require.NoError(t, err)
			case "timeout":
				require.ErrorIs(t, err, context.DeadlineExceeded)
			case "unreachable":
				var trap *wasmtime.Trap
				require.ErrorAs(t, err, &trap)
				require.NotNil(t, trap.Code())
				require.Equal(t, wasmtime.UnreachableCodeReached, *trap.Code())
			default:
				t.Fatalf("unknown expected outcome %q", tc.want)
			}
		})
	}
}
