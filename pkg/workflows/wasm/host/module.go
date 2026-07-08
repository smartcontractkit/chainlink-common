package host

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"math"
	"math/rand"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/host"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/custmsg"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/cresettings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/limits"
	dagsdk "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm"
	wasmdagpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

const v2ImportPrefix = "version_v2"

var (
	defaultTickInterval              = 100 * time.Millisecond
	defaultTimeout                   = 10 * time.Minute
	defaultPrehookTimeout            = 10 * time.Second
	defaultMinMemoryMBs              = uint64(128)
	DefaultInitialFuel               = uint64(100_000_000)
	defaultMaxFetchRequests          = 5
	defaultMaxCompressedBinarySize   = 20 * 1024 * 1024  // 20 MB
	defaultMaxDecompressedBinarySize = 100 * 1024 * 1024 // 100 MB
	defaultMaxResponseSizeBytes      = 5 * 1024 * 1024   // 5 MB
	defaultMaxLogLenBytes            = 1024 * 1024       // 1 MB
	defaultMaxLogCountDONMode        = 10_000
	defaultMaxLogCountNodeMode       = 10_000
	ResponseBufferTooSmall           = "response buffer too small"

	defaultMaxUserMetricPayloadBytes     = uint32(4096) // 4 KB
	defaultMaxUserMetricNameLength       = uint32(128)
	defaultMaxUserMetricLabelsPerMetric  = uint32(10)
	defaultMaxUserMetricLabelValueLength = uint32(256)
)

type DeterminismConfig struct {
	// Seed is the seed used to generate cryptographically insecure random numbers in the module.
	Seed int64
}
type ModuleConfig struct {
	TickInterval     time.Duration
	Timeout          *time.Duration
	PrehookTimeout   *time.Duration
	MaxMemoryMBs     uint64
	MinMemoryMBs     uint64
	MemoryLimiter    limits.BoundLimiter[config.Size] // supersedes Max/MinMemoryMBs if set
	InitialFuel      uint64
	Logger           logger.Logger
	IsUncompressed   bool
	Fetch            func(ctx context.Context, req *FetchRequest) (*FetchResponse, error)
	MaxFetchRequests int
	// PendingCallsLimiter bounds concurrent in-flight capability and secrets
	// calls. When scoped (e.g. ScopeWorkflow), each workflow ID gets its own
	// pool; when global/unscoped, the limit is shared across all callers.
	PendingCallsLimiter          limits.ResourcePoolLimiter[int]
	MaxCompressedBinarySize      uint64
	MaxCompressedBinaryLimiter   limits.BoundLimiter[config.Size] // supersedes MaxCompressedBinarySize if set
	MaxDecompressedBinarySize    uint64
	MaxDecompressedBinaryLimiter limits.BoundLimiter[config.Size] // supersedes MaxDecompressedBinarySize if set
	MaxResponseSizeBytes         uint64
	MaxResponseSizeLimiter       limits.BoundLimiter[config.Size] // supersedes MaxResponseSizeBytes if set

	MaxLogLenBytes      uint32
	MaxLogCountDONMode  uint32
	MaxLogCountNodeMode uint32

	EnableUserMetricsLimiter             limits.GateLimiter
	MaxUserMetricPayloadBytes            uint32
	MaxUserMetricPayloadLimiter          limits.BoundLimiter[config.Size] // supersedes MaxUserMetricPayloadBytes if set
	MaxUserMetricNameLength              uint32
	MaxUserMetricNameLengthLimiter       limits.BoundLimiter[int] // supersedes MaxUserMetricNameLength if set
	MaxUserMetricLabelsPerMetric         uint32
	MaxUserMetricLabelsPerMetricLimiter  limits.BoundLimiter[int] // supersedes MaxUserMetricLabelsPerMetric if set
	MaxUserMetricLabelValueLength        uint32
	MaxUserMetricLabelValueLengthLimiter limits.BoundLimiter[int] // supersedes MaxUserMetricLabelValueLength if set

	// Labeler is used to emit messages from the module.
	Labeler custmsg.MessageEmitter

	// SdkLabeler is called with the discovered v2 import name after module creation.
	// If nil, it defaults to a no-op. Used to add metrics labels (e.g. sdk=name).
	SdkLabeler func(string)

	// If Determinism is set, the module will override the random_get function in the WASI API with
	// the provided seed to ensure deterministic behavior.
	Determinism *DeterminismConfig
}

type ModuleBase = host.ModuleBase

type ModuleV1 interface {
	ModuleBase

	// V1/Legacy API - request either the Workflow Spec or Custom-Compute execution
	Run(ctx context.Context, request *wasmdagpb.Request) (*wasmdagpb.Response, error)
}

type ModuleV2 = host.Module

type ExecutionHelper = host.ExecutionHelper

type module struct {
	// The compiled module is not kept resident for the lifetime of the module
	// instance. Instead the decompressed guest binary is retained together with an
	// on-disk wazero compilation cache (rooted at moduleDir). Each guest invocation
	// (callWasm) spins up its own wazero.Runtime, recompiles the binary - served
	// from the on-disk cache, so it is cheap - runs, and closes the runtime when
	// the invocation returns. Nothing wasm-related is held resident while a
	// suspended execution waits between resumes.
	moduleDir string
	cache     wazero.CompilationCache
	binary    []byte

	cfg *ModuleConfig

	metrics *moduleMetrics

	v2ImportName string

	// callWasm runs a single guest invocation for the v2 (no-DAG) execution path.
	// It is a field so that tests can wrap it - e.g. to perturb the execution
	// between the suspended run and the resume - while still exercising the real
	// Execute loop. It is initialised in newModule to delegate to the package-level
	// callWasm and is invoked by Execute via m.callWasm.
	callWasm callWasmFunc
}

// callWasmFunc runs one guest invocation for the v2 (no-DAG) execution path.
type callWasmFunc func(timeout time.Duration, req *sdkpb.ExecuteRequest, linkWasm linkFn[*sdkpb.ExecutionResult], exec *execution[*sdkpb.ExecutionResult]) (time.Duration, error)

var _ ModuleV1 = (*module)(nil)

// linkFn registers the host-provided imports (the "env" module plus any WASI
// extras) on the runtime before the guest module is instantiated. wazero links
// imports at the runtime level rather than via a per-store linker, so there is
// no instance to return here.
type linkFn[T any] func(m *module, r wazero.Runtime, exec *execution[T]) error

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

func NewModule(ctx context.Context, modCfg *ModuleConfig, binary []byte, opts ...func(*ModuleConfig)) (*module, error) {
	// Apply options to the module config.
	for _, opt := range opts {
		opt(modCfg)
	}

	if modCfg.Logger == nil {
		return nil, errors.New("must provide logger")
	}

	if modCfg.Fetch == nil {
		modCfg.Fetch = func(context.Context, *FetchRequest) (*FetchResponse, error) {
			return nil, errors.New("fetch not implemented")
		}
	}

	if modCfg.MaxFetchRequests == 0 {
		modCfg.MaxFetchRequests = defaultMaxFetchRequests
	}

	if modCfg.PendingCallsLimiter == nil {
		lf := limits.Factory{Logger: modCfg.Logger}
		var err error
		modCfg.PendingCallsLimiter, err = limits.MakeResourcePoolLimiter(lf, cresettings.Default.PerWorkflow.CapabilityConcurrencyLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to make pending calls limiter: %w", err)
		}
	}

	if modCfg.Labeler == nil {
		modCfg.Labeler = &unimplementedMessageEmitter{}
	}

	if modCfg.SdkLabeler == nil {
		modCfg.SdkLabeler = func(string) {}
	}

	if modCfg.TickInterval == 0 {
		modCfg.TickInterval = defaultTickInterval
	}

	if modCfg.Timeout == nil {
		modCfg.Timeout = &defaultTimeout
	}

	if modCfg.PrehookTimeout == nil {
		modCfg.PrehookTimeout = &defaultPrehookTimeout
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

	if modCfg.MaxResponseSizeBytes == 0 {
		modCfg.MaxResponseSizeBytes = uint64(defaultMaxResponseSizeBytes)
	}
	if modCfg.MaxLogLenBytes == 0 {
		modCfg.MaxLogLenBytes = uint32(defaultMaxLogLenBytes)
	}
	if modCfg.MaxLogCountDONMode == 0 {
		modCfg.MaxLogCountDONMode = uint32(defaultMaxLogCountDONMode)
	}
	if modCfg.MaxLogCountNodeMode == 0 {
		modCfg.MaxLogCountNodeMode = uint32(defaultMaxLogCountNodeMode)
	}

	if modCfg.MaxUserMetricPayloadBytes == 0 {
		modCfg.MaxUserMetricPayloadBytes = defaultMaxUserMetricPayloadBytes
	}
	if modCfg.MaxUserMetricNameLength == 0 {
		modCfg.MaxUserMetricNameLength = defaultMaxUserMetricNameLength
	}
	if modCfg.MaxUserMetricLabelsPerMetric == 0 {
		modCfg.MaxUserMetricLabelsPerMetric = defaultMaxUserMetricLabelsPerMetric
	}
	if modCfg.MaxUserMetricLabelValueLength == 0 {
		modCfg.MaxUserMetricLabelValueLength = defaultMaxUserMetricLabelValueLength
	}

	lf := limits.Factory{Logger: modCfg.Logger}

	if modCfg.EnableUserMetricsLimiter == nil {
		modCfg.EnableUserMetricsLimiter = limits.NewGateLimiter(false)
	}

	if modCfg.MaxUserMetricPayloadLimiter == nil {
		limit := settings.Size(config.Size(modCfg.MaxUserMetricPayloadBytes))
		var err error
		modCfg.MaxUserMetricPayloadLimiter, err = limits.MakeUpperBoundLimiter(lf, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to make metric payload size limiter: %w", err)
		}
	}
	if modCfg.MaxUserMetricNameLengthLimiter == nil {
		limit := settings.Int(int(modCfg.MaxUserMetricNameLength))
		var err error
		modCfg.MaxUserMetricNameLengthLimiter, err = limits.MakeUpperBoundLimiter(lf, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to make metric name length limiter: %w", err)
		}
	}
	if modCfg.MaxUserMetricLabelsPerMetricLimiter == nil {
		limit := settings.Int(int(modCfg.MaxUserMetricLabelsPerMetric))
		var err error
		modCfg.MaxUserMetricLabelsPerMetricLimiter, err = limits.MakeUpperBoundLimiter(lf, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to make labels per metric limiter: %w", err)
		}
	}
	if modCfg.MaxUserMetricLabelValueLengthLimiter == nil {
		limit := settings.Int(int(modCfg.MaxUserMetricLabelValueLength))
		var err error
		modCfg.MaxUserMetricLabelValueLengthLimiter, err = limits.MakeUpperBoundLimiter(lf, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to make label value length limiter: %w", err)
		}
	}
	if modCfg.MemoryLimiter == nil {
		// Take the max of the min and the configured max memory mbs.
		// We do this because Go requires a minimum of 16 megabytes to run,
		// and local testing has shown that with less than the min, some
		// binaries may error sporadically.
		modCfg.MaxMemoryMBs = uint64(math.Max(float64(modCfg.MinMemoryMBs), float64(modCfg.MaxMemoryMBs)))
		limit := settings.Size(config.Size(modCfg.MaxMemoryMBs) * config.MByte)
		var err error
		modCfg.MemoryLimiter, err = limits.MakeUpperBoundLimiter(lf, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to make memory limiter: %w", err)
		}
	}
	if modCfg.MaxCompressedBinaryLimiter == nil {
		limit := settings.Size(config.Size(modCfg.MaxCompressedBinarySize))
		var err error
		modCfg.MaxCompressedBinaryLimiter, err = limits.MakeUpperBoundLimiter(lf, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to make compressed binary size limiter: %w", err)
		}
	}
	if modCfg.MaxDecompressedBinaryLimiter == nil {
		limit := settings.Size(config.Size(modCfg.MaxDecompressedBinarySize))
		var err error
		modCfg.MaxDecompressedBinaryLimiter, err = limits.MakeUpperBoundLimiter(lf, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to make  decompressed binary size limiter: %w", err)
		}
	}
	if modCfg.MaxResponseSizeLimiter == nil {
		limit := settings.Size(config.Size(modCfg.MaxResponseSizeBytes))
		var err error
		modCfg.MaxResponseSizeLimiter, err = limits.MakeUpperBoundLimiter(lf, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to make response size limiter: %w", err)
		}
	}

	if !modCfg.IsUncompressed {
		// validate the binary size before decompressing
		// this is to prevent decompression bombs
		if err := modCfg.MaxCompressedBinaryLimiter.Check(ctx, config.SizeOf(binary)); err != nil {
			if errors.Is(err, limits.ErrorBoundLimited[config.Size]{}) {
				return nil, fmt.Errorf("compressed binary size exceeds the maximum allowed size: %w", err)
			}
			return nil, fmt.Errorf("failed to check compressed binary size limit: %w", err)
		}
		maxDecompressedBinarySize, err := modCfg.MaxDecompressedBinaryLimiter.Limit(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get decompressed binary size limit: %w", err)
		}
		rdr := io.LimitReader(brotli.NewReader(bytes.NewBuffer(binary)), int64(maxDecompressedBinarySize+1))
		decompedBinary, err := io.ReadAll(rdr)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress binary: %w", err)
		}

		binary = decompedBinary
	}

	// Validate the decompressed binary size.
	// io.LimitReader prevents decompression bombs by reading up to a set limit, but it will not return an error if the limit is reached.
	// The Read() method will return io.EOF, and ReadAll will gracefully handle it and return nil.
	if err := modCfg.MaxDecompressedBinaryLimiter.Check(ctx, config.SizeOf(binary)); err != nil {
		if errors.Is(err, limits.ErrorBoundLimited[config.Size]{}) {
			return nil, fmt.Errorf("decompressed binary size reached the maximum allowed size: %w", err)
		}
		return nil, fmt.Errorf("failed to check decompressed binary size limit: %w", err)
	}

	return newModule(modCfg, binary)
}

// wasmPageSize is the WebAssembly linear-memory page size in bytes. Memory
// limits are expressed to wazero in pages.
const wasmPageSize = 65536

func newModule(modCfg *ModuleConfig, binary []byte) (*module, error) {
	ctx := context.Background()

	// Each module gets its own on-disk compilation cache. Compiling the guest
	// once here both discovers its imports and warms the cache, so the recompile
	// performed by every callWasm is served from disk rather than re-optimised.
	moduleDir, err := os.MkdirTemp("", "wasm-module-")
	if err != nil {
		return nil, fmt.Errorf("error creating module temp dir: %w", err)
	}

	cache, err := wazero.NewCompilationCacheWithDir(moduleDir)
	if err != nil {
		_ = os.RemoveAll(moduleDir)
		return nil, fmt.Errorf("error creating compilation cache: %w", err)
	}

	r := wazero.NewRuntimeWithConfig(ctx, wazero.NewRuntimeConfig().WithCompilationCache(cache))
	defer r.Close(ctx)

	compiled, err := r.CompileModule(ctx, binary)
	if err != nil {
		_ = cache.Close(ctx)
		_ = os.RemoveAll(moduleDir)
		return nil, fmt.Errorf("error compiling wasm module: %w", err)
	}

	v2ImportName := ""
	for _, imp := range compiled.ImportedFunctions() {
		moduleName, name, ok := imp.Import()
		if ok && moduleName == "env" && strings.HasPrefix(name, v2ImportPrefix) {
			v2ImportName = name
			break
		}
	}

	modCfg.SdkLabeler(v2ImportName)

	metrics, err := newModuleMetrics()
	if err != nil {
		_ = cache.Close(ctx)
		_ = os.RemoveAll(moduleDir)
		return nil, fmt.Errorf("error creating module metrics: %w", err)
	}

	m := &module{
		moduleDir:    moduleDir,
		cache:        cache,
		binary:       binary,
		cfg:          modCfg,
		metrics:      metrics,
		v2ImportName: v2ImportName,
	}
	m.callWasm = func(timeout time.Duration, req *sdkpb.ExecuteRequest, linkWasm linkFn[*sdkpb.ExecutionResult], exec *execution[*sdkpb.ExecutionResult]) (time.Duration, error) {
		return callWasm(timeout, m, req, linkWasm, exec)
	}
	return m, nil
}

func linkNoDAG(m *module, r wazero.Runtime, exec *execution[*sdkpb.ExecutionResult]) error {
	// `now` reads DON/Node time via a background fetcher; start it before the
	// guest can call the import.
	exec.timeFetcher = newTimeFetcher(exec.ctx, exec.executor)
	exec.timeFetcher.Start()

	logger := m.cfg.Logger
	_, err := r.NewHostModuleBuilder("env").
		NewFunctionBuilder().WithFunc(func(context.Context, api.Module) {}).Export(m.v2ImportName).
		NewFunctionBuilder().WithFunc(createSendResponseFn(logger, exec, func() *sdkpb.ExecutionResult {
		return &sdkpb.ExecutionResult{}
	})).Export("send_response").
		NewFunctionBuilder().WithFunc(createCallCapFn(logger, exec)).Export("call_capability").
		NewFunctionBuilder().WithFunc(createAwaitCapsFn(logger, exec)).Export("await_capabilities").
		NewFunctionBuilder().WithFunc(createGetSecretsFn(logger, exec)).Export("get_secrets").
		NewFunctionBuilder().WithFunc(createAwaitSecretsFn(logger, exec)).Export("await_secrets").
		NewFunctionBuilder().WithFunc(exec.log).Export("log").
		NewFunctionBuilder().WithFunc(exec.emitMetric).Export("emit_metric").
		NewFunctionBuilder().WithFunc(exec.switchModes).Export("switch_modes").
		NewFunctionBuilder().WithFunc(exec.getSeed).Export("random_seed").
		NewFunctionBuilder().WithFunc(exec.now).Export("now").
		Instantiate(exec.ctx)
	if err != nil {
		return fmt.Errorf("error instantiating env host module: %w", err)
	}
	return nil
}

func linkLegacyDAG(m *module, r wazero.Runtime, exec *execution[*wasmdagpb.Response]) error {
	logger := m.cfg.Logger
	_, err := r.NewHostModuleBuilder("env").
		NewFunctionBuilder().WithFunc(createSendResponseFn(logger, exec, func() *wasmdagpb.Response {
		return &wasmdagpb.Response{}
	})).Export("sendResponse").
		NewFunctionBuilder().WithFunc(createFetchFn(logger, wasmRead, wasmWrite, wasmWriteUInt32, m.cfg, exec)).Export("fetch").
		NewFunctionBuilder().WithFunc(createEmitFn(logger, exec, m.cfg.Labeler, wasmRead, wasmWrite, wasmWriteUInt32)).Export("emit").
		NewFunctionBuilder().WithFunc(createLogFn(logger)).Export("log").
		Instantiate(exec.ctx)
	if err != nil {
		return fmt.Errorf("error instantiating env host module: %w", err)
	}
	return nil
}

// Start is a no-op: each guest invocation creates and tears down its own wazero
// runtime in callWasm, and execution is bounded by the context deadline rather
// than a shared epoch ticker.
func (m *module) Start() {}

func (m *module) Close() {
	// Release the on-disk compilation cache and its backing temp dir. The wazero
	// runtimes are per-invocation and already closed in callWasm.
	if m.cache != nil {
		_ = m.cache.Close(context.Background())
	}
	_ = os.RemoveAll(m.moduleDir)
}

func (m *module) IsLegacyDAG() bool {
	return m.v2ImportName == ""
}

func (m *module) Execute(ctx context.Context, req *sdkpb.ExecuteRequest, executor ExecutionHelper) (*sdkpb.ExecutionResult, error) {
	if m.IsLegacyDAG() {
		return nil, errors.New("cannot execute a legacy dag workflow")
	}

	if executor == nil {
		return nil, errors.New("invalid capability executor: can't be nil")
	}

	if req == nil {
		return nil, errors.New("invalid request: can't be nil")
	}

	maxResponseSizeBytes, err := m.cfg.MaxResponseSizeLimiter.Limit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get response size limit: %w", err)
	}
	req.MaxResponseSize = uint64(maxResponseSizeBytes)

	h := fnv.New64a()
	executionId := executor.GetWorkflowExecutionID()
	_, _ = h.Write([]byte(executionId))
	donSeed := int64(h.Sum64())

	exec := &execution[*sdkpb.ExecutionResult]{
		capabilityResponses: map[int32]*asyncResponse[sdkpb.CapabilityRequest, sdkpb.CapabilityResponse]{},
		secretsResponses:    map[int32]<-chan *secretsResponse{},
		pendingCallsLimiter: m.cfg.PendingCallsLimiter,
		module:              m,
		executor:            executor,
		donSeed:             donSeed,
		nodeSeed:            int64(rand.Uint64()),
		suspendOnAwait:      req.SuspendOnAwait,
	}

	// The overall timeout bounds the entire execution, including any time spent
	// suspended while waiting for capability responses. Reuse a single
	// deadline-bearing context across every resume so the budget is shared.
	overallTimeout := *m.cfg.Timeout
	switch req.Request.(type) {
	case *sdkpb.ExecuteRequest_PreHook:
		overallTimeout = *m.cfg.PrehookTimeout
	}
	ctxWithTimeout, cancel := context.WithTimeout(ctx, overallTimeout)
	exec.ctx = ctxWithTimeout
	defer cancel()
	overallStart := time.Now()

	m.metrics.IncActiveExecutions(ctx, req.SuspendOnAwait)
	defer m.metrics.DecActiveExecutions(ctx, req.SuspendOnAwait)

	// Accumulated across (re)starts and emitted once the execution completes,
	// regardless of outcome. wasmDuration is the wall-clock time spent executing
	// guest code (which is also the CPU-time signal), waitDuration the time spent
	// suspended waiting for capability responses, and suspensions the number of
	// times the execution suspended.
	var (
		wasmDuration time.Duration
		waitDuration time.Duration
		suspensions  int64
	)
	defer func() {
		m.metrics.RecordExecutionPhase(ctx, req.SuspendOnAwait, phaseWasm, wasmDuration)
		m.metrics.RecordExecutionPhase(ctx, req.SuspendOnAwait, phaseWaiting, waitDuration)
		m.metrics.RecordExecutionPhase(ctx, req.SuspendOnAwait, phaseTotal, time.Since(overallStart))
		m.metrics.RecordSuspensions(ctx, req.SuspendOnAwait, suspensions)
		m.metrics.RecordMemory(ctx, req.SuspendOnAwait, exec.peakMemoryBytes)
	}()

	for {
		// Each (re)start is bounded by the time left in the overall timeout, so a
		// resumed run - and the wait for its capability responses - cannot extend
		// the workflow's total run time beyond the original deadline.
		remaining := overallTimeout - time.Since(overallStart)
		if remaining <= 0 {
			return nil, context.DeadlineExceeded
		}

		executionDuration, err := m.callWasm(remaining, req, linkNoDAG, exec)
		wasmDuration += executionDuration

		switch {
		case containsCode(err, wasm.CodeSuccess):
			if any(exec.response) == nil {
				return nil, errors.New("could not find response for execution")
			}
			return exec.response, nil
		case isSuspendTrap(err):
			m.cfg.Logger.Debugw("received suspension, awaiting responses", "executionID", executionId)
			suspensions++
			// The execution is suspended for as long as we wait for its pending
			// capability responses; track it as such for the duration of the wait.
			m.metrics.IncSuspendedExecutions(ctx, req.SuspendOnAwait)
			waitStart := time.Now()
			// Wait for the pending capability responses, then resume on the next
			// loop iteration with a fresh store.
			werr := func() error {
				for _, id := range exec.awaiting {
					ar, ok := exec.capabilityResponses[id]
					if !ok {
						return fmt.Errorf("missing capability response for awaited callback %d", id)
					}

					if _, werr := ar.wait(ctxWithTimeout); werr != nil {
						return werr
					}
				}
				return nil
			}()
			waitDuration += time.Since(waitStart)
			m.metrics.DecSuspendedExecutions(ctx, req.SuspendOnAwait)
			if werr != nil {
				return nil, werr
			}

			continue
		default:
			// If an error has occurred and the deadline has been reached or exceeded, return a deadline exceeded error.
			// Note - there is no other reliable signal on the error that can be used to infer it is due to epoch deadline
			// being reached, so if an error is returned after the deadline it is assumed it is due to that and return
			// context.DeadlineExceeded.
			if err != nil && ((executionDuration >= remaining-m.cfg.TickInterval) || ctx.Err() != nil) { // As start could be called just before epoch update 1 tick interval is deducted to account for this
				m.cfg.Logger.Errorw("start function returned error after deadline reached, returning deadline exceeded error", "errFromStartFunction", err)
				return nil, context.DeadlineExceeded
			}

			return nil, err
		}
	}
}

// Run is deprecated, use execute instead
func (m *module) Run(ctx context.Context, request *wasmdagpb.Request) (*wasmdagpb.Response, error) {
	if request == nil {
		return nil, errors.New("invalid request: can't be nil")
	}

	if request.Id == "" {
		return nil, errors.New("invalid request: can't be empty")
	}

	if !m.IsLegacyDAG() {
		return nil, errors.New("cannot use Run on a non-legacy dag workflow, use Execute instead")
	}

	maxResponseSizeBytes, err := m.cfg.MaxResponseSizeLimiter.Limit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get response size limit: %w", err)
	}

	computeRequest := request.GetComputeRequest()
	if computeRequest != nil {
		computeRequest.RuntimeConfig = &wasmdagpb.RuntimeConfig{
			MaxResponseSizeBytes: int64(maxResponseSizeBytes),
		}
	}

	exec := &execution[*wasmdagpb.Response]{
		module: m,
	}

	// No reason to run the WASM longer if the outer ctx will cancel.
	// TODO: this should live higher up.
	ctxDeadline, hasDeadline := ctx.Deadline()
	var ctxWithTimeout context.Context
	var cancel func()
	if hasDeadline && ctxDeadline.Before(time.Now().Add(*m.cfg.Timeout)) {
		ctxWithTimeout, cancel = context.WithCancel(ctx)
	} else {
		ctxWithTimeout, cancel = context.WithTimeout(ctx, *m.cfg.Timeout)
	}
	exec.ctx = ctxWithTimeout
	defer cancel()

	execDuration, err := callWasm(*m.cfg.Timeout, m, request, linkLegacyDAG, exec)

	switch {
	case containsCode(err, wasm.CodeSuccess):
		if exec.response == nil {
			return nil, errors.New("could not find response for execution")
		}
		return exec.response, nil
	case containsCode(err, wasm.CodeInvalidResponse):
		return nil, errors.New("invariant violation: error marshaling response")
	case containsCode(err, wasm.CodeInvalidRequest):
		return nil, errors.New("invariant violation: invalid request to runner")
	case containsCode(err, wasm.CodeRunnerErr):
		// legacy DAG captured all errors, since the function didn't return an error
		resp, ok := any(exec).(*execution[*wasmdagpb.Response])
		if ok && resp.response != nil {
			return nil, fmt.Errorf("error executing runner: %s: %w", resp.response.ErrMsg, err)
		}
		return nil, errors.New("error executing runner")
	case containsCode(err, wasm.CodeHostErr):
		return nil, errors.New("invariant violation: host errored during sendResponse")
	}

	// If an error has occurred and the deadline has been reached or exceeded, return a deadline exceeded error.
	// Note - there is no other reliable signal on the error that can be used to infer it is due to epoch deadline
	// being reached, so if an error is returned after the deadline it is assumed it is due to that and return
	// context.DeadlineExceeded.
	if err != nil && ((execDuration >= *m.cfg.Timeout-m.cfg.TickInterval) || ctx.Err() != nil) { // As start could be called just before epoch update 1 tick interval is deducted to account for this
		m.cfg.Logger.Errorw("start function returned error after deadline reached, returning deadline exceeded error", "errFromStartFunction", err)
		return nil, context.DeadlineExceeded
	}

	return nil, err
}

// callWasm performs a single guest invocation in a fresh, self-contained wazero
// runtime. The runtime - and therefore the instance and its linear memory - is
// closed before callWasm returns; everything that must survive a resume lives on
// exec. It returns the error from _start (which encodes the guest exit code as a
// *sys.ExitError, or the suspend trap raised by await_capabilities).
func callWasm[I, O proto.Message](
	timeout time.Duration,
	m *module,
	req I,
	linkWasm linkFn[O],
	exec *execution[O],
) (time.Duration, error) {
	ctx := exec.ctx

	reqpb, err := proto.Marshal(req)
	if err != nil {
		return 0, nil
	}
	reqstr := base64.StdEncoding.EncodeToString(reqpb)

	// Limit memory to max memory megabytes per instance. wazero expresses the
	// limit in pages, so convert from bytes.
	maxMemoryBytes, err := m.cfg.MemoryLimiter.Limit(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get memory limit: %w", err)
	}
	memoryLimitPages := uint32(uint64(maxMemoryBytes) / wasmPageSize)

	// A fresh runtime is used per (re)start; the previous run's runtime has
	// already been closed. WithCloseOnContextDone lets the per-call context
	// deadline (below) interrupt a runaway guest, replacing wasmtime's epoch
	// interruption.
	runtimeCfg := wazero.NewRuntimeConfig().
		WithCompilationCache(m.cache).
		WithCloseOnContextDone(true).
		WithMemoryLimitPages(memoryLimitPages)
	r := wazero.NewRuntimeWithConfig(ctx, runtimeCfg)
	defer r.Close(ctx)

	// Standard WASI (clock, poll, random, stdio, proc_exit, args, ...).
	if _, err = wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		return 0, fmt.Errorf("error instantiating wasi: %w", err)
	}

	// Register the host-provided "env" imports for this execution path.
	if err = linkWasm(m, r, exec); err != nil {
		return 0, fmt.Errorf("error linking wasm: %w", err)
	}

	// Recompile from the on-disk cache (a cheap cache hit) for this invocation
	// only. Nothing compiled is retained between runs, so a suspended execution
	// (which returns from callWasm via the suspend trap) holds nothing while it
	// waits; the next resume recompiles again.
	compiled, err := r.CompileModule(ctx, m.binary)
	if err != nil {
		return 0, fmt.Errorf("error compiling wasm module: %w", err)
	}

	// Do not auto-run _start on instantiation: we call it explicitly below so we
	// can time it and read back the peak memory afterwards.
	modCfg := wazero.NewModuleConfig().
		WithArgs("wasi", reqstr).
		WithStdout(os.Stdout).
		WithSysWalltime().
		WithSysNanotime().
		WithSysNanosleep().
		WithStartFunctions()

	instance, err := r.InstantiateModule(ctx, compiled, modCfg)
	if err != nil {
		return 0, fmt.Errorf("error instantiating wasm module: %w", err)
	}

	start := instance.ExportedFunction("_start")
	if start == nil {
		return 0, errors.New("could not get start function")
	}

	// Bound this invocation by the remaining timeout. When the deadline fires,
	// WithCloseOnContextDone closes the module and the call returns a
	// *sys.ExitError carrying context.DeadlineExceeded.
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	startTime := time.Now()
	_, err = start.Call(callCtx)
	executionDuration := time.Since(startTime)

	// Capture the linear memory the guest grew to before the runtime is closed, so
	// Execute can emit it as the peak memory metric.
	if mem := instance.Memory(); mem != nil {
		if used := int64(mem.Size()); used > exec.peakMemoryBytes {
			exec.peakMemoryBytes = used
		}
	}

	return executionDuration, err
}

// containsCode reports whether err is a guest exit (*sys.ExitError) with the
// given exit code. A successful WASI guest calls proc_exit(0), which wazero
// surfaces as a *sys.ExitError with code 0, so this also matches success.
func containsCode(err error, code int) bool {
	if err == nil {
		return false
	}
	var exitErr *sys.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode() == uint32(code)
	}
	return false
}

// isSuspendTrap reports whether err is the trap raised by await_capabilities to
// suspend the execution (see createAwaitCapsFn). The host function panics with
// errSuspendExecution, which wazero wraps into the error returned by start.Call.
func isSuspendTrap(err error) bool {
	return errors.Is(err, errSuspendExecution)
}

// createSendResponseFn injects the dependency required by a WASM guest to
// send a response back to the host.
func createSendResponseFn[T proto.Message](
	logger logger.Logger,
	exec *execution[T],
	newT func() T) func(_ context.Context, mod api.Module, ptr int32, ptrlen int32) int32 {
	return func(_ context.Context, mod api.Module, ptr int32, ptrlen int32) int32 {
		b, innerErr := wasmRead(mod, ptr, ptrlen)
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

		exec.lock.Lock()
		exec.response = resp
		exec.lock.Unlock()

		return ErrnoSuccess
	}
}

func toSdkReq(req *wasmdagpb.FetchRequest) *FetchRequest {
	h := map[string]string{}
	for k, v := range req.Headers.GetFields() {
		h[k] = v.GetStringValue()
	}

	md := FetchRequestMetadata{}
	if req.Metadata != nil {
		md = FetchRequestMetadata{
			WorkflowID:          req.Metadata.WorkflowId,
			WorkflowName:        req.Metadata.WorkflowName,
			WorkflowOwner:       req.Metadata.WorkflowOwner,
			WorkflowExecutionID: req.Metadata.WorkflowExecutionId,
			DecodedWorkflowName: req.Metadata.DecodedWorkflowName,
		}
	}
	return &FetchRequest{
		FetchRequest: dagsdk.FetchRequest{
			URL:        req.Url,
			Method:     req.Method,
			Headers:    h,
			Body:       req.Body,
			TimeoutMs:  req.TimeoutMs,
			MaxRetries: req.MaxRetries,
		},
		Metadata: md,
	}
}

func fromSdkResp(resp *dagsdk.FetchResponse) (*wasmdagpb.FetchResponse, error) {
	h := map[string]any{}
	if resp.Headers != nil {
		for k, v := range resp.Headers {
			h[k] = v
		}
	}
	m, err := values.WrapMap(h)
	if err != nil {
		return nil, err
	}
	return &wasmdagpb.FetchResponse{
		ExecutionError: resp.ExecutionError,
		ErrorMessage:   resp.ErrorMessage,
		StatusCode:     resp.StatusCode,
		Headers:        values.ProtoMap(m),
		Body:           resp.Body,
	}, nil
}

type FetchRequestMetadata struct {
	WorkflowID          string
	WorkflowName        string
	WorkflowOwner       string
	WorkflowExecutionID string
	DecodedWorkflowName string
}

type FetchRequest struct {
	dagsdk.FetchRequest
	Metadata FetchRequestMetadata
}

// Use an alias here to allow extending the FetchResponse with additional
// metadata in the future, as with the FetchRequest above.
type FetchResponse = dagsdk.FetchResponse

func createFetchFn(
	logger logger.Logger,
	reader unsafeReaderFunc,
	writer unsafeWriterFunc,
	sizeWriter unsafeFixedLengthWriterFunc,
	modCfg *ModuleConfig,
	exec *execution[*wasmdagpb.Response],
) func(_ context.Context, mod api.Module, respptr int32, resplenptr int32, reqptr int32, reqptrlen int32) int32 {
	return func(_ context.Context, mod api.Module, respptr int32, resplenptr int32, reqptr int32, reqptrlen int32) int32 {
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

			if size := writer(mod, respBytes, respptr, int32(len(respBytes))); size == -1 {
				logger.Errorf("%s: %s", errFetchSfx, errors.New("failed to write error response"))
				return ErrnoFault
			}

			if size := sizeWriter(mod, resplenptr, uint32(len(respBytes))); size == -1 {
				logger.Errorf("%s: %s", errFetchSfx, errors.New("failed to write error response length"))
				return ErrnoFault
			}

			return ErrnoSuccess
		}

		b, innerErr := reader(mod, reqptr, reqptrlen)
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

		// limit the number of fetch calls we can make per request
		if exec.fetchRequestsCounter >= modCfg.MaxFetchRequests {
			logger.Errorf("%s: max number of fetch request %d exceeded", errFetchSfx, modCfg.MaxFetchRequests)
			return writeErr(errors.New("max number of fetch requests exceeded"))
		}
		exec.fetchRequestsCounter++

		fetchResp, innerErr := modCfg.Fetch(exec.ctx, toSdkReq(req))
		if innerErr != nil {
			logger.Errorf("%s: %s", errFetchSfx, innerErr)
			return writeErr(innerErr)
		}

		protoResp, innerErr := fromSdkResp(fetchResp)
		if innerErr != nil {
			logger.Errorf("%s: %s", errFetchSfx, innerErr)
			return writeErr(innerErr)
		}

		// convert struct to proto
		respBytes, innerErr := proto.Marshal(protoResp)
		if innerErr != nil {
			logger.Errorf("%s: %s", errFetchSfx, innerErr)
			return writeErr(innerErr)
		}

		if size := writer(mod, respBytes, respptr, int32(len(respBytes))); size == -1 {
			return writeErr(errors.New("failed to write response"))
		}

		if size := sizeWriter(mod, resplenptr, uint32(len(respBytes))); size == -1 {
			return writeErr(errors.New("failed to write response length"))
		}

		return ErrnoSuccess
	}
}

// createEmitFn injects dependencies and builds the emit function exposed by the WASM.  Errors in
// Emit, if any, are returned in the Error Message of the response.
func createEmitFn(
	l logger.Logger,
	exec *execution[*wasmdagpb.Response],
	e custmsg.MessageEmitter,
	reader unsafeReaderFunc,
	writer unsafeWriterFunc,
	sizeWriter unsafeFixedLengthWriterFunc,
) func(_ context.Context, mod api.Module, respptr, resplenptr, msgptr, msglen int32) int32 {
	logErr := func(err error) {
		l.Errorf("error emitting message: %s", err)
	}

	return func(_ context.Context, mod api.Module, respptr, resplenptr, msgptr, msglen int32) int32 {
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

			if size := writer(mod, respBytes, respptr, int32(len(respBytes))); size == -1 {
				logErr(errors.New("failed to write response"))
				return ErrnoFault
			}

			if size := sizeWriter(mod, resplenptr, uint32(len(respBytes))); size == -1 {
				logErr(errors.New("failed to write response length"))
				return ErrnoFault
			}

			return ErrnoSuccess
		}

		b, err := reader(mod, msgptr, msglen)
		if err != nil {
			return writeErr(err)
		}

		_, msg, labels, err := toEmissible(b)
		if err != nil {
			return writeErr(err)
		}

		if err := e.WithMapLabels(labels).Emit(exec.ctx, msg); err != nil {
			return writeErr(err)
		}

		return ErrnoSuccess
	}
}

// createLogFn injects dependencies and builds the log function exposed by the WASM.
func createLogFn(logger logger.Logger) func(_ context.Context, mod api.Module, ptr int32, ptrlen int32) {
	return func(_ context.Context, mod api.Module, ptr int32, ptrlen int32) {
		b, innerErr := wasmRead(mod, ptr, ptrlen)
		if innerErr != nil {
			logger.Errorf("error calling log: %s", innerErr)
			return
		}

		var raw map[string]any
		innerErr = json.Unmarshal(b, &raw)
		if innerErr != nil {
			return
		}

		level := raw["level"]
		delete(raw, "level")

		msg := raw["msg"].(string)
		delete(raw, "msg")
		delete(raw, "ts")

		var args []any
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
type unsafeWriterFunc func(c api.Module, src []byte, ptr, len int32) int64

// unsafeFixedLengthWriterFunc defines behavior for writing a uint32 value to wasm memory at the location defined
// by the ptr.
type unsafeFixedLengthWriterFunc func(c api.Module, ptr int32, val uint32) int64

// unsafeReaderFunc abstractly defines the behavior of reading from WASM memory.  Returns a copy of
// the memory at the given pointer and size.
type unsafeReaderFunc func(c api.Module, ptr, len int32) ([]byte, error)

// wasmMemoryAccessor is the default implementation for unsafely accessing the memory of the WASM module.
// wazero's Memory.Read returns a slice that aliases the underlying linear memory,
// so writes through the returned slice are visible to the guest.
func wasmMemoryAccessor(mod api.Module) []byte {
	mem := mod.Memory()
	if mem == nil {
		return nil
	}
	buf, ok := mem.Read(0, mem.Size())
	if !ok {
		return nil
	}
	return buf
}

// wasmRead returns a copy of the wasm module memory at the given pointer and size.
func wasmRead(mod api.Module, ptr int32, size int32) ([]byte, error) {
	return read(wasmMemoryAccessor(mod), ptr, size)
}

// Read acts on a byte slice that should represent an unsafely accessed slice of memory.  It returns
// a copy of the memory at the given pointer and size.
func read(memory []byte, ptr int32, size int32) ([]byte, error) {
	if size < 0 || ptr < 0 {
		return nil, fmt.Errorf("invalid memory access: ptr: %d, size: %d", ptr, size)
	}

	endLoc := ptr + size
	// users control both ptr and size, a malicious user can overflow them.
	if int(endLoc) > len(memory) || endLoc < 0 {
		return nil, errors.New("out of bounds memory access")
	}

	cd := make([]byte, size)
	copy(cd, memory[ptr:endLoc])
	return cd, nil
}

// wasmWrite copies the given src byte slice into the wasm module memory at the given pointer and size.
func wasmWrite(mod api.Module, src []byte, ptr int32, maxSize int32) int64 {
	return write(wasmMemoryAccessor(mod), src, ptr, maxSize)
}

// wasmWriteUInt32 binary encodes and writes a uint32 to the wasm module memory at the given pointer.
func wasmWriteUInt32(mod api.Module, ptr int32, val uint32) int64 {
	return writeUInt32(wasmMemoryAccessor(mod), ptr, val)
}

// writeUInt32 binary encodes and writes a uint32 to the memory at the given pointer.
func writeUInt32(memory []byte, ptr int32, val uint32) int64 {
	uint32Size := int32(4)
	buffer := make([]byte, uint32Size)
	binary.LittleEndian.PutUint32(buffer, val)
	return write(memory, buffer, ptr, uint32Size)
}

func truncateWasmWrite(mod api.Module, src []byte, ptr int32, size int32) int64 {
	memory := wasmMemoryAccessor(mod)
	if int32(len(memory)) < ptr+size {
		size = int32(len(memory)) - ptr
		src = src[:size]
	}

	// truncateWasmWrite is only called for returning error strings
	// Therefore, we need to return the negated bytes written to indicate the failure to the guest.
	return -write(memory, src, ptr, size)
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
	exec *execution[*sdkpb.ExecutionResult]) func(_ context.Context, mod api.Module, ptr int32, ptrlen int32) int64 {
	return func(_ context.Context, mod api.Module, ptr int32, ptrlen int32) int64 {
		b, innerErr := wasmRead(mod, ptr, ptrlen)
		if innerErr != nil {
			errStr := fmt.Sprintf("error calling wasmRead: %s", innerErr)
			logger.Error(errStr)
			return truncateWasmWrite(mod, []byte(errStr), ptr, ptrlen)
		}

		req := &sdkpb.CapabilityRequest{}
		innerErr = proto.Unmarshal(b, req)
		if innerErr != nil {
			errStr := fmt.Sprintf("error calling proto unmarshal: %s", innerErr)
			logger.Errorf("%s", errStr)
			return truncateWasmWrite(mod, []byte(errStr), ptr, ptrlen)
		}

		if err := exec.callCapAsync(exec.ctx, req); err != nil {
			errStr := fmt.Sprintf("error calling callCapAsync: %s", err)
			logger.Error(errStr)
			// TODO (CAPPL-846): write error to the response buffer, not the request buffer
			return -1
		}

		return 0
	}
}

func createAwaitCapsFn(
	logger logger.Logger,
	exec *execution[*sdkpb.ExecutionResult],
) func(_ context.Context, mod api.Module, awaitRequest, awaitRequestLen, responseBuffer, maxResponseLen int32) int64 {
	return func(_ context.Context, mod api.Module, awaitRequest, awaitRequestLen, responseBuffer, maxResponseLen int32) int64 {
		b, err := wasmRead(mod, awaitRequest, awaitRequestLen)
		if err != nil {
			errStr := fmt.Sprintf("error reading from wasm %s", err)
			logger.Error(errStr)
			return truncateWasmWrite(mod, []byte(errStr), responseBuffer, maxResponseLen)
		}

		req := &sdkpb.AwaitCapabilitiesRequest{}
		err = proto.Unmarshal(b, req)
		if err != nil {
			errStr := err.Error()
			logger.Error(errStr)
			return truncateWasmWrite(mod, []byte(errStr), responseBuffer, maxResponseLen)
		}

		resp, err := exec.awaitCapabilities(exec.ctx, req)
		switch {
		case errors.Is(err, errSuspendExecution):
			// awaitCapabilities has recorded exec.awaiting. Panic to unwind the
			// guest immediately; wazero wraps the panic value into the error
			// returned by start.Call, which the host detects (isSuspendTrap),
			// waits for the pending responses, and resumes by re-instantiating
			// the guest.
			panic(errSuspendExecution)
		case err != nil:
			errStr := err.Error()
			logger.Error(errStr)
			return truncateWasmWrite(mod, []byte(errStr), responseBuffer, maxResponseLen)
		}

		respBytes, err := proto.Marshal(resp)
		if err != nil {
			errStr := err.Error()
			logger.Error(errStr)
			return truncateWasmWrite(mod, []byte(errStr), responseBuffer, maxResponseLen)
		}

		size := wasmWrite(mod, respBytes, responseBuffer, maxResponseLen)
		if size == -1 {
			errStr := ResponseBufferTooSmall
			logger.Error(errStr)
			return truncateWasmWrite(mod, []byte(errStr), responseBuffer, maxResponseLen)
		}

		return size
	}
}

func createGetSecretsFn(
	logger logger.Logger,
	exec *execution[*sdkpb.ExecutionResult]) func(_ context.Context, mod api.Module, req, requestLen, responseBuffer, maxResponseLen int32) int64 {
	return func(_ context.Context, mod api.Module, req, requestLen, responseBuffer, maxResponseLen int32) int64 {
		b, innerErr := wasmRead(mod, req, requestLen)
		if innerErr != nil {
			errStr := fmt.Sprintf("error calling wasmRead: %s", innerErr)
			logger.Error(errStr)
			return truncateWasmWrite(mod, []byte(errStr), responseBuffer, maxResponseLen)
		}

		r := &sdkpb.GetSecretsRequest{}
		innerErr = proto.Unmarshal(b, r)
		if innerErr != nil {
			errStr := fmt.Sprintf("error calling proto unmarshal: %s", innerErr)
			logger.Errorf("%s", errStr)
			return truncateWasmWrite(mod, []byte(errStr), responseBuffer, maxResponseLen)
		}

		if err := exec.getSecretsAsync(exec.ctx, r); err != nil {
			errStr := fmt.Sprintf("error calling getSecretsAsync: %s", err)
			logger.Error(errStr)
			return truncateWasmWrite(mod, []byte(errStr), responseBuffer, maxResponseLen)
		}

		return 0
	}
}

func createAwaitSecretsFn(
	logger logger.Logger,
	exec *execution[*sdkpb.ExecutionResult],
) func(_ context.Context, mod api.Module, awaitRequest, awaitRequestLen, responseBuffer, maxResponseLen int32) int64 {
	return func(_ context.Context, mod api.Module, awaitRequest, awaitRequestLen, responseBuffer, maxResponseLen int32) int64 {
		b, err := wasmRead(mod, awaitRequest, awaitRequestLen)
		if err != nil {
			errStr := fmt.Sprintf("error reading from wasm %s", err)
			logger.Error(errStr)
			return truncateWasmWrite(mod, []byte(errStr), responseBuffer, maxResponseLen)
		}

		req := &sdkpb.AwaitSecretsRequest{}
		err = proto.Unmarshal(b, req)
		if err != nil {
			errStr := err.Error()
			logger.Error(errStr)
			return truncateWasmWrite(mod, []byte(errStr), responseBuffer, maxResponseLen)
		}

		resp, err := exec.awaitSecrets(exec.ctx, req)
		if err != nil {
			errStr := err.Error()
			logger.Error(errStr)
			return truncateWasmWrite(mod, []byte(errStr), responseBuffer, maxResponseLen)
		}

		respBytes, err := proto.Marshal(resp)
		if err != nil {
			errStr := err.Error()
			logger.Error(errStr)
			return truncateWasmWrite(mod, []byte(errStr), responseBuffer, maxResponseLen)
		}

		size := wasmWrite(mod, respBytes, responseBuffer, maxResponseLen)
		if size == -1 {
			errStr := ResponseBufferTooSmall
			logger.Error(errStr)
			return truncateWasmWrite(mod, []byte(errStr), responseBuffer, maxResponseLen)
		}

		return size
	}
}
