package host

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

// moduleMetrics records host-side metrics for wasm module execution.
type moduleMetrics interface {
	IncHostFnPanicRecovered()
	RecordHostCall(ctx context.Context, function string, d time.Duration)
	RecordExecution(ctx context.Context, stats executionStats)
}

type executionStats struct {
	memoryBytes  uint64
	hostCalls    int64
	hostDuration time.Duration
	runDuration  time.Duration
}

func (s executionStats) guestDuration() time.Duration {
	return max(s.runDuration-s.hostDuration, 0)
}

type moduleMetricsImpl struct {
	hostFnPanicRecoveredCount metric.Int64Counter
	hostCallDuration          metric.Int64Histogram
	executionHostCalls        metric.Int64Histogram
	executionMemory           metric.Int64Histogram
	runDuration               metric.Int64Histogram
}

var _ moduleMetrics = &moduleMetricsImpl{}

// 1µs, 4µs, 16µs, 64µs, 256µs, 1.02ms, 4.10ms, 16.4ms, 65.5ms, 262ms, 1.05s, 4.19s, 16.8s, 67.1s, 268s, 1074s (17.9m)
var durationUsBuckets = metric.WithExplicitBucketBoundaries(prometheus.ExponentialBuckets(1, 4, 16)...)

var (
	phaseTotal = metric.WithAttributes(attribute.String("phase", "total"))
	phaseGuest = metric.WithAttributes(attribute.String("phase", "guest"))
	phaseHost  = metric.WithAttributes(attribute.String("phase", "host"))
)

func newModuleMetrics() (moduleMetrics, error) {
	meter := beholder.GetMeter()

	hostFnPanicRecoveredTotal, err := meter.Int64Counter("platform_wasm_host_panic_recovered_total",
		metric.WithDescription("the total number of times we've recovered from a wasmtime-go panic"),
	)
	if err != nil {
		return nil, err
	}

	hostCallDuration, err := meter.Int64Histogram("platform_wasm_host_call_duration_us",
		metric.WithDescription("duration of a single host function call made by a wasm guest"),
		metric.WithUnit("us"),
		durationUsBuckets,
	)
	if err != nil {
		return nil, err
	}

	executionHostCalls, err := meter.Int64Histogram("platform_wasm_execution_host_calls",
		metric.WithDescription("number of host function calls made by a wasm guest during a single execution"),
		// 1, 2, 4, 8, 16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768, 65536
		metric.WithExplicitBucketBoundaries(prometheus.ExponentialBuckets(1, 2, 17)...),
	)
	if err != nil {
		return nil, err
	}

	executionMemory, err := meter.Int64Histogram("platform_wasm_execution_memory_bytes",
		metric.WithDescription("size of the wasm guest's linear memory at the end of an execution"),
		metric.WithUnit("By"),
		// 1MiB, 2MiB, 4MiB, 8MiB, 16MiB, 32MiB, 64MiB, 128MiB, 256MiB, 512MiB, 1GiB
		metric.WithExplicitBucketBoundaries(prometheus.ExponentialBuckets(1<<20, 2, 11)...),
	)
	if err != nil {
		return nil, err
	}

	runDuration, err := meter.Int64Histogram("platform_wasm_run_duration_us",
		metric.WithDescription("wall-clock time of a single wasm run by phase: total, guest (running wasm code) and host (inside host functions)"),
		metric.WithUnit("us"),
		durationUsBuckets,
	)
	if err != nil {
		return nil, err
	}

	return &moduleMetricsImpl{
		hostFnPanicRecoveredCount: hostFnPanicRecoveredTotal,
		hostCallDuration:          hostCallDuration,
		executionHostCalls:        executionHostCalls,
		executionMemory:           executionMemory,
		runDuration:               runDuration,
	}, nil
}

// IncHostFnPanicRecovered records a panic that occurred in a wasm host
// function (e.g. call_capability, get_secrets, log, ...) and was recovered
// before it could propagate out of module.Execute.
func (m *moduleMetricsImpl) IncHostFnPanicRecovered() {
	m.hostFnPanicRecoveredCount.Add(context.Background(), 1)
}

func (m *moduleMetricsImpl) RecordHostCall(ctx context.Context, function string, d time.Duration) {
	m.hostCallDuration.Record(ctx, d.Microseconds(), metric.WithAttributes(attribute.String("function", function)))
}

func (m *moduleMetricsImpl) RecordExecution(ctx context.Context, stats executionStats) {
	m.executionHostCalls.Record(ctx, stats.hostCalls)
	m.executionMemory.Record(ctx, int64(stats.memoryBytes))
	m.runDuration.Record(ctx, stats.runDuration.Microseconds(), phaseTotal)
	m.runDuration.Record(ctx, stats.guestDuration().Microseconds(), phaseGuest)
	m.runDuration.Record(ctx, stats.hostDuration.Microseconds(), phaseHost)
}
