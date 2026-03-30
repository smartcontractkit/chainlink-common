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
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

func TestValidateBaseTriggerRetryInterval(t *testing.T) {
	ctx := context.Background()

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

// flippingJSONGetter supports mutating CRE JSON for dynamic-settings tests.
type flippingJSONGetter struct {
	mu  sync.Mutex
	raw []byte
}

func (f *flippingJSONGetter) setJSON(js string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.raw = []byte(js)
}

func (f *flippingJSONGetter) GetScoped(ctx context.Context, scope settings.Scope, key string) (string, error) {
	f.mu.Lock()
	raw := append([]byte(nil), f.raw...)
	f.mu.Unlock()
	g, err := settings.NewJSONGetter(raw)
	if err != nil {
		return "", err
	}
	return g.GetScoped(ctx, scope, key)
}

func TestBaseTrigger_CRE_DynamicDisableStopsResend(t *testing.T) {
	lggr, err := logger.New()
	require.NoError(t, err)
	ctx := context.Background()

	getter := &flippingJSONGetter{}
	getter.setJSON(`{
		"global": {
			"BaseTriggerRetransmitEnabled": "true",
			"BaseTriggerRetryInterval": "20ms"
		}
	}`)

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

	getter.setJSON(`{
		"global": {
			"BaseTriggerRetransmitEnabled": "false",
			"BaseTriggerRetryInterval": "20ms"
		}
	}`)

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
		"testCap", tRetransmit, 0, 0, nil)
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

func TestBaseTrigger_UndeliveredStateAlerting(t *testing.T) {
	type testCase struct {
		name               string
		warnAfter          time.Duration
		criticalAfter      time.Duration
		expectWarning      bool
		expectCritical     bool
		expectClearedOnAck bool
	}

	tests := []testCase{
		{
			name:          "warning fires",
			warnAfter:     200 * time.Millisecond,
			expectWarning: true,
		},
		{
			name:           "critical fires",
			warnAfter:      100 * time.Millisecond,
			criticalAfter:  300 * time.Millisecond,
			expectWarning:  true,
			expectCritical: true,
		},
		{
			name:               "cleared on ack",
			warnAfter:          200 * time.Millisecond,
			expectWarning:      true,
			expectClearedOnAck: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := NewMemEventStore()
			sendCh := make(chan TriggerAndId[*wrapperspb.BytesValue], 10)

			lggr, err := logger.New()
			require.NoError(t, err)

			b := NewBaseTriggerCapability(
				store,
				func() *wrapperspb.BytesValue { return &wrapperspb.BytesValue{} },
				lggr,
				"testCap",
				50*time.Millisecond,
				tc.warnAfter,
				tc.criticalAfter,
				nil,
			)

			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			b.RegisterTrigger("trig", sendCh)
			require.NoError(t, b.Start(ctx))
			t.Cleanup(func() { b.Stop() })

			te := makeTE(t, "trig", "e1", []byte("x"))
			require.NoError(t, b.DeliverEvent(ctx, te, "trig"))

			// Wait for expected thresholds
			require.Eventually(t, func() bool {
				b.mu.Lock()
				defer b.mu.Unlock()

				state := b.undeliveredAlertStates["trig"]["e1"]
				if state == nil {
					return false
				}

				if tc.expectCritical {
					return state.emittedCritical
				}
				if tc.expectWarning {
					return state.emittedWarning
				}
				return true
			}, 3*time.Second, 10*time.Millisecond)

			if tc.expectClearedOnAck {
				require.NoError(t, b.AckEvent(ctx, "trig", "e1"))

				require.Eventually(t, func() bool {
					b.mu.Lock()
					defer b.mu.Unlock()
					_, exists := b.undeliveredAlertStates["trig"]
					return !exists
				}, 1*time.Second, 10*time.Millisecond)
			}
		})
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
