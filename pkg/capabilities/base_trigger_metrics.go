package capabilities

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
)

type BaseTriggerBeholderMetrics struct {
	capabilityID             string
	retryCount               metric.Int64Counter
	ackCount                 metric.Int64Counter
	ackErrorCount            metric.Int64Counter
	ackMemoryOutcomeCount    metric.Int64Counter
	inboxMissingCount        metric.Int64Counter
	inboxFullCount           metric.Int64Counter
	undeliveredWarningCount  metric.Int64Counter
	undeliveredCriticalCount metric.Int64Counter
	timeToAckMs              metric.Int64Histogram
	ackAttempts              metric.Int64Histogram // attempts distribution at ACK time
	activeRegistrations      metric.Int64UpDownCounter
	pendingEvents            metric.Int64UpDownCounter
	stuckEvents              metric.Int64UpDownCounter
	gaveUpCount              metric.Int64Counter
}

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
	undeliveredWarningCount, err := beholder.GetMeter().Int64Counter("capabilities_base_trigger_undelivered_total")
	if err != nil {
		return nil, err
	}
	undeliveredCriticalCount, err := beholder.GetMeter().Int64Counter("capabilities_base_trigger_undelivered2_total")
	if err != nil {
		return nil, err
	}

	timeToAckMs, err := beholder.GetMeter().Int64Histogram("capabilities_base_trigger_time_to_ack_ms")
	if err != nil {
		return nil, err
	}
	ackAttempts, err := beholder.GetMeter().Int64Histogram("capabilities_base_trigger_ack_attempts")
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

	stuckEvents, err := beholder.GetMeter().Int64UpDownCounter("capabilities_base_trigger_stuck_events")
	if err != nil {
		return nil, err
	}

	gaveUpCount, err := beholder.GetMeter().Int64Counter("capabilities_base_trigger_gave_up_total")
	if err != nil {
		return nil, err
	}

	return &BaseTriggerBeholderMetrics{
		capabilityID:             capabilityID,
		retryCount:               retryCount,
		ackCount:                 ackCount,
		ackErrorCount:            ackErrorCount,
		ackMemoryOutcomeCount:    ackMemoryOutcomeCount,
		inboxMissingCount:        inboxMissingCount,
		inboxFullCount:           inboxFullCount,
		undeliveredWarningCount:  undeliveredWarningCount,
		undeliveredCriticalCount: undeliveredCriticalCount,
		timeToAckMs:              timeToAckMs,
		ackAttempts:              ackAttempts,
		activeRegistrations:      activeRegistrations,
		pendingEvents:            pendingEvents,
		stuckEvents:              stuckEvents,
		gaveUpCount:              gaveUpCount,
	}, nil
}

func (m *BaseTriggerBeholderMetrics) attrs(triggerID, eventID string) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("capability_id", m.capabilityID),
		attribute.String("trigger_id", triggerID),
		attribute.String("event_id", eventID),
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

func (m *BaseTriggerBeholderMetrics) IncRetry(triggerID, eventID string) {
	m.retryCount.Add(context.Background(), 1,
		metric.WithAttributes(m.attrs(triggerID, eventID)...),
	)
}

func (m *BaseTriggerBeholderMetrics) IncAck(triggerID, eventID string) {
	m.ackCount.Add(context.Background(), 1,
		metric.WithAttributes(m.attrs(triggerID, eventID)...),
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

func (m *BaseTriggerBeholderMetrics) ObserveTimeToAck(triggerID, eventID string, d time.Duration, attempts int) {
	m.timeToAckMs.Record(context.Background(), d.Milliseconds(),
		metric.WithAttributes(m.attrs(triggerID, eventID)...),
	)
	m.ackAttempts.Record(context.Background(), int64(attempts),
		metric.WithAttributes(m.attrs(triggerID, eventID)...),
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

func (m *BaseTriggerBeholderMetrics) EmitUndeliveredWarning(triggerID, eventID string) {
	m.undeliveredWarningCount.Add(context.Background(), 1,
		metric.WithAttributes(m.attrs(triggerID, eventID)...),
	)
}

func (m *BaseTriggerBeholderMetrics) EmitUndeliveredCritical(triggerID, eventID string) {
	m.undeliveredCriticalCount.Add(context.Background(), 1,
		metric.WithAttributes(m.attrs(triggerID, eventID)...),
	)
}

func (m *BaseTriggerBeholderMetrics) AddPendingEvents(delta int64) {
	m.pendingEvents.Add(context.Background(), delta,
		metric.WithAttributes(attribute.String("capability_id", m.capabilityID)),
	)
}

func (m *BaseTriggerBeholderMetrics) IncStuckEvent(triggerID, eventID string) {
	m.stuckEvents.Add(context.Background(), 1,
		metric.WithAttributes(m.attrs(triggerID, eventID)...),
	)
}

func (m *BaseTriggerBeholderMetrics) DecStuckEvent(triggerID, eventID string) {
	m.stuckEvents.Add(context.Background(), -1,
		metric.WithAttributes(m.attrs(triggerID, eventID)...),
	)
}

func (m *BaseTriggerBeholderMetrics) IncGaveUp(triggerID, eventID string, attempts int) {
	m.gaveUpCount.Add(context.Background(), 1,
		metric.WithAttributes(
			attribute.String("capability_id", m.capabilityID),
			attribute.String("trigger_id", triggerID),
			attribute.String("event_id", eventID),
			attribute.Int("attempts", attempts),
		),
	)
}

type noopBaseTriggerMetrics struct{}

var _ BaseTriggerMetrics = &noopBaseTriggerMetrics{}

func (noopBaseTriggerMetrics) IncActiveTriggers()                                  {}
func (noopBaseTriggerMetrics) DecActiveTriggers()                                  {}
func (noopBaseTriggerMetrics) IncRetry(string, string)                             {}
func (noopBaseTriggerMetrics) IncAck(string, string)                               {}
func (noopBaseTriggerMetrics) ObserveTimeToAck(string, string, time.Duration, int) {}
func (noopBaseTriggerMetrics) IncInboxMissing(string)                              {}
func (noopBaseTriggerMetrics) IncInboxFull(string)                                 {}
func (noopBaseTriggerMetrics) EmitUndeliveredWarning(string, string)               {}
func (noopBaseTriggerMetrics) EmitUndeliveredCritical(string, string)              {}
func (noopBaseTriggerMetrics) IncAckError(string)                                  {}
func (noopBaseTriggerMetrics) IncAckMemoryOutcome(string)                          {}
func (noopBaseTriggerMetrics) AddPendingEvents(int64)                              {}
func (noopBaseTriggerMetrics) IncStuckEvent(string, string)                        {}
func (noopBaseTriggerMetrics) DecStuckEvent(string, string)                        {}
func (noopBaseTriggerMetrics) IncGaveUp(string, string, int)                       {}
