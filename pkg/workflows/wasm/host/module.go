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
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

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

type MessageEmitter interface {
	// Emit sends a message to the labeler's destination.
	Emit(string) error

	// WithMapLabels sets the labels for the message to be emitted.  Labels are cumulative.
	WithMapLabels(map[string]string) MessageEmitter
}

type ModuleConfig struct {
	TickInterval   time.Duration
	Timeout        *time.Duration
	MaxMemoryMBs   int64
	InitialFuel    uint64
	Logger         logger.Logger
	IsUncompressed bool
	Fetch          func(ctx context.Context, req *wasmpb.FetchRequest) (*wasmpb.FetchResponse, error)

	// Labeler is used to emit messages from the module.
	Labeler MessageEmitter

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

	if modCfg.Labeler == nil {
		modCfg.Labeler = &unimplementedMessageEmitter{}
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
			b, innerErr := wasmRead(caller, ptr, ptrlen)
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
			b, innerErr := wasmRead(caller, ptr, ptrlen)
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

	err = linker.FuncWrap(
		"env",
		"emit",
		createEmitFn(logger, modCfg.Labeler, wasmRead, wasmWrite, wasmWriteUInt32),
	)
	if err != nil {
		return nil, fmt.Errorf("error wrapping emit func: %w", err)
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
		b, innerErr := wasmRead(caller, reqptr, reqptrlen)
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

		if size := wasmWrite(caller, respBytes, respptr, int32(len(respBytes))); size == -1 {
			return ErrnoFault
		}

		if size := wasmWriteUInt32(caller, resplenptr, uint32(len(respBytes))); size == -1 {
			return ErrnoFault
		}

		return ErrnoSuccess
	}
}

// createEmitFn injects dependencies and builds the emit function exposed by the WASM.  Errors in
// Emit, if any, are returned in the Error Message of the response.
func createEmitFn(
	l logger.Logger,
	e MessageEmitter,
	reader unsafeReaderFunc,
	writer unsafeWriterFunc,
	sizeWriter unsafeFixedLengthWriterFunc,
) func(caller *wasmtime.Caller, respptr, resplenptr, msgptr, msglen int32) int32 {
	logErr := func(err error) {
		l.Errorf("error emitting message: %s", err)
	}

	return func(caller *wasmtime.Caller, respptr, resplenptr, msgptr, msglen int32) int32 {
		// writeErr marshals and writes an error response to wasm
		writeErr := func(err error) int32 {
			logErr(err)

			resp := &wasmpb.EmitMessageResponse{
				Error: &wasmpb.Error{
					Message: err.Error(),
				},
			}

			respBytes, perr := proto.Marshal(resp)
			if perr != nil {
				logErr(perr)
				return ErrnoFault
			}

			if size := writer(caller, respBytes, respptr, int32(len(respBytes))); size == -1 {
				logErr(errors.New("failed to write response"))
				return ErrnoFault
			}

			if size := sizeWriter(caller, resplenptr, uint32(len(respBytes))); size == -1 {
				logErr(errors.New("failed to write response length"))
				return ErrnoFault
			}

			return ErrnoSuccess
		}

		b, err := reader(caller, msgptr, msglen)
		if err != nil {
			return writeErr(err)
		}

		msg, labels, err := toEmissible(b)
		if err != nil {
			return writeErr(err)
		}

		if err := e.WithMapLabels(labels).Emit(msg); err != nil {
			return writeErr(err)
		}

		return ErrnoSuccess
	}
}

type unimplementedMessageEmitter struct{}

func (u *unimplementedMessageEmitter) Emit(string) error {
	return errors.New("unimplemented")
}

func (u *unimplementedMessageEmitter) WithMapLabels(map[string]string) MessageEmitter {
	return u
}

func toEmissible(b []byte) (string, map[string]string, error) {
	msg := &wasmpb.EmitMessageRequest{}
	if err := proto.Unmarshal(b, msg); err != nil {
		return "", nil, err
	}

	validated, err := toValidatedLabels(msg)
	if err != nil {
		return "", nil, err
	}

	return msg.Message, validated, nil
}

func toValidatedLabels(msg *wasmpb.EmitMessageRequest) (map[string]string, error) {
	vl, err := values.FromMapValueProto(msg.Labels)
	if err != nil {
		return nil, err
	}

	// Handle the case of no labels before unwrapping.
	if vl == nil {
		vl = values.EmptyMap()
	}

	var labels map[string]string
	if err := vl.UnwrapTo(&labels); err != nil {
		return nil, err
	}

	return labels, nil
}

// unsafeWriterFunc defines behavior for writing directly to wasm memory.  A source slice of bytes
// is written to the location defined by the ptr.
type unsafeWriterFunc func(c *wasmtime.Caller, src []byte, ptr, len int32) int64

// unsafeFixedLengthWriterFunc defines behavior for writing a uint32 value to wasm memory at the location defined
// by the ptr.
type unsafeFixedLengthWriterFunc func(c *wasmtime.Caller, ptr int32, val uint32) int64

// unsafeReaderFunc abstractly defines the behavior of reading from WASM memory.  Returns a copy of
// the memory at the given pointer and size.
type unsafeReaderFunc func(c *wasmtime.Caller, ptr, len int32) ([]byte, error)

// wasmMemoryAccessor is the default implementation for unsafely accessing the memory of the WASM module.
func wasmMemoryAccessor(caller *wasmtime.Caller) []byte {
	return caller.GetExport("memory").Memory().UnsafeData(caller)
}

// wasmRead returns a copy of the wasm module memory at the given pointer and size.
func wasmRead(caller *wasmtime.Caller, ptr int32, size int32) ([]byte, error) {
	return read(wasmMemoryAccessor(caller), ptr, size)
}

// Read acts on a byte slice that should represent an unsafely accessed slice of memory.  It returns
// a copy of the memory at the given pointer and size.
func read(memory []byte, ptr int32, size int32) ([]byte, error) {
	if size < 0 || ptr < 0 {
		return nil, fmt.Errorf("invalid memory access: ptr: %d, size: %d", ptr, size)
	}

	if ptr+size > int32(len(memory)) {
		return nil, errors.New("out of bounds memory access")
	}

	cd := make([]byte, size)
	copy(cd, memory[ptr:ptr+size])
	return cd, nil
}

// wasmWrite copies the given src byte slice into the wasm module memory at the given pointer and size.
func wasmWrite(caller *wasmtime.Caller, src []byte, ptr int32, size int32) int64 {
	return write(wasmMemoryAccessor(caller), src, ptr, size)
}

// wasmWriteUInt32 binary encodes and writes a uint32 to the wasm module memory at the given pointer.
func wasmWriteUInt32(caller *wasmtime.Caller, ptr int32, val uint32) int64 {
	return writeUInt32(wasmMemoryAccessor(caller), ptr, val)
}

// writeUInt32 binary encodes and writes a uint32 to the memory at the given pointer.
func writeUInt32(memory []byte, ptr int32, val uint32) int64 {
	uint32Size := int32(4)
	buffer := make([]byte, uint32Size)
	binary.LittleEndian.PutUint32(buffer, val)
	return write(memory, buffer, ptr, uint32Size)
}

// write copies the given src byte slice into the memory at the given pointer and size.
func write(memory, src []byte, ptr, size int32) int64 {
	if size < 0 || ptr < 0 {
		return -1
	}

	if int32(len(memory)) < ptr+size {
		return -1
	}
	buffer := memory[ptr : ptr+size]
	dataLen := int64(len(src))
	copy(buffer, src)
	return dataLen
}
