package host

import (
	"strconv"
	"testing"

	"github.com/bytecodealliance/wasmtime-go/v47"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/limits"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/host/mocks"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func TestModule_MaxTableElementsLimiter(t *testing.T) {
	growTable := func(t *testing.T, declared, grow int) []byte {
		binary, err := wasmtime.Wat2Wasm(`(module
  (memory (export "memory") 1)
  (table $t ` + strconv.Itoa(declared) + ` funcref)
  (func (export "_start")
    (if (i32.eq (table.grow $t (ref.null func) (i32.const ` + strconv.Itoa(grow) + `)) (i32.const -1))
      (then unreachable))))`)
		require.NoError(t, err)
		return binary
	}

	run := func(t *testing.T, binary []byte, limiter limits.BoundLimiter[int]) error {
		mod, err := NewModule(t.Context(), &ModuleConfig{
			Logger:                  logger.Test(t),
			IsUncompressed:          true,
			MaxTableElementsLimiter: limiter,
		}, binary)
		require.NoError(t, err)
		t.Cleanup(mod.Close)

		helper := mocks.NewMockExecutionHelper(t)
		helper.EXPECT().GetWorkflowExecutionID().Return("id")
		_, err = mod.Execute(t.Context(), &sdkpb.ExecuteRequest{Request: &sdkpb.ExecuteRequest_Trigger{}}, helper)
		return err
	}

	t.Run("default limit", func(t *testing.T) {
		require.NoError(t, run(t, growTable(t, 0, 1_000_000), nil))
		require.ErrorContains(t, run(t, growTable(t, 0, 1_000_001), nil), "unreachable")
	})

	t.Run("growth beyond limit fails", func(t *testing.T) {
		err := run(t, growTable(t, 0, 1_001), limits.NewUpperBoundLimiter(1_000))
		require.ErrorContains(t, err, "unreachable")
	})

	t.Run("growth within limit succeeds", func(t *testing.T) {
		require.NoError(t, run(t, growTable(t, 0, 1_000), limits.NewUpperBoundLimiter(1_000)))
	})

	t.Run("declared size beyond limit fails instantiation", func(t *testing.T) {
		err := run(t, growTable(t, 1_001, 0), limits.NewUpperBoundLimiter(1_000))
		require.ErrorContains(t, err, "error linking wasm")
	})
}
