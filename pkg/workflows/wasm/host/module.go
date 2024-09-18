package host

import (
	"encoding/base64"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"
	"unsafe"

	"github.com/bytecodealliance/wasmtime-go/v23"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

func safeMem(caller *wasmtime.Caller, ptr unsafe.Pointer, size int32) ([]byte, error) {
	mem := caller.GetExport("memory").Memory()
	data := mem.UnsafeData(caller)
	iptr := int32(uintptr(ptr))
	if iptr+size > int32(len(data)) {
		return nil, errors.New("out of bounds memory access")
	}

	cd := make([]byte, size)
	copy(cd, data[iptr:iptr+size])
	return cd, nil
}

type respStore struct {
	m  map[string]*wasmpb.Response
	mu sync.RWMutex
}

func (r *respStore) add(id string, resp *wasmpb.Response) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, found := r.m[id]
	if found {
		return fmt.Errorf("error storing response: response already exists for id: %s", id)
	}

	r.m[id] = resp
	return nil
}

func (r *respStore) get(id string) (*wasmpb.Response, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, found := r.m[id]
	if !found {
		return nil, fmt.Errorf("could not find response for id %s", id)
	}

	return r.m[id], nil
}

var (
	defaultTickInterval = 100 * time.Millisecond
	defaultTimeout      = 300 * time.Millisecond
	defaultMaxMemoryMBs = 64
	defaultInitialFuel  = uint64(100_000_000)
)

type ModuleConfig struct {
	TickInterval time.Duration
	Timeout      *time.Duration
	MaxMemoryMBs int64
	InitialFuel  uint64
	Logger       logger.Logger
}

type Module struct {
	engine *wasmtime.Engine
	module *wasmtime.Module
	linker *wasmtime.Linker

	r *respStore

	cfg *ModuleConfig

	wg     sync.WaitGroup
	stopCh chan struct{}
}

func NewModule(modCfg *ModuleConfig, binary []byte) (*Module, error) {
	if modCfg.Logger == nil {
		return nil, errors.New("must provide logger")
	}

	logger := modCfg.Logger

	if modCfg.TickInterval == 0 {
		modCfg.TickInterval = defaultTickInterval
	}

	if modCfg.Timeout == nil {
		modCfg.Timeout = &defaultTimeout
	}

	if modCfg.InitialFuel == 0 {
		modCfg.InitialFuel = defaultInitialFuel
	}

	// Take the max of the default and the configured max memory mbs.
	// We do this because Go requires a minimum of 16 megabytes to run,
	// and local testing has shown that with less than 64 mbs, some
	// binaries may error sporadically.
	modCfg.MaxMemoryMBs = int64(math.Max(float64(defaultMaxMemoryMBs), float64(modCfg.MaxMemoryMBs)))

	cfg := wasmtime.NewConfig()
	cfg.SetEpochInterruption(true)
	cfg.SetConsumeFuel(true)

	engine := wasmtime.NewEngineWithConfig(cfg)

	mod, err := wasmtime.NewModule(engine, binary)
	if err != nil {
		return nil, fmt.Errorf("error creating wasmtime module: %w", err)
	}

	linker := wasmtime.NewLinker(engine)
	linker.AllowShadowing(true)

	err = linker.DefineWasi()
	if err != nil {
		return nil, err
	}

	r := &respStore{
		m: map[string]*wasmpb.Response{},
	}

	// TODO: Stub out poll_oneoff correctly -- it's unclear what
	// the effect of this naive stub is as this syscall powers
	// notifications for time.Sleep, but will also effect other system notifications.
	// We need this stub to prevent binaries from calling time.Sleep and
	// starving our worker pool as a result.
	err = linker.FuncWrap(
		"wasi_snapshot_preview1",
		"poll_oneoff",
		func(caller *wasmtime.Caller, a int32, b int32, c int32, d int32) int32 {
			return 0
		},
	)
	if err != nil {
		return nil, fmt.Errorf("could not wrap poll_oneoff: %w", err)
	}

	err = linker.FuncWrap(
		"env",
		"sendResponse",
		func(caller *wasmtime.Caller, ptr int32, ptrlen int32) {
			b, innerErr := safeMem(caller, unsafe.Pointer(uintptr(ptr)), ptrlen)
			if innerErr != nil {
				logger.Errorf("error calling sendResponse: %s", err)
				return
			}

			var resp wasmpb.Response
			innerErr = proto.Unmarshal(b, &resp)
			if innerErr != nil {
				logger.Errorf("error calling sendResponse: %s", err)
				return
			}

			innerErr = r.add(resp.Id, &resp)
			if innerErr != nil {
				logger.Errorf("error calling sendResponse: %s", err)
				return
			}
		},
	)
	if err != nil {
		return nil, err
	}

	m := &Module{
		engine: engine,
		module: mod,
		linker: linker,

		r: r,

		cfg: modCfg,

		stopCh: make(chan struct{}),
	}

	return m, nil
}

func (m *Module) Start() {
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()

		ticker := time.NewTicker(m.cfg.TickInterval)
		for {
			select {
			case <-m.stopCh:
				return
			case <-ticker.C:
				m.engine.IncrementEpoch()
			}
		}
	}()
}

func (m *Module) Close() {
	close(m.stopCh)
	m.wg.Wait()

	m.linker.Close()
	m.engine.Close()
	m.module.Close()
}

func (m *Module) Run(request *wasmpb.Request) (*wasmpb.Response, error) {
	store := wasmtime.NewStore(m.engine)
	defer store.Close()

	reqpb, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	reqstr := base64.StdEncoding.EncodeToString(reqpb)

	wasi := wasmtime.NewWasiConfig()
	wasi.SetArgv([]string{"wasi", reqstr})

	store.SetWasi(wasi)
	err = store.SetFuel(m.cfg.InitialFuel)
	if err != nil {
		return nil, fmt.Errorf("error setting fuel: %w", err)
	}

	// Limit memory to max memory megabytes per instance.
	store.Limiter(
		m.cfg.MaxMemoryMBs*int64(math.Pow(10, 6)),
		-1, // tableElements, -1 == default
		1,  // instances
		1,  // tables
		1,  // memories
	)

	deadline := *m.cfg.Timeout / m.cfg.TickInterval
	store.SetEpochDeadline(uint64(deadline))

	instance, err := m.linker.Instantiate(store, m.module)
	if err != nil {
		return nil, err
	}

	start := instance.GetFunc(store, "_start")
	if start == nil {
		return nil, errors.New("could not get start function")
	}

	_, err = start.Call(store)
	switch {
	case containsCode(err, wasm.CodeSuccess):
		resp, innerErr := m.r.get(request.Id)
		if innerErr != nil {
			return nil, innerErr
		}
		return resp, nil
	case containsCode(err, wasm.CodeInvalidResponse):
		return nil, fmt.Errorf("invariant violation: error marshaling response")
	case containsCode(err, wasm.CodeInvalidRequest):
		return nil, fmt.Errorf("invariant violation: invalid request to runner")
	case containsCode(err, wasm.CodeRunnerErr):
		resp, innerErr := m.r.get(request.Id)
		if innerErr != nil {
			return nil, innerErr
		}

		return nil, fmt.Errorf("error executing runner: %s: %w", resp.ErrMsg, innerErr)
	default:
		return nil, err
	}
}

func containsCode(err error, code int) bool {
	return strings.Contains(err.Error(), fmt.Sprintf("exit status %d", code))
}
