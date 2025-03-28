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
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/bytecodealliance/wasmtime-go/v28"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/custmsg"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	wasm "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/legacy"
	wasmdagpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/legacy/pb"
	wasmpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

type RequestData[T proto.Message] struct {
	fetchRequestsCounter int
	response             T
	ctx                  func() context.Context
}

type store[T proto.Message] struct {
	m  map[string]*RequestData[T]
	mu sync.RWMutex
}

func (r *store[T]) add(id string, req *RequestData[T]) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, found := r.m[id]
	if found {
		return fmt.Errorf("error storing response: response already exists for id: %s", id)
	}

	r.m[id] = req
	return nil
}

func (r *store[T]) get(id string) (*RequestData[T], error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, found := r.m[id]
	if !found {
		return nil, fmt.Errorf("could not find request data for id %s", id)
	}

	return r.m[id], nil
}

func (r *store[T]) delete(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.m, id)
}

type triggerStore struct {
	triggers    []*wasmpb.TriggerSubscriptionRequest
	triggerLock sync.Mutex
}

func (t *triggerStore) add(trigger *wasmpb.TriggerSubscriptionRequest) {
	t.triggerLock.Lock()
	defer t.triggerLock.Unlock()
	t.triggers = append(t.triggers, trigger)
}

func (t *triggerStore) get() []*wasmpb.TriggerSubscriptionRequest {
	t.triggerLock.Lock()
	defer t.triggerLock.Unlock()
	return t.triggers
}

func (t *triggerStore) subscribeToTrigger(caller *wasmtime.Caller, subscription, subscriptionLen int32) int32 {
	triggerBytes, err := wasmRead(caller, subscription, subscriptionLen)
	if err != nil {
		return ErrnoFault
	}

	trigger := &wasmpb.TriggerSubscriptionRequest{}
	if err = proto.Unmarshal(triggerBytes, trigger); err != nil {
		return ErrnoFault
	}

	t.add(trigger)
	return ErrnoSuccess
}

var (
	defaultTickInterval              = 100 * time.Millisecond
	defaultTimeout                   = 10 * time.Second
	defaultMinMemoryMBs              = uint64(128)
	DefaultInitialFuel               = uint64(100_000_000)
	defaultMaxFetchRequests          = 5
	defaultMaxCompressedBinarySize   = 20 * 1024 * 1024  // 20 MB
	defaultMaxDecompressedBinarySize = 100 * 1024 * 1024 // 100 MB
)

type DeterminismConfig struct {
	// Seed is the seed used to generate cryptographically insecure random numbers in the module.
	Seed int64
}
type ModuleConfig struct {
	TickInterval              time.Duration
	Timeout                   *time.Duration
	MaxMemoryMBs              uint64
	MinMemoryMBs              uint64
	InitialFuel               uint64
	Logger                    logger.Logger
	IsUncompressed            bool
	Fetch                     func(ctx context.Context, req *wasmdagpb.FetchRequest) (*wasmdagpb.FetchResponse, error)
	MaxFetchRequests          int
	MaxCompressedBinarySize   uint64
	MaxDecompressedBinarySize uint64

	// Labeler is used to emit messages from the module.
	Labeler custmsg.MessageEmitter

	// If Determinism is set, the module will override the random_get function in the WASI API with
	// the provided seed to ensure deterministic behavior.
	Determinism *DeterminismConfig

	CallCapability    func(ctx context.Context, req *wasmpb.CapabilityRequest) ([sdk.IdLen]byte, error)
	AwaitCapabilities func(ctx context.Context, req *wasmpb.AwaitCapabilitiesRequest) (*wasmpb.AwaitCapabilitiesResponse, error)
}

type Module struct {
	engine  *wasmtime.Engine
	module  *wasmtime.Module
	linker  *wasmtime.Linker
	wconfig *wasmtime.Config

	requestStore *store[*wasmdagpb.Response]
	executeStore *store[*wasmpb.ExecutionResult]
	triggers     *triggerStore

	cfg *ModuleConfig

	wg     sync.WaitGroup
	stopCh chan struct{}

	isLegacyDAG bool
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
		modCfg.Fetch = func(context.Context, *wasmdagpb.FetchRequest) (*wasmdagpb.FetchResponse, error) {
			return nil, fmt.Errorf("fetch not implemented")
		}
	}

	if modCfg.MaxFetchRequests == 0 {
		modCfg.MaxFetchRequests = defaultMaxFetchRequests
	}

	if modCfg.Labeler == nil {
		modCfg.Labeler = &unimplementedMessageEmitter{}
	}

	if modCfg.TickInterval == 0 {
		modCfg.TickInterval = defaultTickInterval
	}

	if modCfg.Timeout == nil {
		modCfg.Timeout = &defaultTimeout
	}

	if modCfg.MinMemoryMBs == 0 {
		modCfg.MinMemoryMBs = defaultMinMemoryMBs
	}

	if modCfg.MaxCompressedBinarySize == 0 {
		modCfg.MaxCompressedBinarySize = uint64(defaultMaxCompressedBinarySize)
	}

	if modCfg.MaxDecompressedBinarySize == 0 {
		modCfg.MaxDecompressedBinarySize = uint64(defaultMaxDecompressedBinarySize)
	}

	// Take the max of the min and the configured max memory mbs.
	// We do this because Go requires a minimum of 16 megabytes to run,
	// and local testing has shown that with less than the min, some
	// binaries may error sporadically.
	modCfg.MaxMemoryMBs = uint64(math.Max(float64(modCfg.MinMemoryMBs), float64(modCfg.MaxMemoryMBs)))

	cfg := wasmtime.NewConfig()
	cfg.SetEpochInterruption(true)
	if modCfg.InitialFuel > 0 {
		cfg.SetConsumeFuel(true)
	}

	cfg.CacheConfigLoadDefault()
	cfg.SetCraneliftOptLevel(wasmtime.OptLevelSpeedAndSize)

	// Load testing shows that leaving native unwind info enabled causes a very large slowdown when loading multiple modules.
	cfg.SetNativeUnwindInfo(false)

	engine := wasmtime.NewEngineWithConfig(cfg)
	if !modCfg.IsUncompressed {
		// validate the binary size before decompressing
		// this is to prevent decompression bombs
		if uint64(len(binary)) > modCfg.MaxCompressedBinarySize {
			return nil, fmt.Errorf("compressed binary size exceeds the maximum allowed size of %d bytes", modCfg.MaxCompressedBinarySize)
		}

		rdr := io.LimitReader(brotli.NewReader(bytes.NewBuffer(binary)), int64(modCfg.MaxDecompressedBinarySize+1))
		decompedBinary, err := io.ReadAll(rdr)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress binary: %w", err)
		}

		binary = decompedBinary
	}

	// Validate the decompressed binary size.
	// io.LimitReader prevents decompression bombs by reading up to a set limit, but it will not return an error if the limit is reached.
	// The Read() method will return io.EOF, and ReadAll will gracefully handle it and return nil.
	if uint64(len(binary)) > modCfg.MaxDecompressedBinarySize {
		return nil, fmt.Errorf("decompressed binary size reached the maximum allowed size of %d bytes", modCfg.MaxDecompressedBinarySize)
	}

	mod, err := wasmtime.NewModule(engine, binary)
	if err != nil {
		return nil, fmt.Errorf("error creating wasmtime module: %w", err)
	}

	linker, err := newWasiLinker(modCfg, engine)
	if err != nil {
		return nil, fmt.Errorf("error creating wasi linker: %w", err)
	}

	requestStore := &store[*wasmdagpb.Response]{
		m: map[string]*RequestData[*wasmdagpb.Response]{},
	}

	execStore := &store[*wasmpb.ExecutionResult]{
		m: map[string]*RequestData[*wasmpb.ExecutionResult]{},
	}

	isLegacyDAG := false
	for _, modImport := range mod.Imports() {
		name := modImport.Name()
		if modImport.Module() == "env" && name != nil && *name == "fetch" {
			isLegacyDAG = true
			break
		}
	}

	triggers := &triggerStore{}

	if isLegacyDAG {
		if err = linkLegacyDAG(linker, modCfg, requestStore); err != nil {
			return nil, err
		}
	} else if err = linkNoDAG(linker, modCfg, execStore, triggers); err != nil {
		return nil, err
	}

	m := &Module{
		engine:  engine,
		module:  mod,
		linker:  linker,
		wconfig: cfg,

		requestStore: requestStore,
		executeStore: execStore,
		triggers:     triggers,

		cfg: modCfg,

		stopCh:      make(chan struct{}),
		isLegacyDAG: isLegacyDAG,
	}

	return m, nil
}

func linkNoDAG(linker *wasmtime.Linker, modCfg *ModuleConfig, execStore *store[*wasmpb.ExecutionResult], triggerStore *triggerStore) error {
	logger := modCfg.Logger
	if err := linker.FuncWrap(
		"env",
		"send_response",
		createSendResponseFn(logger, execStore, func() *wasmpb.ExecutionResult {
			return &wasmpb.ExecutionResult{}
		}),
	); err != nil {
		return fmt.Errorf("error wrapping sendResponse func: %w", err)
	}

	if err := linker.FuncWrap(
		"env",
		"subscribe_to_trigger",
		triggerStore.subscribeToTrigger,
	); err != nil {
		return err
	}

	if err := linker.FuncWrap(
		"env",
		"call_capability",
		createCallCapFn(logger, execStore, modCfg.CallCapability),
	); err != nil {
		return fmt.Errorf("error wrapping callcap func: %w", err)
	}

	if err := linker.FuncWrap(
		"env",
		"await_capabilities",
		createAwaitCapsFn(logger, execStore, modCfg.AwaitCapabilities),
	); err != nil {
		return fmt.Errorf("error wrapping awaitcaps func: %w", err)
	}

	return nil
}

func linkLegacyDAG(linker *wasmtime.Linker, modCfg *ModuleConfig, requestStore *store[*wasmdagpb.Response]) error {
	logger := modCfg.Logger
	err := linker.FuncWrap(
		"env",
		"sendResponse",
		createSendResponseFn(logger, requestStore, func() *wasmdagpb.Response {
			return &wasmdagpb.Response{}
		}),
	)
	if err != nil {
		return fmt.Errorf("error wrapping sendResponse func: %w", err)
	}

	err = linker.FuncWrap(
		"env",
		"log",
		createLogFn(logger),
	)
	if err != nil {
		return fmt.Errorf("error wrapping log func: %w", err)
	}

	err = linker.FuncWrap(
		"env",
		"fetch",
		createFetchFn(logger, wasmRead, wasmWrite, wasmWriteUInt32, modCfg, requestStore),
	)
	if err != nil {
		return fmt.Errorf("error wrapping fetch func: %w", err)
	}

	err = linker.FuncWrap(
		"env",
		"emit",
		createEmitFn(logger, requestStore, modCfg.Labeler, wasmRead, wasmWrite, wasmWriteUInt32),
	)
	if err != nil {
		return fmt.Errorf("error wrapping emit func: %w", err)
	}

	return nil
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
	m.wconfig.Close()
}

func (m *Module) IsLegacyDag() bool {
	return m.isLegacyDAG
}

func (m *Module) Execute(ctx context.Context, req *wasmpb.ExecuteRequest) (*wasmpb.ExecutionResult, error) {
	if m.isLegacyDAG {
		return nil, errors.New("cannot execute a legacy dag workflow")
	}

	if req == nil {
		return nil, fmt.Errorf("invalid request: can't be nil")
	}

	return runWasm[*wasmpb.ExecuteRequest, *wasmpb.ExecutionResult](m, ctx, req, m.executeStore)
}

// Run is deprecated, use execute instead
func (m *Module) Run(ctx context.Context, request *wasmdagpb.Request) (*wasmdagpb.Response, error) {
	if request == nil {
		return nil, fmt.Errorf("invalid request: can't be nil")
	}

	if request.Id == "" {
		return nil, fmt.Errorf("invalid request: can't be empty")
	}

	if !m.isLegacyDAG {
		return m.wrapRunForNoDAG(ctx, request)
	}

	return runWasm[*wasmdagpb.Request, *wasmdagpb.Response](m, ctx, request, m.requestStore)
}

type idMessage interface {
	GetId() string
	proto.Message
}

func runWasm[I idMessage, O proto.Message](
	m *Module, ctx context.Context, request I, execStore *store[O]) (O, error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, *m.cfg.Timeout)
	defer cancel()

	var o O

	// we add the request context to the store to make it available to the Fetch fn
	reqId := request.GetId()
	ctxCallback := func() context.Context { return ctxWithTimeout }
	err := execStore.add(reqId, &RequestData[O]{ctx: ctxCallback})
	if err != nil {
		return o, fmt.Errorf("error adding ctx to the store: %w", err)
	}

	// we delete the request data from the store when we're done
	defer m.requestStore.delete(reqId)

	store := wasmtime.NewStore(m.engine)

	defer store.Close()

	reqpb, err := proto.Marshal(request)
	if err != nil {
		return o, err
	}

	reqstr := base64.StdEncoding.EncodeToString(reqpb)

	wasi := wasmtime.NewWasiConfig()
	wasi.InheritStdout()
	defer wasi.Close()

	wasi.SetArgv([]string{"wasi", reqstr})

	store.SetWasi(wasi)

	if m.cfg.InitialFuel > 0 {
		err = store.SetFuel(m.cfg.InitialFuel)
		if err != nil {
			return o, fmt.Errorf("error setting fuel: %w", err)
		}
	}

	// Limit memory to max memory megabytes per instance.
	store.Limiter(
		int64(m.cfg.MaxMemoryMBs)*int64(math.Pow(10, 6)),
		-1, // tableElements, -1 == default
		1,  // instances
		1,  // tables
		1,  // memories
	)

	deadline := *m.cfg.Timeout / m.cfg.TickInterval
	store.SetEpochDeadline(uint64(deadline))

	instance, err := m.linker.Instantiate(store, m.module)
	if err != nil {
		return o, err
	}

	start := instance.GetFunc(store, "_start")
	if start == nil {
		return o, errors.New("could not get start function")
	}

	_, err = start.Call(store)
	switch {
	case containsCode(err, wasm.CodeSuccess):
		if err = m.wrapTriggersToResponse(reqId); err != nil {
			return o, err
		}

		storedRequest, innerErr := execStore.get(reqId)
		if innerErr != nil {
			return o, innerErr
		}

		if any(storedRequest.response) == nil {
			return o, fmt.Errorf("could not find response for id %s", reqId)
		}

		return storedRequest.response, nil
	case containsCode(err, wasm.CodeInvalidResponse):
		return o, fmt.Errorf("invariant violation: error marshaling response")
	case containsCode(err, wasm.CodeInvalidRequest):
		return o, fmt.Errorf("invariant violation: invalid request to runner")
	case containsCode(err, wasm.CodeRunnerErr):
		storedRequest, innerErr := m.requestStore.get(reqId)
		if innerErr != nil {
			return o, innerErr
		}

		return o, fmt.Errorf("error executing runner: %s: %w", storedRequest.response.ErrMsg, err)
	case containsCode(err, wasm.CodeHostErr):
		return o, fmt.Errorf("invariant violation: host errored during sendResponse")
	default:
		return o, err
	}
}

func (m *Module) wrapRunForNoDAG(ctx context.Context, request *wasmdagpb.Request) (*wasmdagpb.Response, error) {
	execRequest := &wasmpb.ExecuteRequest{Id: request.Id, Config: request.Config}
	switch msg := request.Message.(type) {
	case *wasmdagpb.Request_ComputeRequest:
		execRequest.Request = &wasmpb.ExecuteRequest_Trigger{
			Trigger: &wasmpb.Trigger{
				Id:      request.TriggerId,
				Payload: msg.ComputeRequest.Request.Request,
			},
		}
	case *wasmdagpb.Request_SpecRequest:
		execRequest.Request = &wasmpb.ExecuteRequest_Subscribe{Subscribe: &emptypb.Empty{}}
	default:
		return nil, fmt.Errorf("invalid request message type: %T", msg)
	}

	response, err := m.Execute(ctx, execRequest)

	if err != nil {
		return nil, err
	}

	respErr := ""
	var value *pb.Value
	switch r := response.Result.(type) {
	case *wasmpb.ExecutionResult_Error:
		respErr = r.Error
	case *wasmpb.ExecutionResult_Value:
		value = r.Value
	}

	output := &pb.Map{Fields: map[string]*pb.Value{"Output": value}}

	return &wasmdagpb.Response{
		Id:     response.Id,
		ErrMsg: respErr,
		Message: &wasmdagpb.Response_ComputeResponse{
			ComputeResponse: &wasmdagpb.ComputeResponse{
				Response: &capabilitiespb.CapabilityResponse{Value: output},
			},
		},
	}, nil
}

func (m *Module) wrapTriggersToResponse(reqId string) error {
	triggers := m.triggers.get()
	if len(triggers) == 0 {
		return nil
	}
	result, err := values.Wrap(triggers)
	if err != nil {
		return err
	}
	execution, err := m.executeStore.get(reqId)
	if err != nil {
		return err
	}

	resultPb := values.Proto(result)
	execution.response = &wasmpb.ExecutionResult{Id: reqId, Result: &wasmpb.ExecutionResult_Value{Value: resultPb}}

	return nil
}

func containsCode(err error, code int) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), fmt.Sprintf("exit status %d", code))
}

// createSendResponseFn injects the dependency required by a WASM guest to
// send a response back to the host.
func createSendResponseFn[T idMessage](
	logger logger.Logger,
	requestStore *store[T],
	newT func() T) func(caller *wasmtime.Caller, ptr int32, ptrlen int32) int32 {
	return func(caller *wasmtime.Caller, ptr int32, ptrlen int32) int32 {
		b, innerErr := wasmRead(caller, ptr, ptrlen)
		if innerErr != nil {
			logger.Errorf("error calling sendResponse: %s", innerErr)
			return ErrnoFault
		}

		resp := newT()
		innerErr = proto.Unmarshal(b, resp)
		if innerErr != nil {
			logger.Errorf("error calling sendResponse: %s", innerErr)
			return ErrnoFault
		}

		storedReq, innerErr := requestStore.get(resp.GetId())
		if innerErr != nil {
			logger.Errorf("error calling sendResponse: %s", innerErr)
			return ErrnoFault
		}
		storedReq.response = resp

		return ErrnoSuccess
	}
}

func createFetchFn(
	logger logger.Logger,
	reader unsafeReaderFunc,
	writer unsafeWriterFunc,
	sizeWriter unsafeFixedLengthWriterFunc,
	modCfg *ModuleConfig,
	requestStore *store[*wasmdagpb.Response],
) func(caller *wasmtime.Caller, respptr int32, resplenptr int32, reqptr int32, reqptrlen int32) int32 {
	return func(caller *wasmtime.Caller, respptr int32, resplenptr int32, reqptr int32, reqptrlen int32) int32 {
		const errFetchSfx = "error calling fetch"

		// writeErr marshals and writes an error response to wasm
		writeErr := func(err error) int32 {
			resp := &wasmdagpb.FetchResponse{
				ExecutionError: true,
				ErrorMessage:   err.Error(),
			}

			respBytes, perr := proto.Marshal(resp)
			if perr != nil {
				logger.Errorf("%s: %s", errFetchSfx, perr)
				return ErrnoFault
			}

			if size := writer(caller, respBytes, respptr, int32(len(respBytes))); size == -1 {
				logger.Errorf("%s: %s", errFetchSfx, errors.New("failed to write error response"))
				return ErrnoFault
			}

			if size := sizeWriter(caller, resplenptr, uint32(len(respBytes))); size == -1 {
				logger.Errorf("%s: %s", errFetchSfx, errors.New("failed to write error response length"))
				return ErrnoFault
			}

			return ErrnoSuccess
		}

		b, innerErr := reader(caller, reqptr, reqptrlen)
		if innerErr != nil {
			logger.Errorf("%s: %s", errFetchSfx, innerErr)
			return writeErr(innerErr)
		}

		req := &wasmdagpb.FetchRequest{}
		innerErr = proto.Unmarshal(b, req)
		if innerErr != nil {
			logger.Errorf("%s: %s", errFetchSfx, innerErr)
			return writeErr(innerErr)
		}

		storedRequest, innerErr := requestStore.get(req.Id)
		if innerErr != nil {
			logger.Errorf("%s: %s", errFetchSfx, innerErr)
			return writeErr(innerErr)
		}

		// limit the number of fetch calls we can make per request
		if storedRequest.fetchRequestsCounter >= modCfg.MaxFetchRequests {
			logger.Errorf("%s: max number of fetch request %d exceeded", errFetchSfx, modCfg.MaxFetchRequests)
			return writeErr(errors.New("max number of fetch requests exceeded"))
		}
		storedRequest.fetchRequestsCounter++

		fetchResp, innerErr := modCfg.Fetch(storedRequest.ctx(), req)
		if innerErr != nil {
			logger.Errorf("%s: %s", errFetchSfx, innerErr)
			return writeErr(innerErr)
		}

		respBytes, innerErr := proto.Marshal(fetchResp)
		if innerErr != nil {
			logger.Errorf("%s: %s", errFetchSfx, innerErr)
			return writeErr(innerErr)
		}

		if size := writer(caller, respBytes, respptr, int32(len(respBytes))); size == -1 {
			return writeErr(errors.New("failed to write response"))
		}

		if size := sizeWriter(caller, resplenptr, uint32(len(respBytes))); size == -1 {
			return writeErr(errors.New("failed to write response length"))
		}

		return ErrnoSuccess
	}
}

// createEmitFn injects dependencies and builds the emit function exposed by the WASM.  Errors in
// Emit, if any, are returned in the Error Message of the response.
func createEmitFn(
	l logger.Logger,
	requestStore *store[*wasmdagpb.Response],
	e custmsg.MessageEmitter,
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

			resp := &wasmdagpb.EmitMessageResponse{
				Error: &wasmdagpb.Error{
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

		reqID, msg, labels, err := toEmissible(b)
		if err != nil {
			return writeErr(err)
		}

		req, err := requestStore.get(reqID)
		if err != nil {
			logErr(fmt.Errorf("failed to get request from store: %s", err))
			return writeErr(err)
		}

		if err := e.WithMapLabels(labels).Emit(req.ctx(), msg); err != nil {
			return writeErr(err)
		}

		return ErrnoSuccess
	}
}

// createLogFn injects dependencies and builds the log function exposed by the WASM.
func createLogFn(logger logger.Logger) func(caller *wasmtime.Caller, ptr int32, ptrlen int32) {
	return func(caller *wasmtime.Caller, ptr int32, ptrlen int32) {
		b, innerErr := wasmRead(caller, ptr, ptrlen)
		if innerErr != nil {
			logger.Errorf("error calling log: %s", innerErr)
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

		reg, _ := regexp.Compile(`[\r\n\t]|[\x00-\x1F]|[<>\"'\\&%$;:{}\[\]/]`)
		sanitizedMsg := reg.ReplaceAllString(msg, "*")

		switch level {
		case "debug":
			logger.Debugw(sanitizedMsg, args...)
		case "info":
			logger.Infow(sanitizedMsg, args...)
		case "warn":
			logger.Warnw(sanitizedMsg, args...)
		case "error":
			logger.Errorw(sanitizedMsg, args...)
		case "panic":
			logger.Panicw(sanitizedMsg, args...)
		case "fatal":
			logger.Fatalw(sanitizedMsg, args...)
		default:
			logger.Infow(sanitizedMsg, args...)
		}
	}
}

type unimplementedMessageEmitter struct{}

func (u *unimplementedMessageEmitter) Emit(context.Context, string) error {
	return errors.New("unimplemented")
}

func (u *unimplementedMessageEmitter) WithMapLabels(map[string]string) custmsg.MessageEmitter {
	return u
}

func (u *unimplementedMessageEmitter) With(kvs ...string) custmsg.MessageEmitter {
	return u
}

func (u *unimplementedMessageEmitter) Labels() map[string]string {
	return nil
}

func toEmissible(b []byte) (string, string, map[string]string, error) {
	msg := &wasmdagpb.EmitMessageRequest{}
	if err := proto.Unmarshal(b, msg); err != nil {
		return "", "", nil, err
	}

	validated, err := toValidatedLabels(msg)
	if err != nil {
		return "", "", nil, err
	}

	return msg.RequestId, msg.Message, validated, nil
}

func toValidatedLabels(msg *wasmdagpb.EmitMessageRequest) (map[string]string, error) {
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
func wasmWrite(caller *wasmtime.Caller, src []byte, ptr int32, maxSize int32) int64 {
	return write(wasmMemoryAccessor(caller), src, ptr, maxSize)
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

func truncateWasmWrite(caller *wasmtime.Caller, src []byte, ptr int32, size int32) int64 {
	memory := wasmMemoryAccessor(caller)
	if int32(len(memory)) < ptr+size {
		size = int32(len(memory)) - ptr
		src = src[:size]
	}

	return write(memory, src, ptr, size)
}

// write copies the given src byte slice into the memory at the given pointer and max size.
func write(memory, src []byte, ptr, maxSize int32) int64 {
	if ptr < 0 {
		return -1
	}

	if len(src) > int(maxSize) {
		return -1
	}

	if int32(len(memory)) < ptr+maxSize {
		return -1
	}
	buffer := memory[ptr : ptr+int32(len(src))]
	return int64(copy(buffer, src))
}

func createCallCapFn(
	logger logger.Logger,
	store *store[*wasmpb.ExecutionResult],
	callCapAsync func(ctx context.Context, req *wasmpb.CapabilityRequest) ([sdk.IdLen]byte, error)) func(caller *wasmtime.Caller, ptr int32, ptrlen int32, idPtr int32) int64 {
	return func(caller *wasmtime.Caller, ptr int32, ptrlen int32, idPtr int32) int64 {
		b, innerErr := wasmRead(caller, ptr, ptrlen)
		if innerErr != nil {
			errStr := fmt.Sprintf("error calling wasmRead: %s", innerErr)
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), ptr, ptrlen)
		}

		req := &wasmpb.CapabilityRequest{}
		innerErr = proto.Unmarshal(b, req)
		if innerErr != nil {
			errStr := fmt.Sprintf("error calling proto unmarshal: %s", innerErr)
			logger.Errorf(errStr)
			return truncateWasmWrite(caller, []byte(errStr), ptr, ptrlen)
		}

		reqData, err := store.get(req.ExecutionId)
		if err != nil {
			errStr := fmt.Sprintf("error calling store.get: %s", err)
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), ptr, ptrlen)
		}

		id, err := callCapAsync(reqData.ctx(), req)
		if err != nil {
			errStr := fmt.Sprintf("error calling callCapAsync: %s", err)
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), ptr, ptrlen)
		}

		writeLen := wasmWrite(caller, id[:], idPtr, sdk.IdLen)
		if writeLen < 0 {
			errStr := fmt.Sprintf("Response size %d is too large for wasm memory", sdk.IdLen)
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), ptr, ptrlen)
		}

		return writeLen
	}
}

func createAwaitCapsFn(
	logger logger.Logger,
	store *store[*wasmpb.ExecutionResult],
	awaitCaps func(ctx context.Context, req *wasmpb.AwaitCapabilitiesRequest) (*wasmpb.AwaitCapabilitiesResponse, error),
) func(caller *wasmtime.Caller, awaitRequest, awaitRequestLen, responseBuffer, maxResponseLen int32) int64 {
	return func(caller *wasmtime.Caller, awaitRequest, awaitRequestLen, responseBuffer, maxResponseLen int32) int64 {
		b, err := wasmRead(caller, awaitRequest, awaitRequestLen)
		if err != nil {
			errStr := fmt.Sprintf("error reading from wasm %s", err)
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
		}

		req := &wasmpb.AwaitCapabilitiesRequest{}
		err = proto.Unmarshal(b, req)
		if err != nil {
			errStr := err.Error()
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
		}

		reqData, err := store.get(req.ExecId)
		if err != nil {
			errStr := fmt.Sprintf("error calling store.get: %s", err)
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
		}

		resp, err := awaitCaps(reqData.ctx(), req)
		if err != nil {
			errStr := err.Error()
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
		}

		respBytes, err := proto.Marshal(resp)
		if err != nil {
			errStr := err.Error()
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
		}

		size := wasmWrite(caller, respBytes, responseBuffer, maxResponseLen)
		if size == -1 {
			errStr := "cannot write to wasm"
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
		}

		return size
	}
}
