package capabilities

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
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
	m.recs[key(rec.TriggerId, rec.EventId)] = rec
	return nil
}

func (m *memStore) DeleteEvent(ctx context.Context, triggerId, eventId string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.recs, key(triggerId, eventId))
	return nil
}

func (m *memStore) DeleteEventsForTrigger(ctx context.Context, triggerId string) error {
	events, err := m.List(ctx)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	for _, event := range events {
		if event.TriggerId == triggerId {
			delete(m.recs, key(triggerId, event.EventId))
		}
	}
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

func newBase(t *testing.T, store EventStore, lost LostHook) *BaseTriggerCapability[TriggerEvent] {
	lggr, err := logger.New()
	require.NoError(t, err)
	decode := func(te TriggerEvent) (TriggerEvent, error) {
		return te, nil
	}
	return NewBaseTriggerCapability(store, decode, lost, lggr, 100*time.Millisecond, 10*time.Minute)
}

func ctxWithCancel(t *testing.T) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(t.Context())
	return ctx, cancel
}

func TestStart_LoadsAndSendsPersisted(t *testing.T) {
	store := newMemStore()
	lostp := &lostProbe{}
	sendCh := make(chan TriggerEvent, 10)

	// Preload store with one record
	rec := PendingEvent{
		TriggerId:  "trigA",
		EventId:    "e1",
		AnyTypeURL: "type.googleapis.com/some.Msg",
		Payload:    []byte("payload"),
		FirstAt:    time.Now().Add(-1 * time.Minute),
	}
	require.NoError(t, store.Insert(context.Background(), rec))

	b := newBase(t, store, lostp.fn)

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
		//return probe.count() >= 1
	}, 200*time.Millisecond, 5*time.Millisecond)
}

func TestDeliverEvent_PersistsAndSends(t *testing.T) {
	store := newMemStore()
	lostp := &lostProbe{}
	sendCh := make(chan TriggerEvent, 10)

	b := newBase(t, store, lostp.fn)
	ctx, cancel := ctxWithCancel(t)
	defer cancel()

	b.RegisterTrigger("trigA", sendCh)

	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() {
		b.Stop()
		b.UnregisterTrigger("trigA")
	})

	te := TriggerEvent{
		TriggerType: "trigA",
		ID:          "e2",
		Payload:     &anypb.Any{TypeUrl: "type.googleapis.com/thing", Value: []byte("x")},
	}
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
	store := newMemStore()
	sendCh := make(chan TriggerEvent, 10)

	lostp := &lostProbe{}
	b := newBase(t, store, lostp.fn)
	ctx, cancel := ctxWithCancel(t)
	defer cancel()

	b.RegisterTrigger("trigC", sendCh)

	require.NoError(t, b.Start(ctx))
	t.Cleanup(func() {
		b.Stop()
		b.UnregisterTrigger("trigC")
	})

	te := TriggerEvent{
		TriggerType: "trigC",
		ID:          "e3",
		Payload:     &anypb.Any{TypeUrl: "type.googleapis.com/thing", Value: []byte("x")},
	}
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
