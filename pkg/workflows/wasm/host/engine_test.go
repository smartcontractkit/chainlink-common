package host

import (
	"math"
	"testing"
	"time"

	"github.com/bytecodealliance/wasmtime-go/v47"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// spinWat exports a function that busy-loops n times, giving the epoch ticker
// something to interrupt.
const spinWat = `
(module
  (func (export "spin") (param $n i64)
    (loop $l
      (local.set $n (i64.sub (local.get $n) (i64.const 1)))
      (br_if $l (i64.gt_s (local.get $n) (i64.const 0))))))
`

// newSpinner compiles spinWat on e and returns a store armed to trap one epoch
// from now.
func newSpinner(t *testing.T, e *engine) (*wasmtime.Store, *wasmtime.Func) {
	t.Helper()

	wasmBytes, err := wasmtime.Wat2Wasm(spinWat)
	require.NoError(t, err)

	mod, err := wasmtime.NewModule(e.Engine, wasmBytes)
	require.NoError(t, err)
	t.Cleanup(mod.Close)

	store := wasmtime.NewStore(e.Engine)
	t.Cleanup(store.Close)
	store.SetEpochDeadline(1)

	linker := wasmtime.NewLinker(e.Engine)
	defer linker.Close()

	inst, err := linker.Instantiate(store, mod)
	require.NoError(t, err)

	return store, inst.GetFunc(store, "spin")
}

func TestEngine_StartInterruptsRunningCode(t *testing.T) {
	e := newEngine(logger.Test(t))
	// Registered before anything built from the engine, so it is torn down last.
	t.Cleanup(e.close)
	e.engineTickInterval = time.Millisecond
	e.start()

	store, spin := newSpinner(t, e)

	_, err := spin.Call(store, int64(math.MaxInt64))
	require.ErrorContains(t, err, "interrupt")
}

func TestEngine_EpochOnlyAdvancesWhileStarted(t *testing.T) {
	e := newEngine(logger.Test(t))
	t.Cleanup(e.close)

	store, spin := newSpinner(t, e)

	// Never started, so nothing advances the epoch and the armed deadline is
	// never reached, however long the guest runs.
	_, err := spin.Call(store, int64(500_000_000))
	require.NoError(t, err)

	// The deadline really was armed: one tick's worth of progress trips it.
	e.IncrementEpoch()
	_, err = spin.Call(store, int64(500_000_000))
	require.ErrorContains(t, err, "interrupt")
}

func TestEngine_CloseStopsTicker(t *testing.T) {
	e := newEngine(logger.Test(t))
	e.engineTickInterval = time.Millisecond
	e.start()

	time.Sleep(10 * time.Millisecond) // let the ticker run a few times
	e.close()

	// A ticker outliving close would call IncrementEpoch on the deallocated
	// engine and panic the test binary.
	time.Sleep(20 * time.Millisecond)

	require.Panics(t, func() { e.IncrementEpoch() }, "engine should be deallocated")
}
