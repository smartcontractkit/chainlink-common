package capabilities

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

type BaseTriggerBeholderMetrics struct {
	capabilityID              string
	retryCount                metric.Int64Counter
	ackCount                  metric.Int64Counter
	ackErrorCount             metric.Int64Counter
	ackMemoryOutcomeCount     metric.Int64Counter
	inboxMissingCount         metric.Int64Counter
	inboxFullCount            metric.Int64Counter
	timeToAckMs               metric.Int64Histogram
	ackAttempts               metric.Int64Histogram // attempts distribution at ACK time
	activeRegistrations       metric.Int64UpDownCounter
	pendingEvents             metric.Int64UpDownCounter
	stoppedResendingTimestamp metric.Int64Gauge
	muWaitMs                  metric.Int64Histogram
	scanPendingLockHeldMs     metric.Int64Histogram
	preAckedEntries           metric.Int64Gauge
	storeOpDurationMs         metric.Int64Histogram
}

// contentionBuckets span sub-millisecond fast paths through multi-minute stalls.
// The upper bound must stay well above any observed stall: values past the last
// boundary land in +Inf, where histogram_quantile clamps and silently censors the
// tail (a p95 pinned exactly at the top bucket means "at least this", not "this").
var contentionBuckets = metric.WithExplicitBucketBoundaries(
	1, 5, 10, 25, 50, 100, 250, 500, 1_000, 5_000, 30_000, 120_000,
)

var _ BaseTriggerMetrics = &BaseTriggerBeholderMetrics{}

func NewBaseTriggerBeholderMetrics(capabilityID string) (BaseTriggerMetrics, error) {
	retryCount, err := beholder.GetMeter().Int64Counter("capabilities_base_trigger_retry_total")
	if err != nil {
		return nil, err
	}
	ackCount, err := beholder.GetMeter().Int64Counter("capabilities_base_trigger_ack_total")
	if err != nil {
		return nil, err
	}
	ackErrorCount, err := beholder.GetMeter().Int64Counter("capabilities_base_trigger_ack_error_total")
	if err != nil {
		return nil, err
	}
	ackMemoryOutcomeCount, err := beholder.GetMeter().Int64Counter("capabilities_base_trigger_ack_memory_outcome_total")
	if err != nil {
		return nil, err
	}
	inboxMissingCount, err := beholder.GetMeter().Int64Counter("capabilities_base_trigger_inbox_missing_total")
	if err != nil {
		return nil, err
	}
	inboxFullCount, err := beholder.GetMeter().Int64Counter("capabilities_base_trigger_inbox_full_total")
	if err != nil {
		return nil, err
	}
	timeToAckMs, err := beholder.GetMeter().Int64Histogram("capabilities_base_trigger_time_to_ack_ms",
		metric.WithExplicitBucketBoundaries(100, 500, 1_000, 2_000, 5_000, 10_000, 30_000, 60_000, 120_000, 300_000, 600_000),
	)
	if err != nil {
		return nil, err
	}
	ackAttempts, err := beholder.GetMeter().Int64Histogram("capabilities_base_trigger_ack_attempts",
		metric.WithExplicitBucketBoundaries(1, 2, 3, 5, 10, 15, 20, 25, 50),
	)
	if err != nil {
		return nil, err
	}

	activeRegistrations, err := beholder.GetMeter().Int64UpDownCounter("capabilities_base_trigger_active_registrations")
	if err != nil {
		return nil, err
	}

	pendingEvents, err := beholder.GetMeter().Int64UpDownCounter("capabilities_base_trigger_pending_events")
	if err != nil {
		return nil, err
	}

	stoppedResendingTimestamp, err := beholder.GetMeter().Int64Gauge(
		"capabilities_base_trigger_stopped_resending_timestamp",
	)
	if err != nil {
		return nil, err
	}

	muWaitMs, err := beholder.GetMeter().Int64Histogram("capabilities_base_trigger_mu_wait_ms", contentionBuckets)
	if err != nil {
		return nil, err
	}

	scanPendingLockHeldMs, err := beholder.GetMeter().Int64Histogram("capabilities_base_trigger_scan_pending_lock_held_ms", contentionBuckets)
	if err != nil {
		return nil, err
	}

	preAckedEntries, err := beholder.GetMeter().Int64Gauge("capabilities_base_trigger_preacked_entries")
	if err != nil {
		return nil, err
	}

	storeOpDurationMs, err := beholder.GetMeter().Int64Histogram("capabilities_base_trigger_store_op_duration_ms", contentionBuckets)
	if err != nil {
		return nil, err
	}

	return &BaseTriggerBeholderMetrics{
		capabilityID:              capabilityID,
		retryCount:                retryCount,
		ackCount:                  ackCount,
		ackErrorCount:             ackErrorCount,
		ackMemoryOutcomeCount:     ackMemoryOutcomeCount,
		inboxMissingCount:         inboxMissingCount,
		inboxFullCount:            inboxFullCount,
		timeToAckMs:               timeToAckMs,
		ackAttempts:               ackAttempts,
		activeRegistrations:       activeRegistrations,
		pendingEvents:             pendingEvents,
		stoppedResendingTimestamp: stoppedResendingTimestamp,
		muWaitMs:                  muWaitMs,
		scanPendingLockHeldMs:     scanPendingLockHeldMs,
		preAckedEntries:           preAckedEntries,
		storeOpDurationMs:         storeOpDurationMs,
	}, nil
}

func (m *BaseTriggerBeholderMetrics) attrs(triggerID string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("capability_id", m.capabilityID),
		attribute.String("trigger_id", triggerID),
	}
}

func (m *BaseTriggerBeholderMetrics) IncActiveTriggers() {
	m.activeRegistrations.Add(context.Background(), 1,
		metric.WithAttributes(attribute.String("capability_id", m.capabilityID)),
	)
}

func (m *BaseTriggerBeholderMetrics) DecActiveTriggers() {
	m.activeRegistrations.Add(context.Background(), -1,
		metric.WithAttributes(attribute.String("capability_id", m.capabilityID)),
	)
}

func (m *BaseTriggerBeholderMetrics) IncRetry(triggerID string) {
	m.retryCount.Add(context.Background(), 1,
		metric.WithAttributes(m.attrs(triggerID)...),
	)
}

func (m *BaseTriggerBeholderMetrics) IncAck(triggerID string) {
	m.ackCount.Add(context.Background(), 1,
		metric.WithAttributes(m.attrs(triggerID)...),
	)
}

func (m *BaseTriggerBeholderMetrics) IncAckError(reason string) {
	m.ackErrorCount.Add(context.Background(), 1,
		metric.WithAttributes(
			attribute.String("capability_id", m.capabilityID),
			attribute.String("reason", reason),
		),
	)
}

func (m *BaseTriggerBeholderMetrics) IncAckMemoryOutcome(outcome string) {
	m.ackMemoryOutcomeCount.Add(context.Background(), 1,
		metric.WithAttributes(
			attribute.String("capability_id", m.capabilityID),
			attribute.String("outcome", outcome),
		),
	)
}

func (m *BaseTriggerBeholderMetrics) ObserveTimeToAck(triggerID string, d time.Duration, attempts int) {
	m.timeToAckMs.Record(context.Background(), d.Milliseconds(),
		metric.WithAttributes(m.attrs(triggerID)...),
	)
	m.ackAttempts.Record(context.Background(), int64(attempts),
		metric.WithAttributes(m.attrs(triggerID)...),
	)
}

func (m *BaseTriggerBeholderMetrics) IncInboxMissing(triggerID string) {
	m.inboxMissingCount.Add(context.Background(), 1,
		metric.WithAttributes(
			attribute.String("capability_id", m.capabilityID),
			attribute.String("trigger_id", triggerID),
		),
	)
}

func (m *BaseTriggerBeholderMetrics) IncInboxFull(triggerID string) {
	m.inboxFullCount.Add(context.Background(), 1,
		metric.WithAttributes(
			attribute.String("capability_id", m.capabilityID),
			attribute.String("trigger_id", triggerID),
		),
	)
}

func (m *BaseTriggerBeholderMetrics) AddPendingEvents(delta int64) {
	m.pendingEvents.Add(context.Background(), delta,
		metric.WithAttributes(attribute.String("capability_id", m.capabilityID)),
	)
}

func (m *BaseTriggerBeholderMetrics) IncStoppedResending(triggerID string, attempts int) {
	m.stoppedResendingTimestamp.Record(context.Background(), time.Now().Unix(),
		metric.WithAttributes(
			attribute.String("capability_id", m.capabilityID),
			attribute.String("trigger_id", triggerID),
			attribute.Int("attempts", attempts),
		),
	)
}

func (m *BaseTriggerBeholderMetrics) ObserveMuWait(op string, d time.Duration) {
	m.muWaitMs.Record(context.Background(), d.Milliseconds(),
		metric.WithAttributes(
			attribute.String("capability_id", m.capabilityID),
			attribute.String("op", op),
		),
	)
}

func (m *BaseTriggerBeholderMetrics) ObserveScanPendingLockHeld(d time.Duration) {
	m.scanPendingLockHeldMs.Record(context.Background(), d.Milliseconds(),
		metric.WithAttributes(attribute.String("capability_id", m.capabilityID)),
	)
}

func (m *BaseTriggerBeholderMetrics) SetPreAckedEntries(n int64) {
	m.preAckedEntries.Record(context.Background(), n,
		metric.WithAttributes(attribute.String("capability_id", m.capabilityID)),
	)
}

func (m *BaseTriggerBeholderMetrics) ObserveStoreOp(op string, d time.Duration, outcome string) {
	m.storeOpDurationMs.Record(context.Background(), d.Milliseconds(),
		metric.WithAttributes(
			attribute.String("capability_id", m.capabilityID),
			attribute.String("op", op),
			attribute.String("outcome", outcome),
		),
	)
}

type noopBaseTriggerMetrics struct{}

var _ BaseTriggerMetrics = &noopBaseTriggerMetrics{}

func (noopBaseTriggerMetrics) IncActiveTriggers()                           {}
func (noopBaseTriggerMetrics) DecActiveTriggers()                           {}
func (noopBaseTriggerMetrics) IncRetry(string)                              {}
func (noopBaseTriggerMetrics) IncAck(string)                                {}
func (noopBaseTriggerMetrics) ObserveTimeToAck(string, time.Duration, int)  {}
func (noopBaseTriggerMetrics) IncInboxMissing(string)                       {}
func (noopBaseTriggerMetrics) IncInboxFull(string)                          {}
func (noopBaseTriggerMetrics) IncAckError(string)                           {}
func (noopBaseTriggerMetrics) IncAckMemoryOutcome(string)                   {}
func (noopBaseTriggerMetrics) AddPendingEvents(int64)                       {}
func (noopBaseTriggerMetrics) IncStoppedResending(string, int)              {}
func (noopBaseTriggerMetrics) ObserveMuWait(string, time.Duration)          {}
func (noopBaseTriggerMetrics) ObserveScanPendingLockHeld(time.Duration)     {}
func (noopBaseTriggerMetrics) SetPreAckedEntries(int64)                     {}
func (noopBaseTriggerMetrics) ObserveStoreOp(string, time.Duration, string) {}
