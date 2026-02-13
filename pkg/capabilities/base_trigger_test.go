package capabilities

import (
	"context"
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
	return NewBaseTriggerCapability(store, func() *wrapperspb.BytesValue { return &wrapperspb.BytesValue{} }, lggr, "testCap", 100*time.Millisecond)
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
