package capabilities

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

func TestValidateBaseTriggerRetryInterval(t *testing.T) {
	ctx := t.Context()

	t.Run("nil getter errors", func(t *testing.T) {
		err := ValidateBaseTriggerRetryInterval(ctx, nil)
		require.Error(t, err)
	})

	t.Run("positive interval succeeds even when retransmit disabled", func(t *testing.T) {
		getter, err := settings.NewJSONGetter([]byte(`{
			"global": {
				"BaseTriggerRetransmitEnabled": "false",
				"BaseTriggerRetryInterval": "7s"
			}
		}`))
		require.NoError(t, err)
		require.NoError(t, ValidateBaseTriggerRetryInterval(ctx, getter))
	})

	t.Run("zero interval errors", func(t *testing.T) {
		getter, err := settings.NewJSONGetter([]byte(`{
			"global": {
				"BaseTriggerRetransmitEnabled": "true",
				"BaseTriggerRetryInterval": "0s"
			}
		}`))
		require.NoError(t, err)
		err = ValidateBaseTriggerRetryInterval(ctx, getter)
		require.Error(t, err)
		require.Contains(t, err.Error(), "BaseTriggerRetryInterval must be positive")
	})
}

// atomicJSONGetter swaps a whole settings.Getter under a mutex for dynamic-settings tests.
type atomicJSONGetter struct {
	mu sync.Mutex
	g  settings.Getter
}

func (f *atomicJSONGetter) setJSON(js string) error {
	g, err := settings.NewJSONGetter([]byte(js))
	if err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.g = g
	return nil
}

func (f *atomicJSONGetter) GetScoped(ctx context.Context, scope settings.Scope, key string) (string, error) {
	f.mu.Lock()
	g := f.g
	f.mu.Unlock()
	if g == nil {
		return "", errors.New("atomicJSONGetter: no JSON set")
	}
	return g.GetScoped(ctx, scope, key)
}

func TestBaseTrigger_CRE_DynamicDisableStopsResend(t *testing.T) {
	lggr, err := logger.New()
	require.NoError(t, err)
	ctx := t.Context()

	getter := &atomicJSONGetter{}
	require.NoError(t, getter.setJSON(`{
		"global": {
			"BaseTriggerRetransmitEnabled": "true",
			"BaseTriggerRetryInterval": "20ms"
		}
	}`))

	store := NewMemEventStore()
	b, err := NewBaseTriggerCapabilityWithCRESettings(ctx, store,
		func() *wrapperspb.BytesValue { return &wrapperspb.BytesValue{} },
		lggr, "testCap", getter)
	require.NoError(t, err)

	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 10)
	b.RegisterTrigger("trig", sendCh)
	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() { b.Stop() })

	te := makeTE(t, "trig", "e1", []byte("payload"))
	require.NoError(t, b.DeliverEvent(ctx, te, "trig"))

	<-sendCh

	require.NoError(t, getter.setJSON(`{
		"global": {
			"BaseTriggerRetransmitEnabled": "false",
			"BaseTriggerRetryInterval": "20ms"
		}
	}`))

	select {
	case extra := <-sendCh:
		t.Fatalf("unexpected send after disable: %+v", extra)
	case <-time.After(3 * time.Second):
	}
}

func newBase(t *testing.T, store EventStore) *BaseTriggerCapability[*wrapperspb.BytesValue] {
	return newBaseWithRetransmit(t, store, 100*time.Millisecond)
}

func newBaseWithRetransmit(t *testing.T, store EventStore, tRetransmit time.Duration) *BaseTriggerCapability[*wrapperspb.BytesValue] {
	lggr, err := logger.New()
	require.NoError(t, err)
	return NewBaseTriggerCapability(store, func() *wrapperspb.BytesValue { return &wrapperspb.BytesValue{} }, lggr,
		"testCap", tRetransmit, nil)
}

func ctxWithCancel(t *testing.T) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(t.Context())
	return ctx, cancel
}

func makeTE(t *testing.T, trigger, id string, b []byte) TriggerEvent {
	t.Helper()
	m := &wrapperspb.BytesValue{Value: b}
	a, err := anypb.New(m)
	require.NoError(t, err)

	return TriggerEvent{
		TriggerType: trigger,
		ID:          id,
		Payload:     a,
	}
}

func TestStart_LoadsAndSendsPersisted(t *testing.T) {
	store := NewMemEventStore()
	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 10)

	// Preload store with one record
	msg := &wrapperspb.BytesValue{Value: []byte("payload")}
	anyMsg, err := anypb.New(msg)
	require.NoError(t, err)

	rec := PendingEvent{
		TriggerId:  "trigA",
		EventId:    "e1",
		AnyTypeURL: anyMsg.TypeUrl,
		Payload:    anyMsg.Value,
		FirstAt:    time.Now().Add(-1 * time.Minute),
	}
	require.NoError(t, store.Insert(context.Background(), rec))

	b := newBase(t, store)

	b.RegisterTrigger("trigA", sendCh)

	require.NoError(t, b.Start(t.Context()))
	t.Cleanup(func() {
		b.Stop()
		b.UnregisterTrigger("trigA")
	})

	// Initial send triggered on Start
	require.Eventually(t, func() bool {
		select {
		case <-sendCh:
			return true
		default:
			return false
		}
	}, 200*time.Millisecond, 5*time.Millisecond)
}

func TestDeliverEvent_PersistsAndSends(t *testing.T) {
	store := NewMemEventStore()
	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 10)

	b := newBase(t, store)
	ctx, cancel := ctxWithCancel(t)
	defer cancel()

	b.RegisterTrigger("trigA", sendCh)

	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() {
		b.Stop()
		b.UnregisterTrigger("trigA")
	})

	te := makeTE(t, "trigA", "e2", []byte("x"))
	require.NoError(t, b.DeliverEvent(ctx, te, "trigA"))

	recs, _ := store.List(ctx)
	require.Len(t, recs, 1)

	resendCount := 0
	require.Eventually(t, func() bool {
		select {
		case <-sendCh:
			resendCount++
		default:
			break
		}
		return resendCount >= 3
	}, 10*time.Second, 5*time.Millisecond)
}

func TestAckEvent_StopsRetransmit(t *testing.T) {
	store := NewMemEventStore()
	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 10)

	b := newBase(t, store)
	ctx, cancel := ctxWithCancel(t)
	defer cancel()

	b.RegisterTrigger("trigC", sendCh)

	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() {
		b.Stop()
		b.UnregisterTrigger("trigC")
	})

	te := makeTE(t, "trigC", "e3", []byte("x"))
	require.NoError(t, b.DeliverEvent(ctx, te, "trigC"))

	// Wait for at least one send
	require.Eventually(t, func() bool {
		select {
		case <-sendCh:
			return true
		default:
			return false
		}
	}, 300*time.Millisecond, 5*time.Millisecond)

	// call AckEvent to stop retransmitting
	require.NoError(t, b.AckEvent(ctx, "trigC", "e3"))

	// Drain anything already buffered (could have raced with ack)
drain:
	for {
		select {
		case <-sendCh:
		default:
			break drain
		}
	}

	// Now ensure nothing more is sent after a few retransmit periods
	time.Sleep(3 * b.tRetransmit)
	select {
	case got := <-sendCh:
		t.Fatalf("unexpected retransmit after ACK: %+v", got)
	default:
	}
}

func TestRetransmitDisabled_DeliversOnceWithoutPersistence(t *testing.T) {
	store := NewMemEventStore()
	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 10)

	b := newBaseWithRetransmit(t, store, 0)
	ctx := t.Context()

	b.RegisterTrigger("trigA", sendCh)

	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() {
		b.Stop()
		b.UnregisterTrigger("trigA")
	})

	te := makeTE(t, "trigA", "e1", []byte("payload"))
	require.NoError(t, b.DeliverEvent(ctx, te, "trigA"))

	// Should receive the event once
	select {
	case got := <-sendCh:
		require.Equal(t, "e1", got.Id)
	case <-time.After(time.Second):
		t.Fatal("expected event delivery")
	}

	// Store should be empty (no persistence)
	recs, err := store.List(ctx)
	require.NoError(t, err)
	require.Empty(t, recs)

	// Wait a bit and confirm no retransmits
	time.Sleep(200 * time.Millisecond)
	select {
	case got := <-sendCh:
		t.Fatalf("unexpected retransmit: %+v", got)
	default:
	}
}

func TestRetransmitDisabled_AckReconcilesStore(t *testing.T) {
	store := NewMemEventStore()
	b := newBaseWithRetransmit(t, store, 0)

	require.NoError(t, b.Start(t.Context()))
	t.Cleanup(func() { b.Stop() })

	require.NoError(t, b.AckEvent(t.Context(), "anyTrigger", "anyEvent"))
}

func TestBaseTrigger_MaxRetries_GivesUp(t *testing.T) {
	lggr, err := logger.New()
	require.NoError(t, err)
	ctx := t.Context()

	getter := &atomicJSONGetter{}
	require.NoError(t, getter.setJSON(`{
		"global": {
			"BaseTriggerRetransmitEnabled": "true",
			"BaseTriggerRetryInterval": "20ms",
			"BaseTriggerMaxRetries": "3"
		}
	}`))

	store := NewMemEventStore()
	b, err := NewBaseTriggerCapabilityWithCRESettings(ctx, store,
		func() *wrapperspb.BytesValue { return &wrapperspb.BytesValue{} },
		lggr, "testCap", getter)
	require.NoError(t, err)

	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 50)
	b.RegisterTrigger("trig", sendCh)
	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() { b.Stop() })

	te := makeTE(t, "trig", "e1", []byte("payload"))
	require.NoError(t, b.DeliverEvent(ctx, te, "trig"))

	// After max retries the event should be removed from in-memory pending.
	// The store row is intentionally left for the prune loop to clean up later,
	// giving operators time to investigate unrecoverable payloads.
	require.Eventually(t, func() bool {
		b.mu.Lock()
		defer b.mu.Unlock()
		reg := b.byTrigger["trig"]
		if reg == nil {
			return true
		}
		_, still := reg.pending["e1"]
		return !still
	}, 5*time.Second, 10*time.Millisecond, "event should be removed from pending after max retries")

	// Store should still contain the row (prune loop handles cleanup).
	recs, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, recs, 1, "store row should remain for prune loop")
}

func TestBaseTrigger_MaxRetries_AckBeforeLimit(t *testing.T) {
	lggr, err := logger.New()
	require.NoError(t, err)
	ctx := t.Context()

	getter := &atomicJSONGetter{}
	require.NoError(t, getter.setJSON(`{
		"global": {
			"BaseTriggerRetransmitEnabled": "true",
			"BaseTriggerRetryInterval": "20ms",
			"BaseTriggerMaxRetries": "100"
		}
	}`))

	store := NewMemEventStore()
	b, err := NewBaseTriggerCapabilityWithCRESettings(ctx, store,
		func() *wrapperspb.BytesValue { return &wrapperspb.BytesValue{} },
		lggr, "testCap", getter)
	require.NoError(t, err)

	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 50)
	b.RegisterTrigger("trig", sendCh)
	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() { b.Stop() })

	te := makeTE(t, "trig", "e1", []byte("payload"))
	require.NoError(t, b.DeliverEvent(ctx, te, "trig"))

	// Wait for at least one send then ACK.
	<-sendCh
	require.NoError(t, b.AckEvent(ctx, "trig", "e1"))

	// Verify cleared.
	b.mu.Lock()
	reg := b.byTrigger["trig"]
	pendingEmpty := reg == nil || len(reg.pending) == 0
	b.mu.Unlock()
	require.True(t, pendingEmpty)
}

func TestBaseTrigger_PreAck_DeliverAfterAck(t *testing.T) {
	store := NewMemEventStore()
	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 10)

	b := newBase(t, store)
	ctx, cancel := ctxWithCancel(t)
	defer cancel()

	b.RegisterTrigger("trigA", sendCh)

	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() {
		b.Stop()
		b.UnregisterTrigger("trigA")
	})

	// ACK arrives BEFORE the event is delivered (pre-ACK).
	// This simulates the scenario where other capability DON nodes delivered
	// the event first, the workflow engine executed and ACKed, and the ACK
	// reaches this node before its own DeliverEvent is called.
	require.NoError(t, b.AckEvent(ctx, "trigA", "e1"))

	// Now deliver the event — it should be silently skipped.
	te := makeTE(t, "trigA", "e1", []byte("payload"))
	require.NoError(t, b.DeliverEvent(ctx, te, "trigA"))

	// Event should NOT be in the pending map.
	b.mu.Lock()
	reg := b.byTrigger["trigA"]
	_, stillPending := reg.pending["e1"]
	b.mu.Unlock()
	require.False(t, stillPending, "event should not be pending after pre-ACK")

	// Event should NOT be in the store.
	recs, err := store.List(ctx)
	require.NoError(t, err)
	require.Empty(t, recs, "event should not be persisted after pre-ACK")

	// No retransmissions should occur.
	time.Sleep(3 * b.tRetransmit)
	select {
	case got := <-sendCh:
		t.Fatalf("unexpected send after pre-ACKed delivery: %+v", got)
	default:
	}
}

func TestBaseTrigger_PreAck_SecondEventStillDelivers(t *testing.T) {
	store := NewMemEventStore()
	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 10)

	b := newBase(t, store)
	ctx, cancel := ctxWithCancel(t)
	defer cancel()

	b.RegisterTrigger("trigA", sendCh)

	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() {
		b.Stop()
		b.UnregisterTrigger("trigA")
	})

	// Pre-ACK only event e1.
	require.NoError(t, b.AckEvent(ctx, "trigA", "e1"))

	// Deliver e1 (should be skipped) and e2 (should go through normally).
	te1 := makeTE(t, "trigA", "e1", []byte("payload1"))
	require.NoError(t, b.DeliverEvent(ctx, te1, "trigA"))

	te2 := makeTE(t, "trigA", "e2", []byte("payload2"))
	require.NoError(t, b.DeliverEvent(ctx, te2, "trigA"))

	// e2 should be in the store.
	recs, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, recs, 1)
	require.Equal(t, "e2", recs[0].EventId)

	// e2 should be received on the channel.
	select {
	case got := <-sendCh:
		require.Equal(t, "e2", got.Id)
	case <-time.After(time.Second):
		t.Fatal("expected e2 delivery")
	}
}

func TestBaseTrigger_PreAck_CleanedUpByUnregister(t *testing.T) {
	store := NewMemEventStore()
	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 10)

	b := newBase(t, store)
	ctx, cancel := ctxWithCancel(t)
	defer cancel()

	b.RegisterTrigger("trigA", sendCh)
	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() { b.Stop() })

	// Pre-ACK e1.
	require.NoError(t, b.AckEvent(ctx, "trigA", "e1"))

	b.mu.Lock()
	require.Contains(t, b.byTrigger["trigA"].preAcked, "e1")
	b.mu.Unlock()

	// preAcked entries have a 24h TTL and should persist across scanPending cycles.
	time.Sleep(50 * time.Millisecond)
	b.mu.Lock()
	require.Contains(t, b.byTrigger["trigA"].preAcked, "e1", "preAcked entry should persist (24h TTL)")
	b.mu.Unlock()

	// UnregisterTrigger should clean up preAcked entries for that trigger.
	b.UnregisterTrigger("trigA")

	b.mu.Lock()
	_, exists := b.byTrigger["trigA"]
	b.mu.Unlock()
	require.False(t, exists, "preAcked entries should be cleaned up after unregister")
}

func TestBaseTrigger_PreAck_DoubleCheckCatchesRace(t *testing.T) {
	// Simulates the exact race from production: DeliverEvent's first preAcked
	// check passes (no pre-ACK yet), then during store.Insert an ACK arrives.
	// The double-check after Insert must catch it.
	insertStarted := make(chan struct{})
	proceedInsert := make(chan struct{})
	store := &blockingInsertStore{
		MemEventStore:  NewMemEventStore(),
		insertStarted:  insertStarted,
		proceedInsert:  proceedInsert,
		blockNextCount: 1,
	}

	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 10)
	b := newBase(t, store)
	ctx, cancel := ctxWithCancel(t)
	defer cancel()

	b.RegisterTrigger("trigA", sendCh)
	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() { b.Stop() })

	te := makeTE(t, "trigA", "e1", []byte("payload"))

	// Start DeliverEvent in background — it will block during store.Insert.
	deliverDone := make(chan error, 1)
	go func() {
		deliverDone <- b.DeliverEvent(ctx, te, "trigA")
	}()

	// Wait for Insert to start (first preAcked check has already passed).
	<-insertStarted

	// Simulate the ACK arriving while Insert is blocked.
	require.NoError(t, b.AckEvent(ctx, "trigA", "e1"))

	// Unblock the Insert.
	close(proceedInsert)

	// DeliverEvent should succeed (no error) but skip adding to pending.
	require.NoError(t, <-deliverDone)

	// Event should NOT be in pending (double-check removed it).
	b.mu.Lock()
	reg := b.byTrigger["trigA"]
	_, stillPending := reg.pending["e1"]
	b.mu.Unlock()
	require.False(t, stillPending, "event should not be pending after double-check caught pre-ACK")

	// No retransmissions should occur.
	time.Sleep(3 * b.tRetransmit)
	select {
	case got := <-sendCh:
		t.Fatalf("unexpected send after double-check pre-ACK: %+v", got)
	default:
	}
}

// blockingInsertStore wraps MemEventStore and blocks during Insert to simulate
// slow database writes, allowing ACKs to arrive during the write.
type blockingInsertStore struct {
	*MemEventStore
	insertStarted  chan struct{}
	proceedInsert  chan struct{}
	mu2            sync.Mutex
	blockNextCount int
}

func (s *blockingInsertStore) Insert(ctx context.Context, r PendingEvent) error {
	s.mu2.Lock()
	shouldBlock := s.blockNextCount > 0
	if shouldBlock {
		s.blockNextCount--
	}
	s.mu2.Unlock()

	if shouldBlock {
		s.insertStarted <- struct{}{}
		<-s.proceedInsert
	}
	return s.MemEventStore.Insert(ctx, r)
}

func TestBaseTrigger_PreAck_UnregisterClearsCache(t *testing.T) {
	store := NewMemEventStore()
	b := newBase(t, store)

	require.NoError(t, b.Start(t.Context()))
	t.Cleanup(func() { b.Stop() })

	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 10)
	b.RegisterTrigger("trigA", sendCh)

	require.NoError(t, b.AckEvent(t.Context(), "trigA", "e1"))

	b.mu.Lock()
	_, hasReg := b.byTrigger["trigA"]
	b.mu.Unlock()
	require.True(t, hasReg)

	b.UnregisterTrigger("trigA")

	b.mu.Lock()
	_, exists := b.byTrigger["trigA"]
	b.mu.Unlock()
	require.False(t, exists, "preAcked should be cleared on unregister")
}

func TestBaseTrigger_RedeliveryAfterAck_Skipped(t *testing.T) {
	store := NewMemEventStore()
	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 50)

	b := newBase(t, store)
	ctx, cancel := ctxWithCancel(t)
	defer cancel()

	b.RegisterTrigger("trigA", sendCh)

	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() {
		b.Stop()
		b.UnregisterTrigger("trigA")
	})

	// Normal flow: deliver → send → ACK.
	te := makeTE(t, "trigA", "e1", []byte("payload"))
	require.NoError(t, b.DeliverEvent(ctx, te, "trigA"))

	<-sendCh
	require.NoError(t, b.AckEvent(ctx, "trigA", "e1"))

	// After a successful ACK, the event should be recorded in preAcked
	// so that re-deliveries (e.g. from EVM trigger after block finalization)
	// are skipped.
	b.mu.Lock()
	reg := b.byTrigger["trigA"]
	_, inPreAcked := reg.preAcked["e1"]
	_, inPending := reg.pending["e1"]
	b.mu.Unlock()
	require.True(t, inPreAcked, "ACKed event should be in preAcked cache")
	require.False(t, inPending, "ACKed event should not be in pending")

	// Re-deliver the same event (simulates EVM trigger re-delivery after
	// block finalization prunes its unfinalizedSentEventIDs).
	te2 := makeTE(t, "trigA", "e1", []byte("payload"))
	require.NoError(t, b.DeliverEvent(ctx, te2, "trigA"))

	// Should NOT be in pending or store.
	b.mu.Lock()
	reg2 := b.byTrigger["trigA"]
	_, stillPending := reg2.pending["e1"]
	b.mu.Unlock()
	require.False(t, stillPending, "re-delivered event should not be pending")

	recs, err := store.List(ctx)
	require.NoError(t, err)
	require.Empty(t, recs, "re-delivered event should not be persisted")
}

func TestRetryBackoff(t *testing.T) {
	base := 30 * time.Second

	tests := []struct {
		attempts   int
		wantExact  time.Duration
		wantCapped bool
	}{
		{0, base, false},
		{1, 2 * base, false},
		{2, 4 * base, false},
		{3, 8 * base, false},
		{4, time.Duration(backoffMultiplierCap) * base, true},
		{5, time.Duration(backoffMultiplierCap) * base, true},
		{20, time.Duration(backoffMultiplierCap) * base, true},
	}

	for _, tc := range tests {
		got := retryBackoff(base, tc.attempts)
		require.Equal(t, tc.wantExact, got, "attempts=%d", tc.attempts)
	}
}

func TestAddJitter_BoundsAndVariation(t *testing.T) {
	base := 100 * time.Millisecond
	minAllowed := time.Duration(float64(base) * (1 - jitterFraction))
	maxAllowed := time.Duration(float64(base) * (1 + jitterFraction))

	sawDifferent := false
	first := addJitter(base)
	for i := 0; i < 200; i++ {
		got := addJitter(base)
		require.GreaterOrEqual(t, got, minAllowed, "jitter below lower bound")
		require.LessOrEqual(t, got, maxAllowed, "jitter above upper bound")
		if got != first {
			sawDifferent = true
		}
	}
	require.True(t, sawDifferent, "addJitter should produce varied results")
}

func TestAddJitter_ZeroDuration(t *testing.T) {
	require.Equal(t, time.Duration(0), addJitter(0))
	require.Equal(t, time.Duration(-1), addJitter(-1))
}

func TestBaseTrigger_BackoffDelaysRetransmit(t *testing.T) {
	lggr, err := logger.New()
	require.NoError(t, err)
	ctx := t.Context()

	getter := &atomicJSONGetter{}
	require.NoError(t, getter.setJSON(`{
		"global": {
			"BaseTriggerRetransmitEnabled": "true",
			"BaseTriggerRetryInterval": "100ms",
			"BaseTriggerMaxRetries": "10"
		}
	}`))

	store := NewMemEventStore()
	b, err := NewBaseTriggerCapabilityWithCRESettings(ctx, store,
		func() *wrapperspb.BytesValue { return &wrapperspb.BytesValue{} },
		lggr, "testCap", getter)
	require.NoError(t, err)

	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 100)
	b.RegisterTrigger("trig", sendCh)
	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() { b.Stop() })

	te := makeTE(t, "trig", "e1", []byte("payload"))
	require.NoError(t, b.DeliverEvent(ctx, te, "trig"))

	// Drain sends over 500ms. With backoff the gap between retries should
	// grow: ~100ms for 1st retry, ~200ms for 2nd, ~400ms for 3rd.
	// Without backoff we'd see 5+ sends; with it we expect fewer.
	time.Sleep(500 * time.Millisecond)
	count := 0
drain:
	for {
		select {
		case <-sendCh:
			count++
		default:
			break drain
		}
	}
	// First send is immediate (from DeliverEvent), then retries back off.
	// In 500ms with 100ms base, flat retries would yield ~5 sends.
	// With exponential backoff (100, 200, 400ms) we expect ~3.
	require.LessOrEqual(t, count, 4, "backoff should reduce number of resends in the window")
	require.GreaterOrEqual(t, count, 1, "should have at least the initial send")
}

func TestBaseTrigger_MaxSendsPerTick(t *testing.T) {
	store := NewMemEventStore()

	lggr, err := logger.New()
	require.NoError(t, err)

	getter := &atomicJSONGetter{}
	require.NoError(t, getter.setJSON(`{
		"global": {
			"BaseTriggerRetransmitEnabled": "true",
			"BaseTriggerRetryInterval": "50ms",
			"BaseTriggerMaxRetries": "1000"
		}
	}`))

	b, err := NewBaseTriggerCapabilityWithCRESettings(context.Background(), store,
		func() *wrapperspb.BytesValue { return &wrapperspb.BytesValue{} },
		lggr, "testCap", getter)
	require.NoError(t, err)

	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 500)
	b.RegisterTrigger("trig", sendCh)

	// Inject more events than the per-tick cap directly into pending.
	b.mu.Lock()
	reg := b.byTrigger["trig"]
	require.NotNil(t, reg)
	reg.pending = make(map[string]*PendingEvent)
	n := defaultMaxSendsPerTick + 30
	for i := 0; i < n; i++ {
		eid := fmt.Sprintf("e%d", i)
		rec := &PendingEvent{
			TriggerId: "trig",
			EventId:   eid,
			Payload:   []byte("x"),
			FirstAt:   time.Now().Add(-time.Minute),
		}
		reg.pending[eid] = rec
	}
	b.mu.Unlock()

	// Manually call scanPending once.
	b.scanPending()

	// Count how many were actually sent.
	sent := 0
drainCap:
	for {
		select {
		case <-sendCh:
			sent++
		default:
			break drainCap
		}
	}
	require.LessOrEqual(t, sent, defaultMaxSendsPerTick,
		"scanPending should respect the per-tick cap")
}

func TestBaseTrigger_PruneStaleEvents(t *testing.T) {
	lggr, err := logger.New()
	require.NoError(t, err)
	ctx := t.Context()

	store := NewMemEventStore()

	// Pre-populate store with an old event not tracked in memory.
	oldRec := PendingEvent{
		TriggerId: "orphanTrig",
		EventId:   "orphanEvent",
		Payload:   []byte("data"),
		FirstAt:   time.Now().Add(-48 * time.Hour),
		Attempts:  50,
	}
	require.NoError(t, store.Insert(ctx, oldRec))

	// Also add a recent event that should NOT be pruned.
	recentRec := PendingEvent{
		TriggerId: "recentTrig",
		EventId:   "recentEvent",
		Payload:   []byte("fresh"),
		FirstAt:   time.Now(),
		Attempts:  1,
	}
	require.NoError(t, store.Insert(ctx, recentRec))

	getter := &atomicJSONGetter{}
	require.NoError(t, getter.setJSON(`{
		"global": {
			"BaseTriggerRetransmitEnabled": "false",
			"BaseTriggerRetryInterval": "30s",
			"BaseTriggerPruneAge": "1h"
		}
	}`))

	b := NewBaseTriggerCapability(store,
		func() *wrapperspb.BytesValue { return &wrapperspb.BytesValue{} },
		lggr, "testCap", 0, getter)

	// Don't Start the full loop — just call pruneStaleEvents directly.
	b.pruneStaleEvents()

	recs, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, recs, 1, "only the recent event should remain")
	require.Equal(t, "recentEvent", recs[0].EventId)
}

func TestBaseTrigger_PruneSkipsInMemoryEvents(t *testing.T) {
	lggr, err := logger.New()
	require.NoError(t, err)
	ctx := t.Context()

	store := NewMemEventStore()

	// Old event that IS tracked in memory.
	oldRec := PendingEvent{
		TriggerId: "trig",
		EventId:   "inMemory",
		Payload:   []byte("data"),
		FirstAt:   time.Now().Add(-48 * time.Hour),
		Attempts:  50,
	}
	require.NoError(t, store.Insert(ctx, oldRec))

	getter := &atomicJSONGetter{}
	require.NoError(t, getter.setJSON(`{
		"global": {
			"BaseTriggerRetransmitEnabled": "false",
			"BaseTriggerRetryInterval": "30s",
			"BaseTriggerPruneAge": "1h"
		}
	}`))

	b := NewBaseTriggerCapability(store,
		func() *wrapperspb.BytesValue { return &wrapperspb.BytesValue{} },
		lggr, "testCap", 0, getter)

	// Manually put event in memory to simulate it being actively tracked.
	b.mu.Lock()
	reg := b.getOrCreateRegLocked("trig")
	reg.pending["inMemory"] = &oldRec
	b.mu.Unlock()

	b.pruneStaleEvents()

	recs, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, recs, 1, "in-memory event should not be pruned even if old")
}

func TestBaseTrigger_ScanPendingSkipsEventsWithoutInbox(t *testing.T) {
	lggr, err := logger.New()
	require.NoError(t, err)
	ctx := t.Context()

	getter := &atomicJSONGetter{}
	require.NoError(t, getter.setJSON(`{
		"global": {
			"BaseTriggerRetransmitEnabled": "true",
			"BaseTriggerRetryInterval": "50ms",
			"BaseTriggerMaxRetries": "5"
		}
	}`))

	store := NewMemEventStore()

	// Build a valid protobuf payload for the events.
	msg := &wrapperspb.BytesValue{Value: []byte("payload")}
	anyMsg, err := anypb.New(msg)
	require.NoError(t, err)

	// Pre-populate the store with events (simulates restart loading).
	for i := 0; i < 5; i++ {
		require.NoError(t, store.Insert(ctx, PendingEvent{
			TriggerId:  "trig",
			EventId:    fmt.Sprintf("e%d", i),
			AnyTypeURL: anyMsg.GetTypeUrl(),
			Payload:    anyMsg.GetValue(),
			FirstAt:    time.Now().Add(-time.Minute),
		}))
	}

	b, err := NewBaseTriggerCapabilityWithCRESettings(ctx, store,
		func() *wrapperspb.BytesValue { return &wrapperspb.BytesValue{} },
		lggr, "testCap", getter)
	require.NoError(t, err)

	// Do NOT register a trigger (no inbox) — simulates post-restart state
	// before workflow registration arrives.
	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() { b.Stop() })

	// Let several scanPending ticks fire without an inbox.
	time.Sleep(300 * time.Millisecond)

	// Verify no attempts were incremented — events should have been skipped.
	b.mu.Lock()
	for _, rec := range b.byTrigger["trig"].pending {
		require.Equal(t, 0, rec.Attempts,
			"event %s should have 0 attempts when no inbox is registered", rec.EventId)
	}
	b.mu.Unlock()

	// Now register the inbox and let scanPending pick them up.
	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 100)
	b.RegisterTrigger("trig", sendCh)

	time.Sleep(200 * time.Millisecond)

	// Events should now be sent.
	count := 0
drain:
	for {
		select {
		case <-sendCh:
			count++
		default:
			break drain
		}
	}
	require.GreaterOrEqual(t, count, 1,
		"events should be retransmitted once the inbox is registered")

	// Verify attempts were incremented now that inbox exists.
	b.mu.Lock()
	for _, rec := range b.byTrigger["trig"].pending {
		require.Greater(t, rec.Attempts, 0,
			"event %s should have attempts > 0 after inbox is registered", rec.EventId)
	}
	b.mu.Unlock()
}

// errStore is a configurable EventStore that can inject errors on any method.
type errStore struct {
	*MemEventStore
	insertErr      error
	deleteEventErr error
}

func (s *errStore) Insert(ctx context.Context, r PendingEvent) error {
	if s.insertErr != nil {
		return s.insertErr
	}
	return s.MemEventStore.Insert(ctx, r)
}

func (s *errStore) DeleteEvent(ctx context.Context, triggerID, eventID string) error {
	if s.deleteEventErr != nil {
		return s.deleteEventErr
	}
	return s.MemEventStore.DeleteEvent(ctx, triggerID, eventID)
}

func TestBaseTrigger_DeliverEvent_SkipsAlreadyPending(t *testing.T) {
	store := NewMemEventStore()
	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 10)
	b := newBase(t, store)

	b.RegisterTrigger("trig", sendCh)
	require.NoError(t, b.Start(t.Context()))
	t.Cleanup(func() { b.Stop() })

	te := makeTE(t, "trig", "e1", []byte("payload"))

	// First delivery: goes through normally.
	require.NoError(t, b.DeliverEvent(t.Context(), te, "trig"))

	// Second delivery of the same event while still pending: must be silently skipped.
	require.NoError(t, b.DeliverEvent(t.Context(), te, "trig"))

	// Only one store row should exist (the second call must not insert).
	recs, err := store.List(t.Context())
	require.NoError(t, err)
	require.Len(t, recs, 1, "second delivery of a pending event must not insert a duplicate store row")
}

func TestBaseTrigger_DeliverEvent_DuplicateKeyFromStore_Skips(t *testing.T) {
	store := &errStore{
		MemEventStore: NewMemEventStore(),
		insertErr:     errors.New("ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)"),
	}
	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 10)
	b := newBase(t, store)

	b.RegisterTrigger("trig", sendCh)
	require.NoError(t, b.Start(t.Context()))
	t.Cleanup(func() { b.Stop() })

	te := makeTE(t, "trig", "e1", []byte("payload"))

	// The store rejects the insert as a duplicate — DeliverEvent must return nil (no error).
	require.NoError(t, b.DeliverEvent(t.Context(), te, "trig"),
		"duplicate key from store should be silently swallowed")

	// Nothing should have been added to in-memory pending.
	b.mu.Lock()
	reg := b.byTrigger["trig"]
	_, isPending := reg.pending["e1"]
	b.mu.Unlock()
	require.False(t, isPending, "event must not be in-memory pending after duplicate-key skip")
}

func TestBaseTrigger_DeliverEvent_StoreInsertError_Propagates(t *testing.T) {
	storeErr := errors.New("connection refused")
	store := &errStore{
		MemEventStore: NewMemEventStore(),
		insertErr:     storeErr,
	}
	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 10)
	b := newBase(t, store)

	b.RegisterTrigger("trig", sendCh)
	require.NoError(t, b.Start(t.Context()))
	t.Cleanup(func() { b.Stop() })

	te := makeTE(t, "trig", "e1", []byte("payload"))
	err := b.DeliverEvent(t.Context(), te, "trig")
	require.ErrorIs(t, err, storeErr, "generic store insert error must be surfaced to the caller")
}

func TestBaseTrigger_SendToInbox_NoInboxRegistered(t *testing.T) {
	store := NewMemEventStore()
	// tRetransmit=0 disables persistence so DeliverEvent calls sendToInbox directly.
	b := newBaseWithRetransmit(t, store, 0)

	require.NoError(t, b.Start(t.Context()))
	t.Cleanup(func() { b.Stop() })

	te := makeTE(t, "trig", "e1", []byte("payload"))
	err := b.DeliverEvent(t.Context(), te, "trig")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no inbox registered for trigger trig")
}

func TestBaseTrigger_SendToInbox_InboxFull(t *testing.T) {
	store := NewMemEventStore()
	// Capacity-0 channel is always full from the sender's perspective.
	fullCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 0)
	b := newBaseWithRetransmit(t, store, 0)

	b.RegisterTrigger("trig", fullCh)
	require.NoError(t, b.Start(t.Context()))
	t.Cleanup(func() { b.Stop() })

	te := makeTE(t, "trig", "e1", []byte("payload"))
	err := b.DeliverEvent(t.Context(), te, "trig")
	require.Error(t, err)
	require.Contains(t, err.Error(), "inbox full or closed for trigger trig")
}

func TestBaseTrigger_AckEvent_NilPendingRecord(t *testing.T) {
	store := NewMemEventStore()
	b := newBase(t, store)

	require.NoError(t, b.Start(t.Context()))
	t.Cleanup(func() { b.Stop() })

	// Manually plant a nil record — this is an abnormal state but must not panic.
	b.mu.Lock()
	reg := b.getOrCreateRegLocked("trig")
	reg.pending["e1"] = nil
	b.mu.Unlock()

	// AckEvent must not panic and must still reach the store delete.
	require.NoError(t, b.AckEvent(t.Context(), "trig", "e1"))

	// The nil record must be cleaned up from pending.
	b.mu.Lock()
	_, still := b.byTrigger["trig"].pending["e1"]
	b.mu.Unlock()
	require.False(t, still, "nil pending record must be removed after ACK")
}

func TestBaseTrigger_AckEvent_StoreDeleteError_Propagates(t *testing.T) {
	deleteErr := errors.New("store unavailable")
	store := &errStore{
		MemEventStore:  NewMemEventStore(),
		deleteEventErr: deleteErr,
	}
	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 10)
	b := newBase(t, store)

	b.RegisterTrigger("trig", sendCh)
	require.NoError(t, b.Start(t.Context()))
	t.Cleanup(func() { b.Stop() })

	te := makeTE(t, "trig", "e1", []byte("payload"))
	// Use the real MemEventStore for Insert so pending is populated normally;
	// only DeleteEvent is broken.
	store.insertErr = nil
	require.NoError(t, b.DeliverEvent(t.Context(), te, "trig"))

	err := b.AckEvent(t.Context(), "trig", "e1")
	require.ErrorIs(t, err, deleteErr, "store delete error must be returned from AckEvent")
}

func TestIsDuplicateKeyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"unrelated error", errors.New("connection refused"), false},
		{"sqlstate 23505", errors.New("ERROR: duplicate key (SQLSTATE 23505)"), true},
		{"duplicate key text", errors.New("duplicate key value violates unique constraint"), true},
		{"both markers", errors.New("duplicate key 23505"), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, isDuplicateKeyError(tc.err))
		})
	}
}

func TestBaseTrigger_ScanPending_NewEventsBeforeRetries(t *testing.T) {
	store := NewMemEventStore()
	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 20)
	b := newBase(t, store)
	b.RegisterTrigger("trig", sendCh)
	// No Start — call scanPending directly to avoid background-loop interference.

	stale := time.Now().Add(-time.Minute) // ensure backoff has elapsed
	b.mu.Lock()
	reg := b.byTrigger["trig"]
	reg.pending["z-new"] = &PendingEvent{TriggerId: "trig", EventId: "z-new", Attempts: 0, FirstAt: stale, Payload: []byte{}}
	reg.pending["e-retry"] = &PendingEvent{TriggerId: "trig", EventId: "e-retry", Attempts: 3, FirstAt: stale, LastSentAt: stale, Payload: []byte{}}
	reg.pending["a-new"] = &PendingEvent{TriggerId: "trig", EventId: "a-new", Attempts: 0, FirstAt: stale, Payload: []byte{}}
	b.mu.Unlock()

	b.scanPending()

	// Buffered channel is FIFO — read order matches send order.
	var got []string
drain:
	for {
		select {
		case v := <-sendCh:
			got = append(got, v.Id)
		default:
			break drain
		}
	}

	require.Equal(t, []string{"a-new", "z-new", "e-retry"}, got,
		"new events must precede retries; within each tier events sort alphabetically by EventId")
}

func TestBaseTrigger_ScanPending_TriggerIDTiebreaker(t *testing.T) {
	lggr, err := logger.New()
	require.NoError(t, err)

	getter := &atomicJSONGetter{}
	require.NoError(t, getter.setJSON(`{
		"global": {
			"BaseTriggerRetransmitEnabled": "true",
			"BaseTriggerRetryInterval":     "1ms",
			"BaseTriggerMaxSendsPerTick":   "1",
			"BaseTriggerMaxRetries":        "1000"
		}
	}`))

	store := NewMemEventStore()
	b, err := NewBaseTriggerCapabilityWithCRESettings(context.Background(), store,
		func() *wrapperspb.BytesValue { return &wrapperspb.BytesValue{} },
		lggr, "testCap", getter)
	require.NoError(t, err)

	sendChA := make(chan TriggerAndId[*wrapperspb.BytesValue], 5)
	sendChB := make(chan TriggerAndId[*wrapperspb.BytesValue], 5)
	b.RegisterTrigger("trigA", sendChA)
	b.RegisterTrigger("trigB", sendChB)

	// Same EventId on both triggers — only differentiator in the sort is TriggerId.
	stale := time.Now().Add(-time.Minute)
	b.mu.Lock()
	b.byTrigger["trigA"].pending["e1"] = &PendingEvent{TriggerId: "trigA", EventId: "e1", Attempts: 2, FirstAt: stale, LastSentAt: stale, Payload: []byte{}}
	b.byTrigger["trigB"].pending["e1"] = &PendingEvent{TriggerId: "trigB", EventId: "e1", Attempts: 2, FirstAt: stale, LastSentAt: stale, Payload: []byte{}}
	b.mu.Unlock()

	// Cap=1: only the event from the alphabetically first trigger is sent.
	b.scanPending()

	require.Equal(t, 1, len(sendChA), "trigA must be the one event sent under the cap")
	require.Equal(t, 0, len(sendChB), "trigB must be deferred: trigA sorts first by TriggerId")
}

func TestBaseTrigger_ExpirePreAcked_RemovesStaleEntries(t *testing.T) {
	store := NewMemEventStore()
	b := newBase(t, store)

	require.NoError(t, b.Start(t.Context()))
	t.Cleanup(func() { b.Stop() })

	sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 10)
	b.RegisterTrigger("trig", sendCh)

	// Plant two pre-ACK entries: one fresh, one expired (25 h old).
	b.mu.Lock()
	reg := b.byTrigger["trig"]
	reg.preAcked["fresh"] = time.Now().Add(-time.Hour)
	reg.preAcked["stale"] = time.Now().Add(-25 * time.Hour)
	b.mu.Unlock()

	// Call expirePreAcked directly with the current time.
	b.mu.Lock()
	b.expirePreAcked(time.Now())
	b.mu.Unlock()

	b.mu.Lock()
	_, freshExists := b.byTrigger["trig"].preAcked["fresh"]
	_, staleExists := b.byTrigger["trig"].preAcked["stale"]
	b.mu.Unlock()

	require.True(t, freshExists, "entry within TTL must be retained")
	require.False(t, staleExists, "entry older than 24 h must be evicted")
}
