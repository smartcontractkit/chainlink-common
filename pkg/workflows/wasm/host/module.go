package host

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"

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

type RequestData struct {
	response *wasmpb.Response
	ctx      context.Context
}

type store struct {
	m  map[string]*RequestData
	mu sync.RWMutex
}

func (r *store) add(id string, req *RequestData) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	storedReq, found := r.m[id]
	if found && req.response != nil {
		return fmt.Errorf("error storing response: response already exists for id: %s", id)
	}

	// we only add the response as the context has already been added
	if found && req.response == nil {
		r.m[id].response = storedReq.response
		return nil
	}

	r.m[id] = req
	return nil
}

func (r *store) get(id string) (*RequestData, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, found := r.m[id]
	if !found {
		return nil, fmt.Errorf("could not find request data for id %s", id)
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
	Fetch          func(ctx context.Context, req *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error)

	// If Determinism is set, the module will override the random_get function in the WASI API with
	// the provided seed to ensure deterministic behavior.
	Determinism *DeterminismConfig
}

type Module struct {
	engine *wasmtime.Engine
	module *wasmtime.Module
	linker *wasmtime.Linker

	requestStore *store

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

	if modCfg.Fetch == nil {
		modCfg.Fetch = func(context.Context, *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error) {
			return nil, fmt.Errorf("fetch not implemented")
		}
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

	linker, err := newWasiLinker(modCfg, engine)
	if err != nil {
		return nil, fmt.Errorf("error creating wasi linker: %w", err)
	}

	requestStore := &store{
		m: map[string]*RequestData{},
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
				logger.Errorf("error calling sendResponse: %s", innerErr)
				return ErrnoFault
			}

			storedReq, innerErr := requestStore.get(resp.Id)
			if innerErr != nil {
				logger.Errorf("error calling sendResponse: %s", innerErr)
				return ErrnoFault
			}
			storedReq.response = &resp

			return ErrnoSuccess
		},
	)
	if err != nil {
		return nil, fmt.Errorf("error wrapping sendResponse func: %w", err)
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
		return nil, fmt.Errorf("error wrapping log func: %w", err)
	}

	err = linker.FuncWrap(
		"env",
		"fetch",
		fetchFn(logger, modCfg, requestStore),
	)
	if err != nil {
		return nil, fmt.Errorf("error wrapping fetch func: %w", err)
	}

	m := &Module{
		engine: engine,
		module: mod,
		linker: linker,

		requestStore: requestStore,

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

func (m *Module) Run(ctx context.Context, request *wasmpb.Request) (*wasmpb.Response, error) {
	if request == nil {
		return nil, fmt.Errorf("invalid request: can't be nil")
	}

	if request.Id == "" {
		return nil, fmt.Errorf("invalid request: can't be empty")
	}

	// we add the request context to the store to make it available to the Fetch fn
	m.requestStore.add(request.Id, &RequestData{ctx: ctx})

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
		storedRequest, innerErr := m.requestStore.get(request.Id)
		if innerErr != nil {
			return nil, innerErr
		}

		if storedRequest.response == nil {
			return nil, fmt.Errorf("could not find response for id %s", request.Id)
		}

		return storedRequest.response, nil
	case containsCode(err, wasm.CodeInvalidResponse):
		return nil, fmt.Errorf("invariant violation: error marshaling response")
	case containsCode(err, wasm.CodeInvalidRequest):
		return nil, fmt.Errorf("invariant violation: invalid request to runner")
	case containsCode(err, wasm.CodeRunnerErr):
		storedRequest, innerErr := m.requestStore.get(request.Id)
		if innerErr != nil {
			return nil, innerErr
		}

		return nil, fmt.Errorf("error executing runner: %s: %w", storedRequest.response.ErrMsg, innerErr)
	case containsCode(err, wasm.CodeHostErr):
		return nil, fmt.Errorf("invariant violation: host errored during sendResponse")
	default:
		return nil, err
	}
}

func containsCode(err error, code int) bool {
	return strings.Contains(err.Error(), fmt.Sprintf("exit status %d", code))
}

func fetchFn(logger logger.Logger, modCfg *ModuleConfig, store *store) func(caller *wasmtime.Caller, respptr int32, resplenptr int32, reqptr int32, reqptrlen int32) int32 {
	logFetchErr := func(err error) { logger.Errorf("error calling fetch: %s", err.Error()) }
	return func(caller *wasmtime.Caller, respptr int32, resplenptr int32, reqptr int32, reqptrlen int32) int32 {
		b, innerErr := safeMem(caller, reqptr, reqptrlen)
		if innerErr != nil {
			logFetchErr(innerErr)
			return ErrnoFault
		}

		req := &wasmpb.FetchRequest{}
		innerErr = proto.Unmarshal(b, req)
		if innerErr != nil {
			logFetchErr(innerErr)
			return ErrnoFault
		}

		storedRequest, innerErr := store.get(req.Id)
		if innerErr != nil {
			logFetchErr(innerErr)
			return ErrnoFault
		}

		if storedRequest.ctx == nil {
			logFetchErr(errors.New("context is nil"))
			return ErrnoFault
		}

		fetchResp, innerErr := modCfg.Fetch(storedRequest.ctx, req)
		if innerErr != nil {
			logFetchErr(innerErr)
			return ErrnoFault
		}

		respBytes, innerErr := proto.Marshal(fetchResp)
		if innerErr != nil {
			logFetchErr(innerErr)
			return ErrnoFault
		}

		size := copyBuffer(caller, respBytes, respptr, int32(len(respBytes)))
		if size == -1 {
			return ErrnoFault
		}

		uint32Size := int32(4)
		resplenBytes := make([]byte, uint32Size)
		binary.LittleEndian.PutUint32(resplenBytes, uint32(len(respBytes)))
		size = copyBuffer(caller, resplenBytes, resplenptr, uint32Size)
		if size == -1 {
			return ErrnoFault
		}

		return ErrnoSuccess
	}
}
