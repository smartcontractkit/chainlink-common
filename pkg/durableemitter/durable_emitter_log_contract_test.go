package durableemitter

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
)

// Log-message contract pins.
//
// DO NOT change these strings to make a failing test pass without coordinating
// with the CRE observability team.
//
// WHO DEPENDS ON THESE: the CRE E2E-integrity monitoring system (epic
// CRE-5693). Its integrity detectors alert on these exact DurableEmitter warn
// messages to distinguish "event dropped but will be retransmitted" from a
// genuine delivery gap. If the wording in durable_emitter.go changes, those
// log-based detectors silently stop matching and E2E integrity monitoring
// goes blind. Treat a failure here as a breaking change to an external
// contract: update the detectors in lockstep, or keep the message as-is.
const (
	// Emitted from DurableEmitter.deliveryCallback when a batch publish fails
	// and the event is left in the DB for the retransmit loop.
	logMsgDeliveryFailed = "DurableEmitter: failed to deliver event. Relying on retransmit."
	// Emitted from DurableEmitter.Emit when BatchEmitter.QueueMessage rejects
	// the event (buffer full / stopped) and delivery falls back to retransmit.
	logMsgBufferFull = "DurableEmitter: batch emitter buffer full, relying on retransmit"
)

// TestDurableEmitter_LogContract_DeliveryFailure drives a real publish failure
// through the delivery callback (via testBatchEmitter.setPublishErr) and
// asserts the exact warn message the CRE integrity detectors match on.
func TestDurableEmitter_LogContract_DeliveryFailure(t *testing.T) {
	lggr, observed := logger.TestObserved(t, zapcore.WarnLevel)

	store := NewMemDurableEventStore()
	be := newTestBatchEmitter()
	be.setPublishErr(errors.New("chip ingress unavailable"))

	cfg := DefaultConfig()
	cfg.DisablePruning = true

	// retransmitEnabled=false keeps background loops from adding extra log
	// noise; the delivery callback path under test is unaffected.
	em, err := NewDurableEmitter(store, be, false, cfg, lggr, nil)
	require.NoError(t, err)
	servicetest.Run(t, em)

	require.NoError(t, em.Emit(t.Context(), []byte("log-contract"), testEmitAttrs()...))

	// The delivery callback fires asynchronously; FilterMessage is an exact
	// (full-string) match, so this pins the message verbatim.
	require.Eventually(t, func() bool {
		return observed.FilterMessage(logMsgDeliveryFailed).Len() >= 1
	}, 2*time.Second, 10*time.Millisecond,
		"exact warn message %q not observed: the DurableEmitter delivery-failure log contract changed; this breaks the CRE E2E-integrity detectors (CRE-5693)",
		logMsgDeliveryFailed)

	entry := observed.FilterMessage(logMsgDeliveryFailed).All()[0]
	require.Equal(t, zapcore.WarnLevel, entry.Level,
		"delivery-failure log must stay at Warn level; the CRE integrity detectors filter on it")

	// The event must remain persisted for the retransmit loop — that is the
	// promise the message makes.
	require.Equal(t, 1, store.Len(), "event must remain in the store for retransmit after a delivery failure")
}

// rejectingBatchEmitter is a BatchEmitter whose QueueMessage always rejects,
// simulating a full internal buffer. This forces Emit down the
// "buffer full, relying on retransmit" branch.
type rejectingBatchEmitter struct{}

func (rejectingBatchEmitter) QueueMessage(*chipingress.CloudEventPb, func(error)) error {
	return errors.New("buffer full")
}
func (rejectingBatchEmitter) Start(context.Context) {}
func (rejectingBatchEmitter) Stop()                 {}

// TestDurableEmitter_LogContract_BufferFull forces QueueMessage to reject and
// asserts the exact warn message the CRE integrity detectors match on.
func TestDurableEmitter_LogContract_BufferFull(t *testing.T) {
	lggr, observed := logger.TestObserved(t, zapcore.WarnLevel)

	store := NewMemDurableEventStore()

	cfg := DefaultConfig()
	cfg.DisablePruning = true

	em, err := NewDurableEmitter(store, rejectingBatchEmitter{}, false, cfg, lggr, nil)
	require.NoError(t, err)
	servicetest.Run(t, em)

	// Emit succeeds (the insert is durable) even though enqueueing fails;
	// the warn is logged synchronously before Emit returns.
	require.NoError(t, em.Emit(t.Context(), []byte("log-contract"), testEmitAttrs()...))

	require.Equal(t, 1, observed.FilterMessage(logMsgBufferFull).Len(),
		"exact warn message %q not observed: the DurableEmitter buffer-full log contract changed; this breaks the CRE E2E-integrity detectors (CRE-5693)",
		logMsgBufferFull)

	entry := observed.FilterMessage(logMsgBufferFull).All()[0]
	require.Equal(t, zapcore.WarnLevel, entry.Level,
		"buffer-full log must stay at Warn level; the CRE integrity detectors filter on it")

	require.Equal(t, 1, store.Len(), "event must remain in the store for retransmit when the buffer is full")
}
