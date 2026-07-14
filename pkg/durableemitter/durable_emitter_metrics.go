package durableemitter

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// DurableEmitterMetricsConfig enables OpenTelemetry metrics for DurableEmitter.
// Set on Config.Metrics; nil disables instrumentation.
//
// When non-nil, an otel Meter must be supplied to NewDurableEmitter so that
// instruments can be registered. DurableEmitter does not look up a global
// meter on its own — callers are responsible for supplying one (usually via
// otel.Meter("durableemitter") or an equivalently scoped meter from their
// telemetry stack).
type DurableEmitterMetricsConfig struct {
	// PollInterval is how often queue and optional process gauges refresh. Zero = 10s.
	PollInterval time.Duration
	// MaxQueuePayloadBytes, if > 0, records capacity_usage_ratio = queue_payload_bytes / max.
	MaxQueuePayloadBytes int64
}

// publishPhase identifies the delivery path recorded by publish metrics.
type publishPhase int

const (
	publishPhaseBatch publishPhase = iota
	publishPhaseRetransmit
)

func (p publishPhase) String() string {
	switch p {
	case publishPhaseBatch:
		return "batch"
	case publishPhaseRetransmit:
		return "retransmit"
	default:
		return "unknown"
	}
}

type durableEmitterMetrics struct {
	clientName         string
	emitSuccess        metric.Int64Counter
	emitFail           metric.Int64Counter
	emitDuration       metric.Float64Histogram
	emitTotalDuration  metric.Float64Histogram
	publishImmOK       metric.Int64Counter
	publishImmErr      metric.Int64Counter
	publishDuration    metric.Float64Histogram
	publishBatchOK     metric.Int64Counter
	publishBatchErr    metric.Int64Counter
	publishBatchEvOK   metric.Int64Counter
	publishBatchEvErr  metric.Int64Counter
	deliverComplete    metric.Int64Counter
	expiredPurged      metric.Int64Counter
	storeOps           metric.Int64Counter
	storeOpDuration    metric.Float64Histogram
	queueDepth         metric.Int64Gauge
	queuePayloadBytes  metric.Int64Gauge
	queueOldestAgeSec  metric.Float64Gauge
	queueTTLBudgetSec  metric.Int64Gauge
	queueCapacityRatio metric.Float64Gauge
	procHeapInuse      metric.Int64Gauge
	procHeapSys        metric.Int64Gauge
	procCPUUser        metric.Float64Gauge
	procCPUSys         metric.Float64Gauge
	// batchEnqueueBufferFull counts events that could not be handed to the
	// batch emitter because its internal queue was full and must be picked up
	// by the retransmit loop instead. Labels: phase={batch,retransmit}.
	batchEnqueueBufferFull metric.Int64Counter
	// insertCoalescerFill reports the write-coalescer channel fill ratio
	// (len/cap). Only meaningful when InsertBatchSize > 0; otherwise 0.
	insertCoalescerFill metric.Float64Gauge
	// deleteCoalescerFill reports the delete-coalescer channel fill ratio
	// (len/cap). Only meaningful when DeleteBatchSize > 0; otherwise 0.
	deleteCoalescerFill metric.Float64Gauge
}

// durationBuckets provides histogram boundaries (in seconds) tuned for
// sub-millisecond through multi-second latencies. The OTel SDK defaults are
// designed for millisecond-scale integer values and produce wildly wrong
// quantile estimates when values are recorded in fractional seconds.
var durationBuckets = metric.WithExplicitBucketBoundaries(
	0.0001, 0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05,
	0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0,
)

// newDurableEmitterMetrics registers all DurableEmitter instruments on the
// supplied meter. The caller is responsible for the meter's scope (the
// instrument prefix below acts as the metric namespace).
func newDurableEmitterMetrics(meter metric.Meter, clientName string) (*durableEmitterMetrics, error) {
	if meter == nil {
		return nil, fmt.Errorf("durable emitter metrics: meter is nil")
	}
	m := &durableEmitterMetrics{
		clientName: clientName,
	}
	var err error
	if m.emitSuccess, err = meter.Int64Counter(
		"durable_emitter.emit.success",
		metric.WithUnit("{call}"),
		metric.WithDescription("Successful durable Emit calls (insert returned)"),
	); err != nil {
		return nil, err
	}
	if m.emitFail, err = meter.Int64Counter(
		"durable_emitter.emit.failure",
		metric.WithUnit("{call}"),
		metric.WithDescription("Failed Emit calls (before or during insert)"),
	); err != nil {
		return nil, err
	}
	if m.emitDuration, err = meter.Float64Histogram(
		"durable_emitter.emit.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Emit insert path duration (seconds, fractional; aligns with Prometheus _duration_seconds); labels: error={true,false}"),
		durationBuckets,
	); err != nil {
		return nil, err
	}
	if m.emitTotalDuration, err = meter.Float64Histogram(
		"durable_emitter.emit.total_duration",
		metric.WithUnit("s"),
		metric.WithDescription("Full Emit() wall time including event construction, DB insert, and channel enqueue (seconds)"),
		durationBuckets,
	); err != nil {
		return nil, err
	}
	if m.publishImmOK, err = meter.Int64Counter(
		"durable_emitter.publish.immediate.success",
		metric.WithUnit("{call}"),
		metric.WithDescription("Immediate Publish RPC successes"),
	); err != nil {
		return nil, err
	}
	if m.publishImmErr, err = meter.Int64Counter(
		"durable_emitter.publish.immediate.failure",
		metric.WithUnit("{call}"),
		metric.WithDescription("Immediate Publish RPC failures (events await retransmit)"),
	); err != nil {
		return nil, err
	}
	if m.publishDuration, err = meter.Float64Histogram(
		"durable_emitter.publish.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Chip Ingress Publish RPC duration; labels: phase={batch,retransmit}, error={true,false}, client_name"),
		durationBuckets,
	); err != nil {
		return nil, err
	}
	if m.publishBatchOK, err = meter.Int64Counter(
		"durable_emitter.publish.retransmit.batch.success",
		metric.WithUnit("{call}"),
		metric.WithDescription("Unused; batch delivery uses batch.events.* counters"),
	); err != nil {
		return nil, err
	}
	if m.publishBatchErr, err = meter.Int64Counter(
		"durable_emitter.publish.retransmit.batch.failure",
		metric.WithUnit("{call}"),
		metric.WithDescription("Unused; batch delivery uses batch.events.* counters"),
	); err != nil {
		return nil, err
	}
	if m.publishBatchEvOK, err = meter.Int64Counter(
		"durable_emitter.publish.batch.events.success",
		metric.WithUnit("{event}"),
		metric.WithDescription("Batch Publish RPC successes (one count per event in a completed batch); labels: phase={batch,retransmit}"),
	); err != nil {
		return nil, err
	}
	if m.publishBatchEvErr, err = meter.Int64Counter(
		"durable_emitter.publish.batch.events.failure",
		metric.WithUnit("{event}"),
		metric.WithDescription("Batch Publish RPC failures (event stays queued); labels: phase={batch,retransmit}"),
	); err != nil {
		return nil, err
	}
	if m.deliverComplete, err = meter.Int64Counter(
		"durable_emitter.delivery.completed",
		metric.WithUnit("{event}"),
		metric.WithDescription("Events removed from store after successful publish (immediate or retransmit)"),
	); err != nil {
		return nil, err
	}
	if m.expiredPurged, err = meter.Int64Counter(
		"durable_emitter.expired_purged",
		metric.WithUnit("{event}"),
		metric.WithDescription("Events deleted by TTL expiry loop"),
	); err != nil {
		return nil, err
	}
	if m.storeOps, err = meter.Int64Counter(
		"durable_emitter.store.operations",
		metric.WithUnit("{op}"),
		metric.WithDescription("Durable store operations (proxy for DB load / IOPs)"),
	); err != nil {
		return nil, err
	}
	if m.storeOpDuration, err = meter.Float64Histogram(
		"durable_emitter.store.operation.duration",
		metric.WithUnit("s"),
		metric.WithDescription("Durable store operation latency (seconds, fractional)"),
		durationBuckets,
	); err != nil {
		return nil, err
	}
	if m.queueDepth, err = meter.Int64Gauge(
		"durable_emitter.queue.depth",
		metric.WithUnit("{row}"),
		metric.WithDescription("Pending rows in durable queue"),
	); err != nil {
		return nil, err
	}
	if m.queuePayloadBytes, err = meter.Int64Gauge(
		"durable_emitter.queue.payload_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("Sum of payload bytes for pending rows"),
	); err != nil {
		return nil, err
	}
	if m.queueOldestAgeSec, err = meter.Float64Gauge(
		"durable_emitter.queue.oldest_pending_age_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Age of oldest pending row at last poll (longest wait)"),
	); err != nil {
		return nil, err
	}
	if m.queueTTLBudgetSec, err = meter.Int64Gauge(
		"durable_emitter.queue.ttl_budget_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Seconds of TTL headroom for the oldest pending event (EventTTL - oldest age); low/negative → DLQ/expiry pressure. Alert engine decides what 'near' means"),
	); err != nil {
		return nil, err
	}
	if m.queueCapacityRatio, err = meter.Float64Gauge(
		"durable_emitter.queue.capacity_usage_ratio",
		metric.WithUnit("1"),
		metric.WithDescription("queue.payload_bytes / MaxQueuePayloadBytes when max > 0"),
	); err != nil {
		return nil, err
	}
	if m.procHeapInuse, err = meter.Int64Gauge(
		"durable_emitter.process.memory.heap_inuse_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("Go runtime MemStats HeapInuse"),
	); err != nil {
		return nil, err
	}
	if m.procHeapSys, err = meter.Int64Gauge(
		"durable_emitter.process.memory.heap_sys_bytes",
		metric.WithUnit("By"),
		metric.WithDescription("Go runtime MemStats HeapSys"),
	); err != nil {
		return nil, err
	}
	if m.procCPUUser, err = meter.Float64Gauge(
		"durable_emitter.process.cpu.user_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Cumulative user CPU seconds (getrusage; Unix only)"),
	); err != nil {
		return nil, err
	}
	if m.procCPUSys, err = meter.Float64Gauge(
		"durable_emitter.process.cpu.system_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("Cumulative system CPU seconds (getrusage; Unix only)"),
	); err != nil {
		return nil, err
	}
	if m.batchEnqueueBufferFull, err = meter.Int64Counter(
		"durable_emitter.batch_enqueue.buffer_full",
		metric.WithUnit("{event}"),
		metric.WithDescription("Events that could not be handed to the batch emitter (buffer full); event remains in DB for retransmit. Labels: phase={batch,retransmit}."),
	); err != nil {
		return nil, err
	}
	if m.insertCoalescerFill, err = meter.Float64Gauge(
		"durable_emitter.insert_coalescer.queue_fill_ratio",
		metric.WithUnit("1"),
		metric.WithDescription("Write-coalescer channel fill ratio (len/cap); 0 when write coalescing is disabled"),
	); err != nil {
		return nil, err
	}
	if m.deleteCoalescerFill, err = meter.Float64Gauge(
		"durable_emitter.delete_coalescer.queue_fill_ratio",
		metric.WithUnit("1"),
		metric.WithDescription("Delete-coalescer channel fill ratio (len/cap); 0 when delete coalescing is disabled"),
	); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *durableEmitterMetrics) recordStoreOp(ctx context.Context, op string, elapsed time.Duration, opErr error) {
	if m == nil {
		return
	}
	attrs := metric.WithAttributes(
		attribute.String("operation", op),
		attribute.Bool("error", opErr != nil),
	)
	m.storeOps.Add(ctx, 1, attrs)
	m.storeOpDuration.Record(ctx, elapsed.Seconds(), metric.WithAttributes(attribute.String("operation", op)))
}

// recordQueueStats records the DB-derived queue statistics (payload bytes,
// oldest pending age, TTL budget) from an already-observed snapshot. The
// queue depth gauge itself is recorded separately by DurableEmitter from the
// same snapshot's authoritative TotalRows count.
func (m *durableEmitterMetrics) recordQueueStats(ctx context.Context, st DurableQueueStats, maxBytes int64) {
	if m == nil {
		return
	}
	m.queuePayloadBytes.Record(ctx, st.PayloadBytes)
	if st.Depth == 0 {
		m.queueOldestAgeSec.Record(ctx, 0)
	} else {
		m.queueOldestAgeSec.Record(ctx, st.OldestPendingAge.Seconds())
	}
	m.queueTTLBudgetSec.Record(ctx, int64(st.TTLBudget/time.Second))
	if maxBytes > 0 {
		m.queueCapacityRatio.Record(ctx, float64(st.PayloadBytes)/float64(maxBytes))
	}
}

func (m *durableEmitterMetrics) recordEmitDuration(ctx context.Context, elapsed time.Duration, err error) {
	if m == nil {
		return
	}
	m.emitDuration.Record(ctx, elapsed.Seconds(),
		metric.WithAttributes(attribute.Bool("error", err != nil)),
	)
}

func (m *durableEmitterMetrics) recordPublish(ctx context.Context, elapsed time.Duration, phase publishPhase, err error) {
	if m == nil {
		return
	}
	m.publishDuration.Record(ctx, elapsed.Seconds(),
		metric.WithAttributes(
			attribute.String("phase", phase.String()),
			attribute.Bool("error", err != nil),
			attribute.String("client_name", m.clientName),
		),
	)
}

func (m *durableEmitterMetrics) recordPublishBatchEvent(ctx context.Context, phase publishPhase, err error) {
	if m == nil {
		return
	}
	attrs := metric.WithAttributes(
		attribute.String("phase", phase.String()),
		attribute.String("client_name", m.clientName),
	)
	if err != nil {
		m.publishBatchEvErr.Add(ctx, 1, attrs)
	} else {
		m.publishBatchEvOK.Add(ctx, 1, attrs)
	}
}
