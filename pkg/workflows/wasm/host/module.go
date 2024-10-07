package host

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/bytecodealliance/wasmtime-go/v23"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

// safeMem returns a copy of the wasm module memory at the given pointer and size.
func safeMem(caller *wasmtime.Caller, ptr int32, size int32) ([]byte, error) {
	mem := caller.GetExport("memory").Memory()
	data := mem.UnsafeData(caller)
	if ptr+size > int32(len(data)) {
		return nil, errors.New("out of bounds memory access")
	}

	cd := make([]byte, size)
	copy(cd, data[ptr:ptr+size])
	return cd, nil
}

// copyBuffer copies the given src byte slice into the wasm module memory at the given pointer and size.
func copyBuffer(caller *wasmtime.Caller, src []byte, ptr int32, size int32) int64 {
	mem := caller.GetExport("memory").Memory()
	rawData := mem.UnsafeData(caller)
	if int32(len(rawData)) < ptr+size {
		return -1
	}
	buffer := rawData[ptr : ptr+size]
	dataLen := int64(len(src))
	copy(buffer, src)
	return dataLen
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
	DefaultInitialFuel  = uint64(100_000_000)
)

type DeterminismConfig struct {
	// Seed is the seed used to generate cryptographically insecure random numbers in the module.
	Seed int64
}

type ModuleConfig struct {
	TickInterval   time.Duration
	Timeout        *time.Duration
	MaxMemoryMBs   int64
	InitialFuel    uint64
	Logger         logger.Logger
	IsUncompressed bool

	// If Determinism is set, the module will override the random_get function in the WASI API with
	// the provided seed to ensure deterministic behavior.
	Determinism *DeterminismConfig
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

// WithDeterminism sets the Determinism field to a deterministic seed from a known time.
//
// "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
func WithDeterminism() func(*ModuleConfig) {
	return func(cfg *ModuleConfig) {
		t, err := time.Parse(time.RFC3339Nano, "2009-01-03T00:00:00Z")
		if err != nil {
			panic(err)
		}

		cfg.Determinism = &DeterminismConfig{Seed: t.Unix()}
	}
}

func NewModule(modCfg *ModuleConfig, binary []byte, opts ...func(*ModuleConfig)) (*Module, error) {
	// Apply options to the module config.
	for _, opt := range opts {
		opt(modCfg)
	}

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

	// Take the max of the default and the configured max memory mbs.
	// We do this because Go requires a minimum of 16 megabytes to run,
	// and local testing has shown that with less than 64 mbs, some
	// binaries may error sporadically.
	modCfg.MaxMemoryMBs = int64(math.Max(float64(defaultMaxMemoryMBs), float64(modCfg.MaxMemoryMBs)))

	cfg := wasmtime.NewConfig()
	cfg.SetEpochInterruption(true)
	if modCfg.InitialFuel > 0 {
		cfg.SetConsumeFuel(true)
	}

	engine := wasmtime.NewEngineWithConfig(cfg)

	if !modCfg.IsUncompressed {
		rdr := brotli.NewReader(bytes.NewBuffer(binary))
		decompedBinary, err := io.ReadAll(rdr)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress binary: %w", err)
		}

		binary = decompedBinary
	}

	mod, err := wasmtime.NewModule(engine, binary)
	if err != nil {
		return nil, fmt.Errorf("error creating wasmtime module: %w", err)
	}

	linker, err := newWasiLinker(engine)
	if err != nil {
		return nil, fmt.Errorf("error creating wasi linker: %w", err)
	}

	r := &respStore{
		m: map[string]*wasmpb.Response{},
	}

	err = linker.FuncWrap(
		"env",
		"sendResponse",
		func(caller *wasmtime.Caller, ptr int32, ptrlen int32) int32 {
			b, innerErr := safeMem(caller, ptr, ptrlen)
			if innerErr != nil {
				logger.Errorf("error calling sendResponse: %s", err)
				return ErrnoFault
			}

			var resp wasmpb.Response
			innerErr = proto.Unmarshal(b, &resp)
			if innerErr != nil {
				logger.Errorf("error calling sendResponse: %s", err)
				return ErrnoFault
			}

			innerErr = r.add(resp.Id, &resp)
			if innerErr != nil {
				logger.Errorf("error calling sendResponse: %s", err)
				return ErrnoFault
			}

			return ErrnoSuccess
		},
	)
	if err != nil {
		return nil, err
	}

	err = linker.FuncWrap(
		"env",
		"log",
		func(caller *wasmtime.Caller, ptr int32, ptrlen int32) {
			b, innerErr := safeMem(caller, ptr, ptrlen)
			if innerErr != nil {
				logger.Errorf("error calling log: %s", err)
				return
			}

			var raw map[string]interface{}
			innerErr = json.Unmarshal(b, &raw)
			if innerErr != nil {
				return
			}

			level := raw["level"]
			delete(raw, "level")

			msg := raw["msg"].(string)
			delete(raw, "msg")
			delete(raw, "ts")

			var args []interface{}
			for k, v := range raw {
				args = append(args, k, v)
			}

			switch level {
			case "debug":
				logger.Debugw(msg, args...)
			case "info":
				logger.Infow(msg, args...)
			case "warn":
				logger.Warnw(msg, args...)
			case "error":
				logger.Errorw(msg, args...)
			case "panic":
				logger.Panicw(msg, args...)
			case "fatal":
				logger.Fatalw(msg, args...)
			default:
				logger.Infow(msg, args...)
			}
		},
	)
	if err != nil {
		return nil, err
	}

	if modCfg.Determinism != nil {
		err = linker.FuncWrap(
			"wasi_snapshot_preview1",
			"random_get",
			createRandomGet(modCfg),
		)
		if err != nil {
			return nil, err
		}
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

// randomGet overrides the random_get function in the WASI API and fixes the random source with
// a hardcoded seed via insecure randomness.  This is to ensure that workflows are deterministic.
// Randomness should not be used in workflows.
func createRandomGet(cfg *ModuleConfig) func(caller *wasmtime.Caller, buf, bufLen int32) int32 {
	return func(caller *wasmtime.Caller, buf, bufLen int32) int32 {
		if cfg.Determinism == nil {
			return ErrnoInval
		}

		var (
			// Fix the random source with a hardcoded seed
			seed       = cfg.Determinism.Seed
			randSource = rand.New(rand.NewSource(seed)) //nolint:gosec
			randOutput = make([]byte, bufLen)
		)

		// Generate random bytes from the source
		if _, err := io.ReadAtLeast(randSource, randOutput, int(bufLen)); err != nil {
			return ErrnoFault
		}

		// Copy the random bytes into the wasm module memory
		if n := copyBuffer(caller, randOutput, buf, bufLen); n != int64(len(randOutput)) {
			return ErrnoFault
		}

		return ErrnoSuccess
	}
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

	if m.cfg.InitialFuel > 0 {
		err = store.SetFuel(m.cfg.InitialFuel)
		if err != nil {
			return nil, fmt.Errorf("error setting fuel: %w", err)
		}
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
	case containsCode(err, wasm.CodeHostErr):
		return nil, fmt.Errorf("invariant violation: host errored during sendResponse")
	default:
		return nil, err
	}
}

func containsCode(err error, code int) bool {
	return strings.Contains(err.Error(), fmt.Sprintf("exit status %d", code))
}
