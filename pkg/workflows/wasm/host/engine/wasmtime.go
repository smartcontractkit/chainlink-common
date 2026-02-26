package engine

import (
	"fmt"
	"strings"

	"github.com/bytecodealliance/wasmtime-go/v28"
)

const v2ImportPrefix = "version_v2"

// wasmtimeEngine implements Engine using the wasmtime runtime.
type wasmtimeEngine struct{}

func (e *wasmtimeEngine) Load(binary []byte, cfg LoadConfig) (Runtime, error) {
	wcfg := wasmtime.NewConfig()
	wcfg.SetEpochInterruption(true)
	if cfg.InitialFuel > 0 {
		wcfg.SetConsumeFuel(true)
	}
	wcfg.CacheConfigLoadDefault()
	wcfg.SetCraneliftOptLevel(wasmtime.OptLevelSpeedAndSize)
	setUnwinding(wcfg)

	eng := wasmtime.NewEngineWithConfig(wcfg)

	mod, err := wasmtime.NewModule(eng, binary)
	if err != nil {
		return nil, fmt.Errorf("error creating wasmtime module: %w", err)
	}

	v2ImportName := ""
	for _, modImport := range mod.Imports() {
		name := modImport.Name()
		if modImport.Module() == "env" && name != nil && strings.HasPrefix(*name, v2ImportPrefix) {
			v2ImportName = *name
			break
		}
	}

	return &wasmtimeRuntime{
		engine:       eng,
		module:       mod,
		config:       wcfg,
		v2ImportName: v2ImportName,
	}, nil
}

// wasmtimeRuntime holds a compiled wasmtime module and its engine.
type wasmtimeRuntime struct {
	engine       *wasmtime.Engine
	module       *wasmtime.Module
	config       *wasmtime.Config
	v2ImportName string
}

func (r *wasmtimeRuntime) V2ImportName() string { return r.v2ImportName }

func (r *wasmtimeRuntime) NewStore() Store {
	return &wasmtimeStore{
		store:   wasmtime.NewStore(r.engine),
		runtime: r,
	}
}

func (r *wasmtimeRuntime) IncrementEpoch() { r.engine.IncrementEpoch() }

func (r *wasmtimeRuntime) Close() {
	r.engine.Close()
	r.module.Close()
	r.config.Close()
}

// wasmtimeStore wraps a wasmtime.Store and implements Store.
type wasmtimeStore struct {
	store   *wasmtime.Store
	runtime *wasmtimeRuntime
}

func (s *wasmtimeStore) SetWasi(argv []string) {
	wasi := wasmtime.NewWasiConfig()
	wasi.InheritStdout()
	wasi.SetArgv(argv)
	s.store.SetWasi(wasi)
}

func (s *wasmtimeStore) SetFuel(fuel uint64) error {
	return s.store.SetFuel(fuel)
}

func (s *wasmtimeStore) SetLimiter(memoryBytes, tableElements, instances, tables, memories int64) {
	s.store.Limiter(memoryBytes, tableElements, instances, tables, memories)
}

func (s *wasmtimeStore) SetEpochDeadline(deadline uint64) {
	s.store.SetEpochDeadline(deadline)
}

func (s *wasmtimeStore) LinkNoDAG(exec *Execution) (Instance, error) {
	linker := wasmtime.NewLinker(s.runtime.engine)
	linker.AllowShadowing(true)

	if err := linker.DefineWasi(); err != nil {
		return nil, err
	}

	if err := linker.FuncWrap(
		"wasi_snapshot_preview1", "poll_oneoff",
		wrapPollOneoff(exec.PollOneoff),
	); err != nil {
		return nil, err
	}

	if err := linker.FuncWrap(
		"wasi_snapshot_preview1", "clock_time_get",
		wrapClockTimeGet(exec.ClockTimeGet),
	); err != nil {
		return nil, err
	}

	if err := linker.FuncWrap(
		"env", exec.V2ImportName,
		func(caller *wasmtime.Caller) {},
	); err != nil {
		return nil, fmt.Errorf("error wrapping v2 import func: %w", err)
	}

	if err := linker.FuncWrap("env", "send_response",
		wrapSendResponse(exec.SendResponse),
	); err != nil {
		return nil, fmt.Errorf("error wrapping sendResponse func: %w", err)
	}

	if err := linker.FuncWrap("env", "call_capability",
		wrapCallCapability(exec.CallCapability),
	); err != nil {
		return nil, fmt.Errorf("error wrapping callcap func: %w", err)
	}

	if err := linker.FuncWrap("env", "await_capabilities",
		wrapAwaitCapabilities(exec.AwaitCapabilities),
	); err != nil {
		return nil, fmt.Errorf("error wrapping awaitcaps func: %w", err)
	}

	if err := linker.FuncWrap("env", "get_secrets",
		wrapGetSecrets(exec.GetSecrets),
	); err != nil {
		return nil, fmt.Errorf("error wrapping get_secrets func: %w", err)
	}

	if err := linker.FuncWrap("env", "await_secrets",
		wrapAwaitSecrets(exec.AwaitSecrets),
	); err != nil {
		return nil, fmt.Errorf("error wrapping await_secrets func: %w", err)
	}

	if err := linker.FuncWrap("env", "log",
		wrapLog(exec.Log),
	); err != nil {
		return nil, fmt.Errorf("error wrapping log func: %w", err)
	}

	if err := linker.FuncWrap("env", "switch_modes",
		wrapSwitchModes(exec.SwitchModes),
	); err != nil {
		return nil, fmt.Errorf("error wrapping switchModes func: %w", err)
	}

	if err := linker.FuncWrap("env", "random_seed", exec.RandomSeed); err != nil {
		return nil, fmt.Errorf("error wrapping getSeed func: %w", err)
	}

	if err := linker.FuncWrap("env", "now",
		wrapNow(exec.Now),
	); err != nil {
		return nil, fmt.Errorf("error wrapping now func: %w", err)
	}

	inst, err := linker.Instantiate(s.store, s.runtime.module)
	if err != nil {
		return nil, err
	}
	return &wasmtimeInstance{instance: inst, store: s.store}, nil
}

func (s *wasmtimeStore) LinkLegacyDAG(exec *LegacyExecution) (Instance, error) {
	linker := wasmtime.NewLinker(s.runtime.engine)
	linker.AllowShadowing(true)

	if err := linker.DefineWasi(); err != nil {
		return nil, err
	}

	if err := linker.FuncWrap(
		"wasi_snapshot_preview1", "poll_oneoff",
		wrapPollOneoff(exec.PollOneoff),
	); err != nil {
		return nil, err
	}

	if err := linker.FuncWrap(
		"wasi_snapshot_preview1", "clock_time_get",
		wrapClockTimeGet(exec.ClockTimeGet),
	); err != nil {
		return nil, err
	}

	if exec.RandomGet != nil {
		if err := linker.FuncWrap(
			"wasi_snapshot_preview1", "random_get",
			wrapRandomGet(exec.RandomGet),
		); err != nil {
			return nil, err
		}
	}

	if err := linker.FuncWrap("env", "sendResponse",
		wrapSendResponse(exec.SendResponse),
	); err != nil {
		return nil, fmt.Errorf("error wrapping sendResponse func: %w", err)
	}

	if err := linker.FuncWrap("env", "fetch",
		wrapFetch(exec.Fetch),
	); err != nil {
		return nil, fmt.Errorf("error wrapping fetch func: %w", err)
	}

	if err := linker.FuncWrap("env", "emit",
		wrapEmit(exec.Emit),
	); err != nil {
		return nil, fmt.Errorf("error wrapping emit func: %w", err)
	}

	if err := linker.FuncWrap("env", "log",
		wrapLog(exec.Log),
	); err != nil {
		return nil, fmt.Errorf("error wrapping log func: %w", err)
	}

	inst, err := linker.Instantiate(s.store, s.runtime.module)
	if err != nil {
		return nil, err
	}
	return &wasmtimeInstance{instance: inst, store: s.store}, nil
}

func (s *wasmtimeStore) Close() { s.store.Close() }

// wasmtimeInstance wraps a wasmtime.Instance.
type wasmtimeInstance struct {
	instance *wasmtime.Instance
	store    *wasmtime.Store
}

func (i *wasmtimeInstance) CallStart() error {
	start := i.instance.GetFunc(i.store, "_start")
	if start == nil {
		return fmt.Errorf("could not get start function")
	}
	_, err := start.Call(i.store)
	return err
}

// wasmtimeCallerAdapter wraps *wasmtime.Caller to implement MemoryAccessor.
type wasmtimeCallerAdapter struct {
	c *wasmtime.Caller
}

func (a *wasmtimeCallerAdapter) Memory() []byte {
	return a.c.GetExport("memory").Memory().UnsafeData(a.c)
}

// Wrapper functions that convert *wasmtime.Caller to MemoryAccessor and
// delegate to the engine-agnostic closures.

func wrapSendResponse(fn func(MemoryAccessor, int32, int32) int32) func(*wasmtime.Caller, int32, int32) int32 {
	return func(caller *wasmtime.Caller, ptr, ptrlen int32) int32 {
		return fn(&wasmtimeCallerAdapter{caller}, ptr, ptrlen)
	}
}

func wrapCallCapability(fn func(MemoryAccessor, int32, int32) int64) func(*wasmtime.Caller, int32, int32) int64 {
	return func(caller *wasmtime.Caller, ptr, ptrlen int32) int64 {
		return fn(&wasmtimeCallerAdapter{caller}, ptr, ptrlen)
	}
}

func wrapAwaitCapabilities(fn func(MemoryAccessor, int32, int32, int32, int32) int64) func(*wasmtime.Caller, int32, int32, int32, int32) int64 {
	return func(caller *wasmtime.Caller, a, b, c, d int32) int64 {
		return fn(&wasmtimeCallerAdapter{caller}, a, b, c, d)
	}
}

func wrapGetSecrets(fn func(MemoryAccessor, int32, int32, int32, int32) int64) func(*wasmtime.Caller, int32, int32, int32, int32) int64 {
	return func(caller *wasmtime.Caller, a, b, c, d int32) int64 {
		return fn(&wasmtimeCallerAdapter{caller}, a, b, c, d)
	}
}

func wrapAwaitSecrets(fn func(MemoryAccessor, int32, int32, int32, int32) int64) func(*wasmtime.Caller, int32, int32, int32, int32) int64 {
	return func(caller *wasmtime.Caller, a, b, c, d int32) int64 {
		return fn(&wasmtimeCallerAdapter{caller}, a, b, c, d)
	}
}

func wrapLog(fn func(MemoryAccessor, int32, int32)) func(*wasmtime.Caller, int32, int32) {
	return func(caller *wasmtime.Caller, ptr, ptrlen int32) {
		fn(&wasmtimeCallerAdapter{caller}, ptr, ptrlen)
	}
}

func wrapSwitchModes(fn func(MemoryAccessor, int32)) func(*wasmtime.Caller, int32) {
	return func(caller *wasmtime.Caller, mode int32) {
		fn(&wasmtimeCallerAdapter{caller}, mode)
	}
}

func wrapNow(fn func(MemoryAccessor, int32) int32) func(*wasmtime.Caller, int32) int32 {
	return func(caller *wasmtime.Caller, resultTs int32) int32 {
		return fn(&wasmtimeCallerAdapter{caller}, resultTs)
	}
}

func wrapPollOneoff(fn func(MemoryAccessor, int32, int32, int32, int32) int32) func(*wasmtime.Caller, int32, int32, int32, int32) int32 {
	return func(caller *wasmtime.Caller, a, b, c, d int32) int32 {
		return fn(&wasmtimeCallerAdapter{caller}, a, b, c, d)
	}
}

func wrapClockTimeGet(fn func(MemoryAccessor, int32, int64, int32) int32) func(*wasmtime.Caller, int32, int64, int32) int32 {
	return func(caller *wasmtime.Caller, id int32, precision int64, resultTs int32) int32 {
		return fn(&wasmtimeCallerAdapter{caller}, id, precision, resultTs)
	}
}

func wrapFetch(fn func(MemoryAccessor, int32, int32, int32, int32) int32) func(*wasmtime.Caller, int32, int32, int32, int32) int32 {
	return func(caller *wasmtime.Caller, a, b, c, d int32) int32 {
		return fn(&wasmtimeCallerAdapter{caller}, a, b, c, d)
	}
}

func wrapEmit(fn func(MemoryAccessor, int32, int32, int32, int32) int32) func(*wasmtime.Caller, int32, int32, int32, int32) int32 {
	return func(caller *wasmtime.Caller, a, b, c, d int32) int32 {
		return fn(&wasmtimeCallerAdapter{caller}, a, b, c, d)
	}
}

func wrapRandomGet(fn func(MemoryAccessor, int32, int32) int32) func(*wasmtime.Caller, int32, int32) int32 {
	return func(caller *wasmtime.Caller, buf, bufLen int32) int32 {
		return fn(&wasmtimeCallerAdapter{caller}, buf, bufLen)
	}
}
