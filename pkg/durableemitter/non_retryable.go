package durableemitter

import (
	"strings"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
)

// nonRetryablePublishErrorMarkers lists substrings that identify chip-ingress
// publish failures that will never succeed on retry for a given event. We match
// on the error message rather than a typed error code so this stays decoupled
// from the chip-ingress error API/version — we only inspect the error we
// actually receive on our side.
//
// This is the single extension point for non-retryable handling: add further
// permanently-failing markers here as they are identified.
var nonRetryablePublishErrorMarkers = []string{
	// The event's schema is not registered in the chip-ingress schema registry.
	// Republishing the same event will keep failing until the schema is
	// registered server-side, so there is no point retransmitting it.
	"PUBLISH_ERROR_CODE_SCHEMA_MISSING",
}

// nonRetryablePublishError reports whether err represents a permanent chip
// publish failure that should not be retransmitted, returning the matched
// marker for logging/metrics.
func nonRetryablePublishError(err error) (string, bool) {
	if err == nil {
		return "", false
	}
	msg := err.Error()
	for _, marker := range nonRetryablePublishErrorMarkers {
		if strings.Contains(msg, marker) {
			return marker, true
		}
	}
	return "", false
}

// dropNonRetryable removes an event from persistence after a permanent publish
// failure and logs a warning. The event is intentionally not retransmitted.
// pendingCount is decremented so queue-depth metrics stay accurate.
func (d *DurableEmitter) dropNonRetryable(id int64, eventPb *chipingress.CloudEventPb, reason string) {
	d.eng.Warnw("DurableEmitter: dropping event with non-retryable publish error; will not retransmit",
		"id", id,
		"eventID", eventPb.GetId(),
		"source", eventPb.GetSource(),
		"type", eventPb.GetType(),
		"reason", reason,
	)

	ctx, cancel := d.stopCh.CtxWithTimeout(d.cfg.PublishTimeout)
	defer cancel()

	if err := d.store.Delete(ctx, id); err != nil {
		d.eng.Errorw("DurableEmitter: failed to delete non-retryable event from store", "id", id, "error", err)
		return
	}

	d.decPending(1)
	if d.metrics != nil {
		d.metrics.nonRetryableDropped.Add(ctx, 1)
	}
}
