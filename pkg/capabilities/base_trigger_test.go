package capabilities

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
)

type memStore struct {
	mu   sync.Mutex
	recs map[string]PendingEvent
}

func newMemStore() *memStore {
	return &memStore{recs: make(map[string]PendingEvent)}
}

func (m *memStore) Insert(ctx context.Context, rec PendingEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recs[key(rec.TriggerId, rec.WorkflowId, rec.EventId)] = rec
	return nil
}

func (m *memStore) Delete(ctx context.Context, triggerId, workflowId, eventId string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.recs, key(triggerId, workflowId, eventId))
	return nil
}

func (m *memStore) List(ctx context.Context) ([]PendingEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]PendingEvent, 0, len(m.recs))
	for _, r := range m.recs {
		out = append(out, r)
	}
	return out, nil
}

type sendProbe struct {
	mu    sync.Mutex
	calls []struct {
		TE         TriggerEvent
		WorkflowID string
	}
	// if set, return error for the first N sends
	failFirstN int32
}

func (p *sendProbe) fn(ctx context.Context, te TriggerEvent, workflowId string) error {
	// Optionally fail some sends
	if atomic.LoadInt32(&p.failFirstN) > 0 {
		atomic.AddInt32(&p.failFirstN, -1)
		return assertErr // sentinel below
	}
	p.mu.Lock()
	p.calls = append(p.calls, struct {
		TE         TriggerEvent
		WorkflowID string
	}{te, workflowId})
	p.mu.Unlock()
	return nil
}

func (p *sendProbe) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.calls)
}

var assertErr = &temporaryError{"boom"}

type temporaryError struct{ s string }

func (e *temporaryError) Error() string { return e.s }

// lost hook probe
type lostProbe struct {
	mu    sync.Mutex
	calls []PendingEvent
}

func (p *lostProbe) fn(ctx context.Context, rec PendingEvent) {
	p.mu.Lock()
	p.calls = append(p.calls, rec)
	p.mu.Unlock()
}

func (p *lostProbe) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.calls)
}

func newBase(t *testing.T, store EventStore, send OutboundSend, lost LostHook) *BaseTriggerCapability {
	t.Helper()
	b := &BaseTriggerCapability{
		tRetransmit: 30 * time.Millisecond,
		tMax:        120 * time.Millisecond,
		store:       store,
		send:        send,
		lost:        lost,
		// lggr:     your test logger if desired
	}
	return b
}

func ctxWithCancel(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(t.Context())
	return ctx, cancel
}

func TestStart_LoadsAndSendsPersisted(t *testing.T) {
	store := newMemStore()
	probe := &sendProbe{}
	lostp := &lostProbe{}

	// Preload store with one record
	rec := PendingEvent{
		TriggerId:  "trigA",
		WorkflowId: "wf1",
		EventId:    "e1",
		AnyTypeURL: "type.googleapis.com/some.Msg",
		Payload:    []byte("payload"),
		FirstAt:    time.Now().Add(-1 * time.Minute),
	}
	require.NoError(t, store.Insert(context.Background(), rec))

	b := newBase(t, store, probe.fn, lostp.fn)
	ctx, cancel := ctxWithCancel(t)
	defer cancel()

	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() { b.cancel(); b.wg.Wait() })

	// Initial send triggered on Start
	require.Eventually(t, func() bool { return probe.count() >= 1 }, 200*time.Millisecond, 5*time.Millisecond)
}

func TestDeliverEvent_PersistsAndSends(t *testing.T) {
	store := newMemStore()
	probe := &sendProbe{}
	lostp := &lostProbe{}
	b := newBase(t, store, probe.fn, lostp.fn)
	ctx, cancel := ctxWithCancel(t)
	defer cancel()

	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() { b.cancel(); b.wg.Wait() })

	te := TriggerEvent{
		TriggerType: "trigB",
		ID:          "e2",
		Payload:     &anypb.Any{TypeUrl: "type.googleapis.com/thing", Value: []byte("x")},
	}
	require.NoError(t, b.deliverEvent(ctx, te, []string{"wf1", "wf2"}))

	// Persisted twice (two workflows)
	recs, _ := store.List(ctx)
	require.Len(t, recs, 2)

	// Sent twice
	require.Eventually(t, func() bool { return probe.count() >= 2 }, 200*time.Millisecond, 5*time.Millisecond)
}

func TestAckEvent_StopsRetransmit(t *testing.T) {
	store := newMemStore()
	probe := &sendProbe{}
	lostp := &lostProbe{}
	b := newBase(t, store, probe.fn, lostp.fn)
	ctx, cancel := ctxWithCancel(t)
	defer cancel()

	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() { b.cancel(); b.wg.Wait() })

	te := TriggerEvent{
		TriggerType: "trigC",
		ID:          "e3",
		Payload:     &anypb.Any{TypeUrl: "type.googleapis.com/thing", Value: []byte("x")},
	}
	require.NoError(t, b.deliverEvent(ctx, te, []string{"wf1"}))
	require.Eventually(t, func() bool { return probe.count() >= 1 }, 200*time.Millisecond, 5*time.Millisecond)

	// Ack and ensure no more sends occur after a couple of retransmit periods
	require.NoError(t, b.AckEvent(ctx, "trigC", "wf1", "e3"))
	sentBefore := probe.count()
	time.Sleep(3 * b.tRetransmit)
	require.Equal(t, sentBefore, probe.count(), "no further retransmits after ACK")
}

func TestRetryThenLost_AfterTmax(t *testing.T) {
	store := newMemStore()
	probe := &sendProbe{failFirstN: 1000} // always fail
	lostp := &lostProbe{}
	b := newBase(t, store, probe.fn, lostp.fn)

	// tighten timers for the test
	b.tRetransmit = 20 * time.Millisecond
	b.tMax = 80 * time.Millisecond

	ctx, cancel := ctxWithCancel(t)
	defer cancel()

	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() { b.cancel(); b.wg.Wait() })

	te := TriggerEvent{
		TriggerType: "trigD",
		ID:          "e4",
		Payload:     &anypb.Any{TypeUrl: "type.googleapis.com/thing", Value: []byte("x")},
	}
	require.NoError(t, b.deliverEvent(ctx, te, []string{"wf1"}))

	// Should attempt several sends, then mark lost and delete from store
	require.Eventually(t, func() bool { return lostp.count() >= 1 }, 500*time.Millisecond, 5*time.Millisecond)

	// Ensure the record is gone from the store after lost
	recs, _ := store.List(ctx)
	require.Len(t, recs, 0)
}

func TestTrySendErrorIsIgnoredByCallSites(t *testing.T) {
	store := newMemStore()
	probe := &sendProbe{failFirstN: 1} // first send fails
	lostp := &lostProbe{}
	b := newBase(t, store, probe.fn, lostp.fn)
	ctx, cancel := ctxWithCancel(t)
	defer cancel()

	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() { b.cancel(); b.wg.Wait() })

	te := TriggerEvent{
		TriggerType: "trigE",
		ID:          "e5",
		Payload:     &anypb.Any{TypeUrl: "type.googleapis.com/thing", Value: []byte("x")},
	}
	// deliverEvent should not return an error even if the first send fails;
	// retransmitLoop will retry later.
	require.NoError(t, b.deliverEvent(ctx, te, []string{"wf1"}))

	// Eventually should succeed on a subsequent attempt (after the first forced failure)
	require.Eventually(t, func() bool { return probe.count() >= 1 }, 300*time.Millisecond, 5*time.Millisecond)
}
