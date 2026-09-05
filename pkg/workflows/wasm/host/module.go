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
	"sync"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/bytecodealliance/wasmtime-go/v48"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/host"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/custmsg"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/cresettings"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/limits"
	dagsdk "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm"
	wasmdagpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

const v2ImportPrefix = "version_v2"

// memoryExportName is the export name the host memory accessors resolve the
// guest's linear memory by.
const memoryExportName = "memory"

// callCapabilityV2ParamCount is the number of params the V2 call_capability
// import declares (req, reqLen, responseBuffer, maxResponseLen).
const callCapabilityV2ParamCount = 4

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

	// MaxSubscriptionsLimiter bounds nsubscriptions in the WASI poll_oneoff host
	// call. It uses the default value of cresettings.Default.WASMPollOneoffSubscriptionLimit;
	// inject a limiter to provide dynamic settings.
	MaxSubscriptionsLimiter limits.BoundLimiter[int]

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

	// guestStdoutFile and guestStderrFile are the paths the WASM guest's stdout/stderr are
	// redirected to. They always default to os.DevNull so the guest can never write to the
	// host's own stdout/stderr; unexported so callers outside this package can't override
	// that. Tests in this package may set them directly to a temp file to inspect guest output.
	guestStdoutFile string
	guestStderrFile string
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
	engine  *wasmtime.Engine
	module  *wasmtime.Module
	wconfig *wasmtime.Config

	cfg             *ModuleConfig
	defaultLimiters moduleLimiters

	metrics moduleMetrics

	wg     sync.WaitGroup
	stopCh chan struct{}

	v2ImportName string

	// callCapParams records the number of parameters the guest's
	// call_capability import declares. 2 = legacy V1 (no response buffer),
	// 4 = V2 (with response buffer). 0 = not imported (e.g. legacy DAG).
	callCapParams int

	// linkV2 wires the host functions the v2/NoDAG guest imports. It defaults
	// to linkNoDAG; tests may override it to substitute a host function
	// implementation (e.g. one that panics) without going through a real
	// compiled wasm binary.
	linkV2 linkFn[*sdkpb.ExecutionResult]
}

var _ ModuleV1 = (*module)(nil)

type linkFn[T any] func(ctx context.Context, m *module, store *wasmtime.Store, exec *execution[T]) (*wasmtime.Instance, error)

type moduleLimiters struct {
	pendingCalls                  limits.ResourcePoolLimiter[int]
	enableUserMetrics             limits.GateLimiter
	maxUserMetricPayload          limits.BoundLimiter[config.Size]
	maxUserMetricNameLength       limits.BoundLimiter[int]
	maxUserMetricLabelsPerMetric  limits.BoundLimiter[int]
	maxUserMetricLabelValueLength limits.BoundLimiter[int]
	memory                        limits.BoundLimiter[config.Size]
	maxCompressedBinary           limits.BoundLimiter[config.Size]
	maxDecompressedBinary         limits.BoundLimiter[config.Size]
	maxResponseSize               limits.BoundLimiter[config.Size]
	maxSubscriptions              limits.BoundLimiter[int]
}

func (l *moduleLimiters) close() {
	closers := [...]io.Closer{
		l.pendingCalls,
		l.enableUserMetrics,
		l.maxUserMetricPayload,
		l.maxUserMetricNameLength,
		l.maxUserMetricLabelsPerMetric,
		l.maxUserMetricLabelValueLength,
		l.memory,
		l.maxCompressedBinary,
		l.maxDecompressedBinary,
		l.maxResponseSize,
		l.maxSubscriptions,
	}
	for i := len(closers) - 1; i >= 0; i-- {
		if closers[i] != nil {
			_ = closers[i].Close()
		}
	}
}

func limiterOrDefault[T io.Closer](configured, defaultLimiter T) T {
	if any(configured) != nil {
		return configured
	}
	return defaultLimiter
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

// NewModule creates a WASM module. Limiters omitted from modCfg are created
// internally, owned by the returned module, and not written back to modCfg.
// Caller-provided or subsequently configured limiters remain caller-owned.
// Limiter fields may be set or replaced after construction, but must not be
// reset to nil.
func NewModule(ctx context.Context, modCfg *ModuleConfig, binary []byte, opts ...func(*ModuleConfig)) (*module, error) {
	var defaultLimiters moduleLimiters
	cleanupLimiters := true
	defer func() {
		if cleanupLimiters {
			defaultLimiters.close()
		}
	}()

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

	if modCfg.guestStdoutFile == "" {
		modCfg.guestStdoutFile = os.DevNull
	}

	if modCfg.guestStderrFile == "" {
		modCfg.guestStderrFile = os.DevNull
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

	if modCfg.PendingCallsLimiter == nil {
		limiter, err := limits.MakeResourcePoolLimiter(lf, cresettings.Default.PerWorkflow.CapabilityConcurrencyLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to make pending calls limiter: %w", err)
		}
		defaultLimiters.pendingCalls = limiter
	}

	if modCfg.EnableUserMetricsLimiter == nil {
		defaultLimiters.enableUserMetrics = limits.NewGateLimiter(false)
	}

	if modCfg.MaxUserMetricPayloadLimiter == nil {
		defaultLimiters.maxUserMetricPayload = limits.NewUpperBoundLimiter(config.Size(modCfg.MaxUserMetricPayloadBytes))
	}
	if modCfg.MaxUserMetricNameLengthLimiter == nil {
		defaultLimiters.maxUserMetricNameLength = limits.NewUpperBoundLimiter(int(modCfg.MaxUserMetricNameLength))
	}
	if modCfg.MaxUserMetricLabelsPerMetricLimiter == nil {
		defaultLimiters.maxUserMetricLabelsPerMetric = limits.NewUpperBoundLimiter(int(modCfg.MaxUserMetricLabelsPerMetric))
	}
	if modCfg.MaxUserMetricLabelValueLengthLimiter == nil {
		defaultLimiters.maxUserMetricLabelValueLength = limits.NewUpperBoundLimiter(int(modCfg.MaxUserMetricLabelValueLength))
	}
	if modCfg.MemoryLimiter == nil {
		// Take the max of the min and the configured max memory mbs.
		// We do this because Go requires a minimum of 16 megabytes to run,
		// and local testing has shown that with less than the min, some
		// binaries may error sporadically.
		modCfg.MaxMemoryMBs = uint64(math.Max(float64(modCfg.MinMemoryMBs), float64(modCfg.MaxMemoryMBs)))
		defaultLimiters.memory = limits.NewUpperBoundLimiter(config.Size(modCfg.MaxMemoryMBs) * config.MByte)
	}
	if modCfg.MaxCompressedBinaryLimiter == nil {
		defaultLimiters.maxCompressedBinary = limits.NewUpperBoundLimiter(config.Size(modCfg.MaxCompressedBinarySize))
	}
	if modCfg.MaxDecompressedBinaryLimiter == nil {
		defaultLimiters.maxDecompressedBinary = limits.NewUpperBoundLimiter(config.Size(modCfg.MaxDecompressedBinarySize))
	}
	if modCfg.MaxResponseSizeLimiter == nil {
		defaultLimiters.maxResponseSize = limits.NewUpperBoundLimiter(config.Size(modCfg.MaxResponseSizeBytes))
	}
	if modCfg.MaxSubscriptionsLimiter == nil {
		defaultLimiters.maxSubscriptions = limits.NewUpperBoundLimiter(cresettings.Default.WASMPollOneoffSubscriptionLimit.DefaultValue)
	}

	maxCompressedBinaryLimiter := limiterOrDefault(modCfg.MaxCompressedBinaryLimiter, defaultLimiters.maxCompressedBinary)
	maxDecompressedBinaryLimiter := limiterOrDefault(modCfg.MaxDecompressedBinaryLimiter, defaultLimiters.maxDecompressedBinary)

	if !modCfg.IsUncompressed {
		// validate the binary size before decompressing
		// this is to prevent decompression bombs
		if err := maxCompressedBinaryLimiter.Check(ctx, config.SizeOf(binary)); err != nil {
			if errors.Is(err, limits.ErrorBoundLimited[config.Size]{}) {
				return nil, fmt.Errorf("compressed binary size exceeds the maximum allowed size: %w", err)
			}
			return nil, fmt.Errorf("failed to check compressed binary size limit: %w", err)
		}
		maxDecompressedBinarySize, err := maxDecompressedBinaryLimiter.Limit(ctx)
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
	if err := maxDecompressedBinaryLimiter.Check(ctx, config.SizeOf(binary)); err != nil {
		if errors.Is(err, limits.ErrorBoundLimited[config.Size]{}) {
			return nil, fmt.Errorf("decompressed binary size reached the maximum allowed size: %w", err)
		}
		return nil, fmt.Errorf("failed to check decompressed binary size limit: %w", err)
	}

	metrics, err := newModuleMetrics()
	if err != nil {
		return nil, fmt.Errorf("failed to create module metrics: %w", err)
	}

	m, err := newModule(modCfg, binary, metrics)
	if err != nil {
		return nil, err
	}
	m.defaultLimiters = defaultLimiters
	cleanupLimiters = false
	return m, nil
}

func newModule(modCfg *ModuleConfig, binary []byte, metrics moduleMetrics) (*module, error) {
	cfg := wasmtime.NewConfig()
	cfg.SetEpochInterruption(true)
	if modCfg.InitialFuel > 0 {
		cfg.SetConsumeFuel(true)
	}
	if err := cfg.CacheConfigLoadDefault(); err != nil {
		modCfg.Logger.Errorw("failed to load cache config, continuing without cache", "error", err)
	}
	cfg.SetCraneliftOptLevel(wasmtime.OptLevelSpeedAndSize)
	SetUnwinding(cfg) // Handled differently based on host OS.

	engine := wasmtime.NewEngineWithConfig(cfg)

	mod, err := wasmtime.NewModule(engine, binary)
	if err != nil {
		engine.Close()
		return nil, fmt.Errorf("error creating wasmtime module: %w", err)
	}

	// Every host callback reaches the guest through its exported linear memory,
	// so a module that doesn't export one under the expected name can't be run
	// at all. Reject it here rather than letting the first callback dereference
	// a missing or wrong-typed export.
	if err = requireMemoryExport(mod); err != nil {
		mod.Close()
		engine.Close()
		return nil, err
	}

	v2ImportName := ""
	callCapParams := 0
	for _, modImport := range mod.Imports() {
		name := modImport.Name()
		if modImport.Module() == "env" && name != nil {
			if strings.HasPrefix(*name, v2ImportPrefix) {
				v2ImportName = *name
			}
			if *name == "call_capability" {
				if ft := modImport.Type().FuncType(); ft != nil {
					callCapParams = len(ft.Params())
				}
			}
		}
	}

	modCfg.SdkLabeler(v2ImportName)

	return &module{
		engine:        engine,
		module:        mod,
		wconfig:       cfg,
		cfg:           modCfg,
		metrics:       metrics,
		stopCh:        make(chan struct{}),
		v2ImportName:  v2ImportName,
		callCapParams: callCapParams,
		linkV2:        linkNoDAG,
	}, nil
}

func linkNoDAG(_ context.Context, m *module, store *wasmtime.Store, exec *execution[*sdkpb.ExecutionResult]) (*wasmtime.Instance, error) {
	linker, err := newWasiLinker(exec, m.engine)
	if err != nil {
		return nil, err
	}
	defer linker.Close()

	if err = linker.FuncWrap(
		"env",
		m.v2ImportName,
		func(caller *wasmtime.Caller) {}); err != nil {
		return nil, fmt.Errorf("error wrapping log func: %w", err)
	}

	logger := m.cfg.Logger
	if err = linker.FuncWrap(
		"env",
		"send_response",
		createSendResponseFn(logger, exec, func() *sdkpb.ExecutionResult {
			return &sdkpb.ExecutionResult{}
		}),
	); err != nil {
		return nil, fmt.Errorf("error wrapping sendResponse func: %w", err)
	}

	if err = linker.FuncWrap(
		"env",
		"call_capability",
		createCallCapFn(logger, exec, m.callCapParams),
	); err != nil {
		return nil, fmt.Errorf("error wrapping callcap func: %w", err)
	}

	if err = linker.FuncWrap(
		"env",
		"await_capabilities",
		createAwaitCapsFn(logger, exec),
	); err != nil {
		return nil, fmt.Errorf("error wrapping awaitcaps func: %w", err)
	}

	if err = linker.FuncWrap(
		"env",
		"get_secrets",
		createGetSecretsFn(logger, exec),
	); err != nil {
		return nil, fmt.Errorf("error wrapping get_secrets func: %w", err)
	}

	if err = linker.FuncWrap(
		"env",
		"await_secrets",
		createAwaitSecretsFn(logger, exec),
	); err != nil {
		return nil, fmt.Errorf("error wrapping await_secrets func: %w", err)
	}

	if err = linker.FuncWrap(
		"env",
		"log",
		exec.log,
	); err != nil {
		return nil, fmt.Errorf("error wrapping log func: %w", err)
	}

	if err = linker.FuncWrap(
		"env",
		"emit_metric",
		exec.emitMetric,
	); err != nil {
		return nil, fmt.Errorf("error wrapping emit_metric func: %w", err)
	}

	if err = linker.FuncWrap(
		"env",
		"switch_modes",
		exec.switchModes); err != nil {
		return nil, fmt.Errorf("error wrapping switchModes func: %w", err)
	}

	if err = linker.FuncWrap(
		"env",
		"random_seed",
		exec.getSeed); err != nil {
		return nil, fmt.Errorf("error wrapping getSeed func: %w", err)
	}

	if err = linker.FuncWrap(
		"env",
		"now",
		exec.now); err != nil {
		return nil, fmt.Errorf("error wrapping get_time func: %w", err)
	}

	return linker.Instantiate(store, m.module)
}

func linkLegacyDAG(ctx context.Context, m *module, store *wasmtime.Store, exec *execution[*wasmdagpb.Response]) (*wasmtime.Instance, error) {
	linker, err := newDagWasiLinker(ctx, m)
	if err != nil {
		return nil, err
	}
	defer linker.Close()

	logger := m.cfg.Logger

	if err = linker.FuncWrap(
		"env",
		"sendResponse",
		createSendResponseFn(logger, exec, func() *wasmdagpb.Response {
			return &wasmdagpb.Response{}
		}),
	); err != nil {
		return nil, fmt.Errorf("error wrapping sendResponse func: %w", err)
	}

	err = linker.FuncWrap(
		"env",
		"fetch",
		createFetchFn(logger, wasmRead, wasmWrite, wasmWriteUInt32, m.cfg, exec),
	)
	if err != nil {
		return nil, fmt.Errorf("error wrapping fetch func: %w", err)
	}

	err = linker.FuncWrap(
		"env",
		"emit",
		createEmitFn(logger, exec, m.cfg.Labeler, wasmRead, wasmWrite, wasmWriteUInt32),
	)
	if err != nil {
		return nil, fmt.Errorf("error wrapping emit func: %w", err)
	}

	if err := linker.FuncWrap(
		"env",
		"log",
		createLogFn(logger),
	); err != nil {
		return nil, fmt.Errorf("error wrapping log func: %w", err)
	}

	return linker.Instantiate(store, m.module)
}

func (m *module) Start() {
	m.wg.Go(func() {
		ticker := time.NewTicker(m.cfg.TickInterval)
		defer ticker.Stop()
		for {
			select {
			case <-m.stopCh:
				return
			case <-ticker.C:
				m.engine.IncrementEpoch()
			}
		}
	})
}

// Close may wait for a blocked acquisition from the internally owned pending
// calls limiter.
func (m *module) Close() {
	close(m.stopCh)
	m.wg.Wait()

	m.defaultLimiters.close()
	m.engine.Close()
	m.module.Close()
	m.wconfig.Close()
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

	setMaxResponseSize := func(r *sdkpb.ExecuteRequest, maxSize uint64) {
		r.MaxResponseSize = maxSize
	}

	timeout := *m.cfg.Timeout
	switch req.Request.(type) {
	case *sdkpb.ExecuteRequest_PreHook:
		timeout = *m.cfg.PrehookTimeout
	}
	return runWasm(ctx, m, req, setMaxResponseSize, m.linkV2, executor, timeout)
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

	setMaxResponseSize := func(r *wasmdagpb.Request, maxSize uint64) {
		computeRequest := r.GetComputeRequest()
		if computeRequest != nil {
			computeRequest.RuntimeConfig = &wasmdagpb.RuntimeConfig{
				MaxResponseSizeBytes: int64(maxSize),
			}
		}
	}

	return runWasm(ctx, m, request, setMaxResponseSize, linkLegacyDAG, nil, *m.cfg.Timeout)
}

// callStart looks up and invokes the wasm module's _start function, but
// recovers any panic. It's sufficient to just recover here, wasmtime-go
// recovers panics that originated from guest calls to the host, and
// re-panics them in Go.
// See https://pkg.go.dev/github.com/bytecodealliance/wasmtime-go/v47#Func.Call.
func callStart(m *module, instance *wasmtime.Instance, store *wasmtime.Store) (result any, err error) {
	defer func() {
		if r := recover(); r != nil {
			m.cfg.Logger.Errorw("panic during wasm execution", "panic", r)
			m.metrics.IncHostFnPanicRecovered()
			err = fmt.Errorf("panic during wasm execution: %v", r)
		}
	}()

	start := instance.GetFunc(store, "_start")
	if start == nil {
		return nil, errors.New("could not get start function")
	}

	return start.Call(store)
}

func runWasm[I, O proto.Message](
	ctx context.Context,
	m *module,
	request I,
	setMaxResponseSize func(i I, maxSize uint64),
	linkWasm linkFn[O],
	helper ExecutionHelper,
	maxTimeout time.Duration,
) (O, error) {
	var o O

	// No reason to run the WASM longer if the outer ctx will cancel.
	ctxDeadline, hasDeadline := ctx.Deadline()
	var ctxWithTimeout context.Context
	var cancel func()

	if hasDeadline && ctxDeadline.Before(time.Now().Add(maxTimeout)) {
		ctxWithTimeout, cancel = context.WithCancel(ctx)
	} else {
		ctxWithTimeout, cancel = context.WithTimeout(ctx, maxTimeout)
	}

	defer cancel()

	store := wasmtime.NewStore(m.engine)

	defer store.Close()

	maxResponseSizeBytes, err := limiterOrDefault(m.cfg.MaxResponseSizeLimiter, m.defaultLimiters.maxResponseSize).Limit(ctx)
	if err != nil {
		return o, fmt.Errorf("failed to get response size limit: %w", err)
	}
	setMaxResponseSize(request, uint64(maxResponseSizeBytes))
	reqpb, err := proto.Marshal(request)
	if err != nil {
		return o, err
	}

	reqstr := base64.StdEncoding.EncodeToString(reqpb)

	wasi := wasmtime.NewWasiConfig()
	defer wasi.Close()
	if err := wasi.SetStdoutFile(m.cfg.guestStdoutFile); err != nil {
		return o, fmt.Errorf("error setting guest stdout file: %w", err)
	}
	if err := wasi.SetStderrFile(m.cfg.guestStderrFile); err != nil {
		return o, fmt.Errorf("error setting guest stderr file: %w", err)
	}
	wasi.SetArgv([]string{"wasi", reqstr})

	store.SetWasi(wasi)

	if m.cfg.InitialFuel > 0 {
		err = store.SetFuel(m.cfg.InitialFuel)
		if err != nil {
			return o, fmt.Errorf("error setting fuel: %w", err)
		}
	}

	// Limit memory to max memory megabytes per instance.
	maxMemoryBytes, err := limiterOrDefault(m.cfg.MemoryLimiter, m.defaultLimiters.memory).Limit(ctx)
	if err != nil {
		return o, fmt.Errorf("failed to get memory limit: %w", err)
	}
	store.Limiter(
		int64(maxMemoryBytes/config.MByte)*int64(math.Pow(10, 6)),
		-1, // tableElements, -1 == default
		1,  // instances
		1,  // tables
		1,  // memories
	)

	deadline := maxTimeout / m.cfg.TickInterval
	store.SetEpochDeadline(uint64(deadline))

	h := fnv.New64a()
	if helper != nil {
		executionId := helper.GetWorkflowExecutionID()
		_, _ = h.Write([]byte(executionId))
	}

	donSeed := int64(h.Sum64())

	exec := &execution[O]{
		ctx:                 ctxWithTimeout,
		capabilityResponses: map[int32]<-chan *sdkpb.CapabilityResponse{},
		secretsResponses:    map[int32]<-chan *secretsResponse{},
		usedCallbackIDs:     map[string]bool{},
		pendingCallsLimiter: limiterOrDefault(m.cfg.PendingCallsLimiter, m.defaultLimiters.pendingCalls),
		module:              m,
		executor:            helper,
		donSeed:             donSeed,
		nodeSeed:            int64(rand.Uint64()),
	}

	instance, err := linkWasm(ctxWithTimeout, m, store, exec)
	if err != nil {
		return o, fmt.Errorf("error linking wasm: %w", err)
	}

	startTime := time.Now()
	_, err = callStart(m, instance, store)
	executionDuration := time.Since(startTime)

	// The error codes below are only returned by the v1 legacy DAG workflow.
	switch {
	case containsCode(err, wasm.CodeSuccess):
		if any(exec.response) == nil {
			return o, errors.New("could not find response for execution")
		}
		return exec.response, nil
	case containsCode(err, wasm.CodeInvalidResponse):
		return o, errors.New("invariant violation: error marshaling response")
	case containsCode(err, wasm.CodeInvalidRequest):
		return o, errors.New("invariant violation: invalid request to runner")
	case containsCode(err, wasm.CodeRunnerErr):
		// legacy DAG captured all errors, since the function didn't return an error
		resp, ok := any(exec).(*execution[*wasmdagpb.Response])
		if ok && resp.response != nil {
			return o, fmt.Errorf("error executing runner: %s: %w", resp.response.ErrMsg, err)
		}
		return o, errors.New("error executing runner")
	case containsCode(err, wasm.CodeHostErr):
		return o, errors.New("invariant violation: host errored during sendResponse")
	}

	// If an error has occurred and the deadline has been reached or exceeded, return a deadline exceeded error.
	// Note - there is no other reliable signal on the error that can be used to infer it is due to epoch deadline
	// being reached, so if an error is returned after the deadline it is assumed it is due to that and return
	// context.DeadlineExceeded.
	if err != nil && ((executionDuration >= maxTimeout-m.cfg.TickInterval) || ctx.Err() != nil) { // As start could be called just before epoch update 1 tick interval is deducted to account for this
		m.cfg.Logger.Errorw("start function returned error after deadline reached, returning deadline exceeded error", "errFromStartFunction", err)
		return o, context.DeadlineExceeded
	}

	return o, err
}

func containsCode(err error, code int) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), fmt.Sprintf("exit status %d", code))
}

// createSendResponseFn injects the dependency required by a WASM guest to
// send a response back to the host.
func createSendResponseFn[T proto.Message](
	logger logger.Logger,
	exec *execution[T],
	newT func() T,
) func(caller *wasmtime.Caller, ptr int32, ptrlen int32) int32 {
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
	exec *execution[*wasmdagpb.Response],
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
func createLogFn(logger logger.Logger) func(caller *wasmtime.Caller, ptr int32, ptrlen int32) {
	return func(caller *wasmtime.Caller, ptr int32, ptrlen int32) {
		b, innerErr := wasmRead(caller, ptr, ptrlen)
		if innerErr != nil {
			logger.Errorf("error calling log: %s", innerErr)
			return
		}

		innerErr = logRawMessage(logger, b)
		if innerErr != nil {
			logger.Errorf("error calling log: %s", innerErr)
			return
		}
	}
}

var logRawMessageReg = regexp.MustCompile(`[\r\n\t]|[\x00-\x1F]|[<>\"'\\&%$;:{}\[\]/]`)

// logRawMessage decodes a JSON-encoded log message received from the WASM guest and
// logs it at the appropriate level.
func logRawMessage(logger logger.Logger, b []byte) error {
	var raw map[string]any
	innerErr := json.Unmarshal(b, &raw)
	if innerErr != nil {
		return innerErr
	}

	level := raw["level"]
	delete(raw, "level")

	msg, ok := raw["msg"].(string)
	if !ok {
		return fmt.Errorf("could not coerce msg to string, got %T", raw["msg"])
	}
	delete(raw, "msg")
	delete(raw, "ts")

	var args []any
	for k, v := range raw {
		args = append(args, k, v)
	}

	sanitizedMsg := logRawMessageReg.ReplaceAllString(msg, "*")

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
		// The guest should never be able to panic/exit the host
		logger.Errorw("[panic] "+sanitizedMsg, args...)
	case "fatal":
		// The guest should never be able to panic/exit the host
		logger.Errorw("[fatal] "+sanitizedMsg, args...)
	default:
		logger.Infow(sanitizedMsg, args...)
	}

	return nil
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
	return caller.GetExport(memoryExportName).Memory().UnsafeData(caller)
}

// requireMemoryExport verifies that the module exports linear memory under the
// name the host memory accessors look it up by.
func requireMemoryExport(mod *wasmtime.Module) error {
	for _, export := range mod.Exports() {
		if export.Name() != memoryExportName {
			continue
		}

		if export.Type().MemoryType() == nil {
			return fmt.Errorf("module export %q is not a memory", memoryExportName)
		}

		return nil
	}

	return fmt.Errorf("module does not export %q", memoryExportName)
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
	if ptr < 0 || size < 0 || int64(ptr) > int64(len(memory)) {
		return -1
	}

	// widen to int64 so a guest-supplied ptr/size near math.MaxInt32 cannot
	// overflow the bounds check and drive size negative below.
	if int64(ptr)+int64(size) > int64(len(memory)) {
		size = int32(len(memory)) - ptr
	}
	if int(size) < len(src) {
		src = src[:size]
	}

	// truncateWasmWrite is only called for returning error strings
	// Therefore, we need to return the negated bytes written to indicate the failure to the guest.
	return -write(memory, src, ptr, size)
}

// write copies the given src byte slice into the memory at the given pointer and max size.
func write(memory, src []byte, ptr, maxSize int32) int64 {
	if ptr < 0 || maxSize < 0 {
		return -1
	}

	if len(src) > int(maxSize) {
		return -1
	}

	// widen to int64 so a guest-supplied ptr near math.MaxInt32 cannot overflow
	// the bounds check into a negative value and bypass it.
	if int64(ptr)+int64(maxSize) > int64(len(memory)) {
		return -1
	}
	buffer := memory[ptr : ptr+int32(len(src))]
	return int64(copy(buffer, src))
}

// callCapability is the core host function for call_capability. It reads a
// CapabilityRequest from the request buffer, dispatches it asynchronously via
// callCapAsync, and writes any synchronous error string to the response buffer.
// The return value protocol: >= 0 is success (the async response comes later
// via await_capabilities), < 0 means the error string is in
// responseBuffer[:-returnValue].
func callCapability(
	caller *wasmtime.Caller,
	logger logger.Logger,
	exec *execution[*sdkpb.ExecutionResult],
	ptr, ptrlen, responseBuffer, maxResponseLen int32,
) int64 {
	b, innerErr := wasmRead(caller, ptr, ptrlen)
	if innerErr != nil {
		errStr := fmt.Sprintf("error calling wasmRead: %s", innerErr)
		logger.Error(errStr)
		return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
	}

	req := &sdkpb.CapabilityRequest{}
	innerErr = proto.Unmarshal(b, req)
	if innerErr != nil {
		errStr := fmt.Sprintf("error calling proto unmarshal: %s", innerErr)
		logger.Errorf("%s", errStr)
		return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
	}

	if err := exec.callCapAsync(exec.ctx, req); err != nil {
		errStr := fmt.Sprintf("error calling callCapAsync: %s", err)
		logger.Error(errStr)
		return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
	}

	return 0
}

// createCallCapFnV1 is the legacy 2-param host function for call_capability.
// It passes the request buffer as the response buffer, matching the existing
// behavior where error strings are written to the request buffer. This
// function exists for backward compatibility with WASM modules compiled
// against older SDKs that declare a 2-param call_capability import.
func createCallCapFnV1(
	logger logger.Logger,
	exec *execution[*sdkpb.ExecutionResult],
) func(caller *wasmtime.Caller, ptr int32, ptrlen int32) int64 {
	return func(caller *wasmtime.Caller, ptr int32, ptrlen int32) int64 {
		return callCapability(caller, logger, exec, ptr, ptrlen, ptr, ptrlen)
	}
}

// createCallCapFnV2 is the new 4-param host function for call_capability.
// It accepts a dedicated response buffer and writes error strings to it
// instead of the request buffer.
func createCallCapFnV2(
	logger logger.Logger,
	exec *execution[*sdkpb.ExecutionResult],
) func(caller *wasmtime.Caller, ptr int32, ptrlen int32, responseBuffer int32, maxResponseLen int32) int64 {
	return func(caller *wasmtime.Caller, ptr int32, ptrlen int32, responseBuffer int32, maxResponseLen int32) int64 {
		return callCapability(caller, logger, exec, ptr, ptrlen, responseBuffer, maxResponseLen)
	}
}

// createCallCapFn selects the appropriate call_capability host function
// based on the param count the guest module's import declares. 4 params = V2
// (with response buffer), anything else = V1 (legacy, backward compatible).
func createCallCapFn(logger logger.Logger, exec *execution[*sdkpb.ExecutionResult], callCapParams int) any {
	if callCapParams == callCapabilityV2ParamCount {
		return createCallCapFnV2(logger, exec)
	}
	return createCallCapFnV1(logger, exec)
}

func createAwaitCapsFn(
	logger logger.Logger,
	exec *execution[*sdkpb.ExecutionResult],
) func(caller *wasmtime.Caller, awaitRequest, awaitRequestLen, responseBuffer, maxResponseLen int32) int64 {
	return func(caller *wasmtime.Caller, awaitRequest, awaitRequestLen, responseBuffer, maxResponseLen int32) int64 {
		b, err := wasmRead(caller, awaitRequest, awaitRequestLen)
		if err != nil {
			errStr := fmt.Sprintf("error reading from wasm %s", err)
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
		}

		req := &sdkpb.AwaitCapabilitiesRequest{}
		err = proto.Unmarshal(b, req)
		if err != nil {
			errStr := err.Error()
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
		}

		resp, err := exec.awaitCapabilities(exec.ctx, req)
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
			errStr := ResponseBufferTooSmall
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
		}

		return size
	}
}

func createGetSecretsFn(
	logger logger.Logger,
	exec *execution[*sdkpb.ExecutionResult],
) func(caller *wasmtime.Caller, req, requestLen, responseBuffer, maxResponseLen int32) int64 {
	return func(caller *wasmtime.Caller, req, requestLen, responseBuffer, maxResponseLen int32) int64 {
		b, innerErr := wasmRead(caller, req, requestLen)
		if innerErr != nil {
			errStr := fmt.Sprintf("error calling wasmRead: %s", innerErr)
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
		}

		r := &sdkpb.GetSecretsRequest{}
		innerErr = proto.Unmarshal(b, r)
		if innerErr != nil {
			errStr := fmt.Sprintf("error calling proto unmarshal: %s", innerErr)
			logger.Errorf("%s", errStr)
			return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
		}

		if err := exec.getSecretsAsync(exec.ctx, r); err != nil {
			errStr := fmt.Sprintf("error calling getSecretsAsync: %s", err)
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
		}

		return 0
	}
}

func createAwaitSecretsFn(
	logger logger.Logger,
	exec *execution[*sdkpb.ExecutionResult],
) func(caller *wasmtime.Caller, awaitRequest, awaitRequestLen, responseBuffer, maxResponseLen int32) int64 {
	return func(caller *wasmtime.Caller, awaitRequest, awaitRequestLen, responseBuffer, maxResponseLen int32) int64 {
		b, err := wasmRead(caller, awaitRequest, awaitRequestLen)
		if err != nil {
			errStr := fmt.Sprintf("error reading from wasm %s", err)
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
		}

		req := &sdkpb.AwaitSecretsRequest{}
		err = proto.Unmarshal(b, req)
		if err != nil {
			errStr := err.Error()
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
		}

		resp, err := exec.awaitSecrets(exec.ctx, req)
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
			errStr := ResponseBufferTooSmall
			logger.Error(errStr)
			return truncateWasmWrite(caller, []byte(errStr), responseBuffer, maxResponseLen)
		}

		return size
	}
}
