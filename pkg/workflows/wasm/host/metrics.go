package host

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

// Execution phase labels for the execution duration histogram:
//   - phaseWasm:    wall-clock time executing guest wasm code inside callWasm.
//   - phaseWaiting: wall-clock time suspended, waiting for pending capability
//     responses.
//   - phaseTotal:   total wall-clock time for the execution, end to end.
const (
	phaseWasm    = "wasm"
	phaseWaiting = "waiting"
	phaseTotal   = "total"
)

// moduleMetrics holds the beholder instruments used to observe wasm module
// executions. Instrument names are shared process-wide, so multiple modules can
// safely construct their own moduleMetrics; the meter returns the same
// underlying instrument for a given name.
type moduleMetrics struct {
	activeExecutions    metric.Int64UpDownCounter
	suspendedExecutions metric.Int64UpDownCounter
	suspensionsPerExec  metric.Int64Histogram
	executionDurationMs metric.Int64Histogram
	memoryBytes         metric.Int64Histogram
}

func newModuleMetrics() (*moduleMetrics, error) {
	meter := beholder.GetMeter()

	activeExecutions, err := meter.Int64UpDownCounter("platform_wasm_host_active_executions",
		metric.WithDescription("Number of wasm module executions currently running"),
		metric.WithUnit("{execution}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create active_executions counter: %w", err)
	}

	suspendedExecutions, err := meter.Int64UpDownCounter("platform_wasm_host_suspended_executions",
		metric.WithDescription("Number of wasm module executions currently suspended waiting for capability responses"),
		metric.WithUnit("{execution}"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create suspended_executions counter: %w", err)
	}

	suspensionsPerExec, err := meter.Int64Histogram("platform_wasm_host_suspensions_per_execution",
		metric.WithDescription("Number of times an execution suspended to await capability responses before completing"),
		metric.WithUnit("{suspension}"),
		metric.WithExplicitBucketBoundaries(0, 1, 2, 3, 5, 10, 20, 50, 100),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create suspensions_per_execution histogram: %w", err)
	}

	executionDurationMs, err := meter.Int64Histogram("platform_wasm_host_execution_duration_ms",
		metric.WithDescription("Wall-clock time spent in an execution, by phase (wasm, waiting) plus the end-to-end total"),
		metric.WithUnit("ms"),
		metric.WithExplicitBucketBoundaries(1, 5, 10, 50, 100, 250, 500, 1_000, 2_000, 5_000, 10_000, 30_000, 60_000, 120_000, 300_000, 600_000),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create execution_duration_ms histogram: %w", err)
	}

	memoryBytes, err := meter.Int64Histogram("platform_wasm_host_memory_bytes",
		metric.WithDescription("Peak linear memory in bytes used by the wasm module across an execution"),
		metric.WithUnit("By"),
		metric.WithExplicitBucketBoundaries(1<<20, 4<<20, 16<<20, 32<<20, 64<<20, 128<<20, 256<<20, 512<<20, 1<<30),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create memory_bytes histogram: %w", err)
	}

	return &moduleMetrics{
		activeExecutions:    activeExecutions,
		suspendedExecutions: suspendedExecutions,
		suspensionsPerExec:  suspensionsPerExec,
		executionDurationMs: executionDurationMs,
		memoryBytes:         memoryBytes,
	}, nil
}

// suspensionEnabledAttr tags a metric with whether the execution has
// suspend/resume-on-await enabled, so the two populations can be distinguished.
func suspensionEnabledAttr(suspensionEnabled bool) attribute.KeyValue {
	return attribute.Bool("suspension_enabled", suspensionEnabled)
}

// IncActiveExecutions marks an execution as started (a).
func (m *moduleMetrics) IncActiveExecutions(ctx context.Context, suspensionEnabled bool) {
	m.activeExecutions.Add(ctx, 1, metric.WithAttributes(suspensionEnabledAttr(suspensionEnabled)))
}

// DecActiveExecutions marks an execution as finished (a).
func (m *moduleMetrics) DecActiveExecutions(ctx context.Context, suspensionEnabled bool) {
	m.activeExecutions.Add(ctx, -1, metric.WithAttributes(suspensionEnabledAttr(suspensionEnabled)))
}

// IncSuspendedExecutions marks an execution as suspended, waiting for capability
// responses (b).
func (m *moduleMetrics) IncSuspendedExecutions(ctx context.Context, suspensionEnabled bool) {
	m.suspendedExecutions.Add(ctx, 1, metric.WithAttributes(suspensionEnabledAttr(suspensionEnabled)))
}

// DecSuspendedExecutions marks a suspended execution as resumed (b).
func (m *moduleMetrics) DecSuspendedExecutions(ctx context.Context, suspensionEnabled bool) {
	m.suspendedExecutions.Add(ctx, -1, metric.WithAttributes(suspensionEnabledAttr(suspensionEnabled)))
}

// RecordSuspensions records how many times an execution suspended before
// completing (c).
func (m *moduleMetrics) RecordSuspensions(ctx context.Context, suspensionEnabled bool, suspensions int64) {
	m.suspensionsPerExec.Record(ctx, suspensions, metric.WithAttributes(suspensionEnabledAttr(suspensionEnabled)))
}

// RecordExecutionPhase records the wall-clock time spent in a single phase of an
// execution (d). phase is one of phaseWasm, phaseWaiting or phaseTotal.
func (m *moduleMetrics) RecordExecutionPhase(ctx context.Context, suspensionEnabled bool, phase string, d time.Duration) {
	m.executionDurationMs.Record(ctx, d.Milliseconds(),
		metric.WithAttributes(suspensionEnabledAttr(suspensionEnabled), attribute.String("phase", phase)),
	)
}

// RecordMemory records the peak linear memory used by an execution (e). Note
// that the CPU-time counterpart of (e) is measured as wall-clock time spent in
// wasm - the phaseWasm bucket of the execution duration histogram (d).
func (m *moduleMetrics) RecordMemory(ctx context.Context, suspensionEnabled bool, memoryBytes int64) {
	m.memoryBytes.Record(ctx, memoryBytes, metric.WithAttributes(suspensionEnabledAttr(suspensionEnabled)))
}
