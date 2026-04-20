package beholder

import (
	"context"
	"runtime"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

// DurableEmitterMetricsConfig enables OpenTelemetry metrics for DurableEmitter.
// Set on DurableEmitterConfig.Metrics; nil disables instrumentation.
//
// Instruments are registered on beholder.GetMeter() (same path as capabilities
// and monitoring metrics). Ensure beholder.SetClient has been called with a
// configured client before NewDurableEmitter when metrics are enabled.
type DurableEmitterMetricsConfig struct {
	// PollInterval is how often queue and optional process gauges refresh. Zero = 10s.
	PollInterval time.Duration
	// NearExpiryLead is the window before EventTTL used for queue.near_ttl (DLQ pressure proxy). Zero = 5m.
	NearExpiryLead time.Duration
	// MaxQueuePayloadBytes, if > 0, records capacity_usage_ratio = queue_payload_bytes / max.
	MaxQueuePayloadBytes int64
	// RecordProcessStats records Go heap gauges and, on Unix, cumulative CPU seconds (getrusage).
	RecordProcessStats bool
}

type durableEmitterMetrics struct {
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
	queueDepthMax      metric.Int64Gauge
	queuePayloadBytes  metric.Int64Gauge
	queueOldestAgeSec  metric.Float64Gauge
	queueNearTTL       metric.Int64Gauge
	queueCapacityRatio metric.Float64Gauge
	procHeapInuse      metric.Int64Gauge
	procHeapSys        metric.Int64Gauge
	procCPUUser        metric.Float64Gauge
	procCPUSys         metric.Float64Gauge
}

// durationBuckets provides histogram boundaries (in seconds) tuned for
// sub-millisecond through multi-second latencies. The OTel SDK defaults are
// designed for millisecond-scale integer values and produce wildly wrong
// quantile estimates when values are recorded in fractional seconds.
var durationBuckets = metric.WithExplicitBucketBoundaries(
	0.0001, 0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05,
	0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0,
)

func newDurableEmitterMetrics() (*durableEmitterMetrics, error) {
	meter := GetMeter()
	m := &durableEmitterMetrics{}
	var err error
	if m.emitSuccess, err = durableEmitterMetricEmitSuccess.NewInt64Counter(meter); err != nil {
		return nil, err
	}
	if m.emitFail, err = durableEmitterMetricEmitFailure.NewInt64Counter(meter); err != nil {
		return nil, err
	}
	if m.emitDuration, err = meter.Float64Histogram(
		durableEmitterMetricEmitDuration.Name,
		metric.WithUnit(durableEmitterMetricEmitDuration.Unit),
		metric.WithDescription(durableEmitterMetricEmitDuration.Description),
		durationBuckets,
	); err != nil {
		return nil, err
	}
	if m.emitTotalDuration, err = meter.Float64Histogram(
		durableEmitterMetricEmitTotalDuration.Name,
		metric.WithUnit(durableEmitterMetricEmitTotalDuration.Unit),
		metric.WithDescription(durableEmitterMetricEmitTotalDuration.Description),
		durationBuckets,
	); err != nil {
		return nil, err
	}
	if m.publishImmOK, err = durableEmitterMetricPublishImmSuccess.NewInt64Counter(meter); err != nil {
		return nil, err
	}
	if m.publishImmErr, err = durableEmitterMetricPublishImmFailure.NewInt64Counter(meter); err != nil {
		return nil, err
	}
	if m.publishDuration, err = meter.Float64Histogram(
		durableEmitterMetricPublishDuration.Name,
		metric.WithUnit(durableEmitterMetricPublishDuration.Unit),
		metric.WithDescription(durableEmitterMetricPublishDuration.Description),
		durationBuckets,
	); err != nil {
		return nil, err
	}
	if m.publishBatchOK, err = durableEmitterMetricPublishBatchSuccess.NewInt64Counter(meter); err != nil {
		return nil, err
	}
	if m.publishBatchErr, err = durableEmitterMetricPublishBatchFailure.NewInt64Counter(meter); err != nil {
		return nil, err
	}
	if m.publishBatchEvOK, err = durableEmitterMetricPublishBatchEvSuccess.NewInt64Counter(meter); err != nil {
		return nil, err
	}
	if m.publishBatchEvErr, err = durableEmitterMetricPublishBatchEvFailure.NewInt64Counter(meter); err != nil {
		return nil, err
	}
	if m.deliverComplete, err = durableEmitterMetricDeliveryCompleted.NewInt64Counter(meter); err != nil {
		return nil, err
	}
	if m.expiredPurged, err = durableEmitterMetricExpiredPurged.NewInt64Counter(meter); err != nil {
		return nil, err
	}
	if m.storeOps, err = durableEmitterMetricStoreOperations.NewInt64Counter(meter); err != nil {
		return nil, err
	}
	if m.storeOpDuration, err = meter.Float64Histogram(
		durableEmitterMetricStoreOpDuration.Name,
		metric.WithUnit(durableEmitterMetricStoreOpDuration.Unit),
		metric.WithDescription(durableEmitterMetricStoreOpDuration.Description),
		durationBuckets,
	); err != nil {
		return nil, err
	}
	if m.queueDepth, err = durableEmitterMetricQueueDepth.NewInt64Gauge(meter); err != nil {
		return nil, err
	}
	if m.queueDepthMax, err = durableEmitterMetricQueueDepthMax.NewInt64Gauge(meter); err != nil {
		return nil, err
	}
	if m.queuePayloadBytes, err = durableEmitterMetricQueuePayloadBytes.NewInt64Gauge(meter); err != nil {
		return nil, err
	}
	if m.queueOldestAgeSec, err = durableEmitterMetricQueueOldestAgeSec.NewFloat64Gauge(meter); err != nil {
		return nil, err
	}
	if m.queueNearTTL, err = durableEmitterMetricQueueNearTTL.NewInt64Gauge(meter); err != nil {
		return nil, err
	}
	if m.queueCapacityRatio, err = durableEmitterMetricQueueCapacityRatio.NewFloat64Gauge(meter); err != nil {
		return nil, err
	}
	if m.procHeapInuse, err = durableEmitterMetricProcHeapInuse.NewInt64Gauge(meter); err != nil {
		return nil, err
	}
	if m.procHeapSys, err = durableEmitterMetricProcHeapSys.NewInt64Gauge(meter); err != nil {
		return nil, err
	}
	if m.procCPUUser, err = durableEmitterMetricProcCPUUser.NewFloat64Gauge(meter); err != nil {
		return nil, err
	}
	if m.procCPUSys, err = durableEmitterMetricProcCPUSys.NewFloat64Gauge(meter); err != nil {
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

// pollQueueGauges refreshes DB-derived queue statistics (payload bytes, oldest
// pending age, near-TTL count). Queue depth itself is tracked atomically by
// DurableEmitter.incPending/decPending and recorded there.
func (m *durableEmitterMetrics) pollQueueGauges(ctx context.Context, obs DurableQueueObserver, ttl, lead time.Duration, maxBytes int64) {
	if m == nil || obs == nil {
		return
	}
	st, err := obs.ObserveDurableQueue(ctx, ttl, lead)
	if err != nil {
		return
	}
	m.queuePayloadBytes.Record(ctx, st.PayloadBytes)
	if st.Depth == 0 {
		m.queueOldestAgeSec.Record(ctx, 0)
	} else {
		m.queueOldestAgeSec.Record(ctx, st.OldestPendingAge.Seconds())
	}
	m.queueNearTTL.Record(ctx, st.NearTTLCount)
	if maxBytes > 0 {
		m.queueCapacityRatio.Record(ctx, float64(st.PayloadBytes)/float64(maxBytes))
	}
}

func (m *durableEmitterMetrics) recordProcessMem(ctx context.Context) {
	if m == nil {
		return
	}
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	m.procHeapInuse.Record(ctx, int64(ms.HeapInuse))
	m.procHeapSys.Record(ctx, int64(ms.HeapSys))
}

func (m *durableEmitterMetrics) recordPublish(ctx context.Context, elapsed time.Duration, phase string, err error) {
	if m == nil {
		return
	}
	m.publishDuration.Record(ctx, elapsed.Seconds(),
		metric.WithAttributes(
			attribute.String("phase", phase),
			attribute.Bool("error", err != nil),
		),
	)
}
