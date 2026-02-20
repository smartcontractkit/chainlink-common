package capabilities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// helper builds a PendingEvent with small defaults
func makePendingEvent(triggerID, eventID string, payload []byte, offset time.Duration) PendingEvent {
	now := time.Now().Add(offset)
	return PendingEvent{
		TriggerId:  triggerID,
		EventId:    eventID,
		AnyTypeURL: "type.googleapis.com/test.Msg",
		Payload:    append([]byte(nil), payload...),
		FirstAt:    now,
		LastSentAt: time.Time{},
		Attempts:   0,
	}
}

func TestMemEventStore_InsertListDelete(t *testing.T) {
	ctx := t.Context()
	store := NewMemEventStore()

	// insert a few events across two triggers
	e1 := makePendingEvent("trigA", "e1", []byte("p1"), -3*time.Minute)
	e2 := makePendingEvent("trigA", "e2", []byte("p2"), -2*time.Minute)
	e3 := makePendingEvent("trigB", "e3", []byte("p3"), -1*time.Minute)

	require.NoError(t, store.Insert(ctx, e1))
	require.NoError(t, store.Insert(ctx, e2))
	require.NoError(t, store.Insert(ctx, e3))

	// List should return three records (order is not guaranteed)
	recs, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, recs, 3)

	// Build a map for easier assertions
	got := make(map[string]PendingEvent)
	for _, r := range recs {
		got[r.EventId] = r
	}

	require.Contains(t, got, "e1")
	require.Equal(t, []byte("p1"), got["e1"].Payload)
	require.Contains(t, got, "e2")
	require.Equal(t, []byte("p2"), got["e2"].Payload)
	require.Contains(t, got, "e3")
	require.Equal(t, []byte("p3"), got["e3"].Payload)

	// Delete a single event and ensure it's gone
	require.NoError(t, store.DeleteEvent(ctx, "trigA", "e1"))
	recs, err = store.List(ctx)
	require.NoError(t, err)
	require.Len(t, recs, 2)
	for _, r := range recs {
		require.NotEqual(t, "e1", r.EventId)
	}

	// Delete all for triggerA
	require.NoError(t, store.DeleteEventsForTrigger(ctx, "trigA"))
	recs, err = store.List(ctx)
	require.NoError(t, err)
	require.Len(t, recs, 1)
	require.Equal(t, "e3", recs[0].EventId)
}

func TestMemEventStore_InsertIdempotentAndUpdateDelivery(t *testing.T) {
	ctx := t.Context()
	store := NewMemEventStore()

	triggerID := "trig-upd"
	eventID := "evt1"

	first := makePendingEvent(triggerID, eventID, []byte("first"), -1*time.Minute)
	require.NoError(t, store.Insert(ctx, first))

	// Verify stored value
	recs, err := store.List(ctx)
	require.NoError(t, err)
	require.Len(t, recs, 1)
	require.Equal(t, []byte("first"), recs[0].Payload)
	require.Equal(t, 0, recs[0].Attempts)
	require.True(t, recs[0].LastSentAt.IsZero())

	// Insert again with same (trigger,event) but different payload -> should overwrite per Insert behavior
	second := makePendingEvent(triggerID, eventID, []byte("second"), -30*time.Second)
	require.NoError(t, store.Insert(ctx, second))

	recs, err = store.List(ctx)
	require.NoError(t, err)
	require.Len(t, recs, 1)
	require.Equal(t, []byte("second"), recs[0].Payload)

	// UpdateDelivery: set attempts and lastSentAt
	attempts := 3
	lastSent := time.Now().UTC().Truncate(time.Second)
	require.NoError(t, store.UpdateDelivery(ctx, triggerID, eventID, lastSent, attempts))

	// Verify persisted fields
	recs, err = store.List(ctx)
	require.NoError(t, err)
	require.Len(t, recs, 1)
	require.Equal(t, attempts, recs[0].Attempts)
	require.True(t, recs[0].LastSentAt.Truncate(time.Second).Equal(lastSent))
}

func TestMemEventStore_UpdateDelivery_NotFound(t *testing.T) {
	ctx := t.Context()
	store := NewMemEventStore()

	err := store.UpdateDelivery(ctx, "no-such-trigger", "no-such-event", time.Now(), 1)
	require.Error(t, err)
}
