package capabilities

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func newBase(t *testing.T, store EventStore) *BaseTriggerCapability[*wrapperspb.BytesValue] {
	lggr, err := logger.New()
	require.NoError(t, err)
	return NewBaseTriggerCapability(store, func() *wrapperspb.BytesValue { return &wrapperspb.BytesValue{} }, lggr, "testCap", 100*time.Millisecond, nil)
}

func newBaseWithMetrics(t *testing.T, store EventStore, metrics BaseTriggerMetrics, t1, t2 time.Duration) *BaseTriggerCapability[*wrapperspb.BytesValue] {
	lggr, err := logger.New()
	require.NoError(t, err)

	return NewBaseTriggerCapability(
		store,
		func() *wrapperspb.BytesValue { return &wrapperspb.BytesValue{} },
		lggr,
		"testCap",
		100*time.Millisecond,
		&BaseTriggerOpts{
			Metrics:           metrics,
			UndeliveredAfter:  t1,
			UndeliveredAfter2: t2,
		},
	)
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

type mockMetrics struct {
	mu sync.Mutex

	retries           []string
	acks              []string
	timeToAckObserved []string
	undelivered1      []string
	undelivered2      []string
	inboxMissing      []string
	inboxFull         []string
}

func (m *mockMetrics) IncRetry(triggerID, eventID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.retries = append(m.retries, triggerID+":"+eventID)
}

func (m *mockMetrics) IncAck(triggerID, eventID string, attempts int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.acks = append(m.acks, triggerID+":"+eventID)
}

func (m *mockMetrics) ObserveTimeToAck(triggerID, eventID string, d time.Duration, attempts int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.timeToAckObserved = append(m.timeToAckObserved, triggerID+":"+eventID)
}

func (m *mockMetrics) IncInboxMissing(triggerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inboxMissing = append(m.inboxMissing, triggerID)
}

func (m *mockMetrics) IncInboxFull(triggerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.inboxFull = append(m.inboxFull, triggerID)
}

func (m *mockMetrics) EmitUndelivered(triggerID, eventID string, age time.Duration, attempts int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.undelivered1 = append(m.undelivered1, triggerID+":"+eventID)
}

func (m *mockMetrics) EmitUndelivered2(triggerID, eventID string, age time.Duration, attempts int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.undelivered2 = append(m.undelivered2, triggerID+":"+eventID)
}

func TestBaseTrigger_Metrics(t *testing.T) {
	type testCase struct {
		name string
		t1   time.Duration
		t2   time.Duration
		run  func(t *testing.T,
			b *BaseTriggerCapability[*wrapperspb.BytesValue],
			mock *mockMetrics,
			sendCh chan TriggerAndId[*wrapperspb.BytesValue],
		)
	}

	tests := []testCase{
		{
			name: "retry increments metric",
			run: func(t *testing.T, b *BaseTriggerCapability[*wrapperspb.BytesValue], mock *mockMetrics, sendCh chan TriggerAndId[*wrapperspb.BytesValue]) {
				ctx := t.Context()

				b.RegisterTrigger("trig", sendCh)
				require.NoError(t, b.Start(ctx))
				t.Cleanup(func() { b.Stop() })

				te := makeTE(t, "trig", "e1", []byte("x"))
				require.NoError(t, b.DeliverEvent(ctx, te, "trig"))

				require.Eventually(t, func() bool {
					mock.mu.Lock()
					defer mock.mu.Unlock()
					return len(mock.retries) > 0
				}, 1*time.Second, 10*time.Millisecond)
			},
		},
		{
			name: "ack metrics fire",
			run: func(t *testing.T, b *BaseTriggerCapability[*wrapperspb.BytesValue], mock *mockMetrics, sendCh chan TriggerAndId[*wrapperspb.BytesValue]) {
				ctx := t.Context()

				b.RegisterTrigger("trig", sendCh)
				require.NoError(t, b.Start(ctx))
				t.Cleanup(func() { b.Stop() })

				te := makeTE(t, "trig", "e2", []byte("x"))
				require.NoError(t, b.DeliverEvent(ctx, te, "trig"))

				require.Eventually(t, func() bool {
					select {
					case <-sendCh:
						return true
					default:
						return false
					}
				}, 1*time.Second, 10*time.Millisecond)

				require.NoError(t, b.AckEvent(ctx, "trig", "e2"))

				require.Eventually(t, func() bool {
					mock.mu.Lock()
					defer mock.mu.Unlock()
					return len(mock.acks) == 1 &&
						len(mock.timeToAckObserved) == 1
				}, 1*time.Second, 10*time.Millisecond)
			},
		},
		{
			name: "undelivered threshold fires once",
			t1:   200 * time.Millisecond,
			run: func(t *testing.T, b *BaseTriggerCapability[*wrapperspb.BytesValue], mock *mockMetrics, sendCh chan TriggerAndId[*wrapperspb.BytesValue]) {
				ctx := t.Context()

				b.RegisterTrigger("trig", sendCh)
				require.NoError(t, b.Start(ctx))
				t.Cleanup(func() { b.Stop() })

				te := makeTE(t, "trig", "e3", []byte("x"))
				require.NoError(t, b.DeliverEvent(ctx, te, "trig"))

				require.Eventually(t, func() bool {
					mock.mu.Lock()
					defer mock.mu.Unlock()
					return len(mock.undelivered1) == 1
				}, 2*time.Second, 10*time.Millisecond)

				// Ensure it doesn't fire again
				time.Sleep(500 * time.Millisecond)

				mock.mu.Lock()
				defer mock.mu.Unlock()
				require.Len(t, mock.undelivered1, 1)
			},
		},
		{
			name: "undelivered2 fires after larger threshold",
			t1:   100 * time.Millisecond,
			t2:   300 * time.Millisecond,
			run: func(t *testing.T, b *BaseTriggerCapability[*wrapperspb.BytesValue], mock *mockMetrics, sendCh chan TriggerAndId[*wrapperspb.BytesValue]) {
				ctx := t.Context()

				b.RegisterTrigger("trig", sendCh)
				require.NoError(t, b.Start(ctx))
				t.Cleanup(func() { b.Stop() })

				te := makeTE(t, "trig", "e4", []byte("x"))
				require.NoError(t, b.DeliverEvent(ctx, te, "trig"))

				require.Eventually(t, func() bool {
					mock.mu.Lock()
					defer mock.mu.Unlock()
					return len(mock.undelivered2) == 1
				}, 3*time.Second, 10*time.Millisecond)
			},
		},
		{
			name: "undelivered cleared on ack",
			t1:   200 * time.Millisecond,
			run: func(t *testing.T, b *BaseTriggerCapability[*wrapperspb.BytesValue], mock *mockMetrics, sendCh chan TriggerAndId[*wrapperspb.BytesValue]) {
				ctx := t.Context()

				b.RegisterTrigger("trig", sendCh)
				require.NoError(t, b.Start(ctx))
				t.Cleanup(func() { b.Stop() })

				te := makeTE(t, "trig", "e5", []byte("x"))
				require.NoError(t, b.DeliverEvent(ctx, te, "trig"))

				require.Eventually(t, func() bool {
					mock.mu.Lock()
					defer mock.mu.Unlock()
					return len(mock.undelivered1) == 1
				}, 2*time.Second, 10*time.Millisecond)

				require.NoError(t, b.AckEvent(ctx, "trig", "e5"))

				// Give some time to ensure no re-fire
				time.Sleep(500 * time.Millisecond)

				mock.mu.Lock()
				defer mock.mu.Unlock()
				require.Len(t, mock.undelivered1, 1)
			},
		},
	}

	for _, tc := range tests {
		tc := tc // capture loop variable
		t.Run(tc.name, func(t *testing.T) {
			store := NewMemEventStore()                                   // <-- moved inside
			sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 10) // <-- moved inside
			mock := &mockMetrics{}

			b := newBaseWithMetrics(t, store, mock, tc.t1, tc.t2)
			tc.run(t, b, mock, sendCh)
		})
	}
}
