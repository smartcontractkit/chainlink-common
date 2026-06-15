package resourcemanager

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/beholder/beholdertest"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
	meteringpb "github.com/smartcontractkit/chainlink-protos/metering/go"
)

var testIdentity = ResourceIdentity{
	Product:      "cre-mainline",
	Environment:  "production",
	Zone:         "wf-zone-a",
	DONID:        "don-1",
	NodeID:       "node-1",
	Service:      "cron-trigger",
	Resource:     "trigger_registrations",
	ResourceType: "operations",
}

type emitCall struct {
	body    []byte
	attrKVs []any
}

type fakeEmitter struct {
	mu    sync.Mutex
	err   error
	calls []emitCall
}

func (f *fakeEmitter) Emit(_ context.Context, body []byte, attrKVs ...any) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, emitCall{body: body, attrKVs: attrKVs})
	return f.err
}

func (f *fakeEmitter) snapshot() []emitCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]emitCall(nil), f.calls...)
}

// fakeMeterable is a test Meterable whose active entries can be swapped under
// lock to exercise registration/unregistration and changing utilization.
type fakeMeterable struct {
	identity ResourceIdentity

	mu      sync.Mutex
	entries []SnapshotEntry
}

func (f *fakeMeterable) ResourceIdentity() ResourceIdentity { return f.identity }

func (f *fakeMeterable) GetUtilization(_ context.Context) []SnapshotEntry {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]SnapshotEntry(nil), f.entries...)
}

func (f *fakeMeterable) set(entries []SnapshotEntry) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.entries = entries
}

// attrMap converts beholder-style attribute key/value pairs into a map.
func attrMap(t *testing.T, attrKVs []any) map[string]any {
	t.Helper()
	require.Zero(t, len(attrKVs)%2, "attrKVs must be key/value pairs")
	m := make(map[string]any, len(attrKVs)/2)
	for i := 0; i < len(attrKVs); i += 2 {
		key, ok := attrKVs[i].(string)
		require.True(t, ok, "attr key must be a string")
		m[key] = attrKVs[i+1]
	}
	return m
}

// waitForSnapshotCount polls until the emitter has recorded want snapshot emissions.
func waitForSnapshotCount(t *testing.T, emitter *fakeEmitter, want int) {
	t.Helper()
	require.Eventually(t, func() bool {
		return len(decodeSnapshots(t, emitter.snapshot())) == want
	}, time.Second, time.Millisecond)
}

func TestEmitMeterRecord_Gating(t *testing.T) {
	tests := []struct {
		name      string
		enabled   bool
		emitter   *fakeEmitter
		wantEmits int
	}{
		{name: "disabled does not emit", enabled: false, emitter: &fakeEmitter{}, wantEmits: 0},
		{name: "disabled with nil emitter does not panic", enabled: false, emitter: nil, wantEmits: 0},
		{name: "enabled with nil emitter does not panic", enabled: true, emitter: nil, wantEmits: 0},
		{name: "enabled emits", enabled: true, emitter: &fakeEmitter{}, wantEmits: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ResourceManagerConfig{Enabled: tt.enabled}
			if tt.emitter != nil {
				cfg.Emitter = tt.emitter
			}
			rm := NewResourceManager(logger.Test(t), cfg)

			rm.EmitMeterRecord(t.Context(), testIdentity, meteringpb.MeterAction_METER_ACTION_RESERVE,
				NewUtilization(testIdentity.WithResourceID("trigger-1"), meteringpb.MeterAction_METER_ACTION_RESERVE, 1, "event-1"))

			if tt.emitter != nil {
				assert.Len(t, tt.emitter.calls, tt.wantEmits)
			}
		})
	}
}

func TestEmitMeterRecord_Success(t *testing.T) {
	emitter := &fakeEmitter{}
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{Enabled: true, Emitter: emitter})

	id := testIdentity.WithResourceID("trigger-1")
	utilization := NewUtilization(id, meteringpb.MeterAction_METER_ACTION_RESERVE, 1, "event-1")
	before := time.Now()
	rm.EmitMeterRecord(t.Context(), id, meteringpb.MeterAction_METER_ACTION_RESERVE, utilization)
	after := time.Now()

	require.Len(t, emitter.calls, 1)
	call := emitter.calls[0]

	attrs := attrMap(t, call.attrKVs)
	assert.Equal(t, "metering.v1.meter_record", attrs[beholder.AttrKeyDataSchema])
	assert.Equal(t, "platform", attrs[beholder.AttrKeyDomain])
	assert.Equal(t, "metering.v1.MeterRecord", attrs[beholder.AttrKeyEntity])

	var record meteringpb.MeterRecord
	require.NoError(t, proto.Unmarshal(call.body, &record))
	require.NotNil(t, record.GetIdentity())
	assert.Equal(t, id.Product, record.GetIdentity().GetProduct())
	assert.Equal(t, id.NodeID, record.GetIdentity().GetNodeId())
	assert.Equal(t, id.Service, record.GetIdentity().GetService())
	assert.Equal(t, id.Resource, record.GetIdentity().GetResource())
	assert.Equal(t, id.ResourceType, record.GetIdentity().GetResourceType())
	assert.Equal(t, "trigger-1", record.GetIdentity().GetResourceId())
	assert.Equal(t, meteringpb.MeterAction_METER_ACTION_RESERVE, record.GetAction())

	require.NotNil(t, record.GetUtilization())
	assert.Equal(t, int64(1), record.GetUtilization().GetValue())
	assert.Equal(t,
		IdempotencyKey(id, meteringpb.MeterAction_METER_ACTION_RESERVE, "event-1"),
		record.GetUtilization().GetIdempotencyKey())

	require.NotNil(t, record.GetTimestamp())
	ts := record.GetTimestamp().AsTime()
	assert.False(t, ts.Before(before.Add(-time.Second)), "timestamp too early: %s", ts)
	assert.False(t, ts.After(after.Add(time.Second)), "timestamp too late: %s", ts)
}

func TestEmitMeterRecord_EmitFailureIsSwallowed(t *testing.T) {
	emitter := &fakeEmitter{err: errors.New("collector unavailable")}
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{Enabled: true, Emitter: emitter})

	id := testIdentity.WithResourceID("trigger-1")
	require.NotPanics(t, func() {
		rm.EmitMeterRecord(t.Context(), id, meteringpb.MeterAction_METER_ACTION_RELEASE,
			NewUtilization(id, meteringpb.MeterAction_METER_ACTION_RELEASE, 1, "event-2"))
	})
	require.Len(t, emitter.calls, 1)

	// A subsequent emission still goes through; the failure left no state behind.
	rm.EmitMeterRecord(t.Context(), id, meteringpb.MeterAction_METER_ACTION_RELEASE,
		NewUtilization(id, meteringpb.MeterAction_METER_ACTION_RELEASE, 1, "event-3"))
	require.Len(t, emitter.calls, 2)
}

// TestEmitMeterRecord_BeholderObserver wires the manager to the global
// beholder emitter the way production does. beholdertest.NewObserver is not
// parallel-safe, so this test must not run in parallel.
func TestEmitMeterRecord_BeholderObserver(t *testing.T) {
	observer := beholdertest.NewObserver(t)

	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		Enabled: true,
		Emitter: beholder.GetEmitter(),
	})

	id := testIdentity.WithResourceID("trigger-1")
	rm.EmitMeterRecord(t.Context(), id, meteringpb.MeterAction_METER_ACTION_RESERVE,
		NewUtilization(id, meteringpb.MeterAction_METER_ACTION_RESERVE, 1, "event-1"))

	msgs := observer.Messages(t, beholder.AttrKeyEntity, "metering.v1.MeterRecord")
	require.Len(t, msgs, 1)
	assert.Equal(t, "metering.v1.meter_record", msgs[0].Attrs[beholder.AttrKeyDataSchema])
	assert.Equal(t, "platform", msgs[0].Attrs[beholder.AttrKeyDomain])

	var record meteringpb.MeterRecord
	require.NoError(t, proto.Unmarshal(msgs[0].Body, &record))
	assert.Equal(t, id.Service, record.GetIdentity().GetService())
}

// TestResourceManager_StartClose asserts the manager starts and closes cleanly
// both with snapshots enabled and with the loop disabled (zero interval).
func TestResourceManager_StartClose(t *testing.T) {
	t.Run("snapshot loop enabled", func(t *testing.T) {
		rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
			Enabled:          true,
			Emitter:          &fakeEmitter{},
			SnapshotInterval: time.Hour,
		})
		require.NoError(t, rm.Start(t.Context()))
		require.NoError(t, rm.Close())
	})

	t.Run("snapshot loop disabled by zero interval", func(t *testing.T) {
		rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
			Enabled: true,
			Emitter: &fakeEmitter{},
			// SnapshotInterval unset -> disabled.
		})
		require.NoError(t, rm.Start(t.Context()))
		require.NoError(t, rm.Close())
	})
}

// decodeSnapshots extracts every MeterSnapshot from the emitter's recorded
// calls.
func decodeSnapshots(t *testing.T, calls []emitCall) []*meteringpb.MeterSnapshot {
	t.Helper()
	var snaps []*meteringpb.MeterSnapshot
	for _, c := range calls {
		attrs := attrMap(t, c.attrKVs)
		if attrs[beholder.AttrKeyEntity] != "metering.v1.MeterSnapshot" {
			continue
		}
		assert.Equal(t, "metering.v1.meter_snapshot", attrs[beholder.AttrKeyDataSchema])
		assert.Equal(t, "platform", attrs[beholder.AttrKeyDomain])
		var s meteringpb.MeterSnapshot
		require.NoError(t, proto.Unmarshal(c.body, &s))
		snaps = append(snaps, &s)
	}
	return snaps
}

// snapshotKey recomputes the expected idempotency key for a decoded snapshot
// from its own timestamp and interval, so the assertion does not depend on the
// wall clock at emit time.
func snapshotKey(s *meteringpb.MeterSnapshot, id ResourceIdentity) string {
	bucket := s.GetTimestamp().AsTime().Truncate(s.GetInterval().AsDuration()).Unix()
	return SnapshotIdempotencyKey(id, bucket)
}

func TestEmitSnapshot_OnePerResource(t *testing.T) {
	emitter := &fakeEmitter{}
	clock := clockwork.NewFakeClockAt(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		Enabled:          true,
		Emitter:          emitter,
		SnapshotInterval: 30 * time.Second,
		Clock:            clock,
	})

	m := &fakeMeterable{identity: testIdentity}
	id1 := testIdentity.WithResourceID("trigger-1")
	id2 := testIdentity.WithResourceID("trigger-2")
	m.set([]SnapshotEntry{
		{Identity: id1, Value: 3},
		{Identity: id2, Value: 5},
	})
	unregister := rm.Register(m)
	defer unregister()

	servicetest.Run(t, rm)
	require.NoError(t, clock.BlockUntilContext(t.Context(), 1))
	clock.Advance(30 * time.Second)
	waitForSnapshotCount(t, emitter, 2)

	// Exactly one MeterSnapshot per active resource — never a batch.
	snaps := decodeSnapshots(t, emitter.snapshot())
	require.Len(t, snaps, 2)

	byID := map[string]*meteringpb.MeterSnapshot{}
	for _, s := range snaps {
		// Each snapshot carries the full per-resource identity.
		require.NotNil(t, s.GetIdentity())
		assert.Equal(t, testIdentity.Service, s.GetIdentity().GetService())
		require.NotNil(t, s.GetInterval())
		assert.Equal(t, 30*time.Second, s.GetInterval().AsDuration())
		require.NotNil(t, s.GetTimestamp())
		byID[s.GetIdentity().GetResourceId()] = s
	}

	s1 := byID["trigger-1"]
	require.NotNil(t, s1)
	assert.Equal(t, int64(3), s1.GetUtilization().GetValue())
	assert.Equal(t, snapshotKey(s1, id1), s1.GetUtilization().GetIdempotencyKey())

	s2 := byID["trigger-2"]
	require.NotNil(t, s2)
	assert.Equal(t, int64(5), s2.GetUtilization().GetValue())
	assert.Equal(t, snapshotKey(s2, id2), s2.GetUtilization().GetIdempotencyKey())
}

func TestEmitSnapshot_EmptyListEmitsNothing(t *testing.T) {
	emitter := &fakeEmitter{}
	clock := clockwork.NewFakeClockAt(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		Enabled:          true,
		Emitter:          emitter,
		SnapshotInterval: time.Minute,
		Clock:            clock,
	})

	m := &fakeMeterable{identity: testIdentity} // no active entries
	unregister := rm.Register(m)
	defer unregister()

	servicetest.Run(t, rm)

	// Nothing active -> no snapshots. Billing zeroes the resource out by its
	// absence from subsequent snapshots, not by an explicit empty record.
	assert.Empty(t, decodeSnapshots(t, emitter.snapshot()))
}

func TestEmitSnapshots_KeyStableWithinInterval(t *testing.T) {
	emitter := &fakeEmitter{}
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		Enabled:          true,
		Emitter:          emitter,
		SnapshotInterval: time.Hour, // wide bucket so all ticks share one interval
	})

	m := &fakeMeterable{identity: testIdentity}
	m.set([]SnapshotEntry{{Identity: testIdentity.WithResourceID("trigger-1"), Value: 1}})
	unregister := rm.Register(m)
	defer unregister()

	// Drive three ticks directly (no timer) to keep the test deterministic.
	rm.emitSnapshots(t.Context())
	rm.emitSnapshots(t.Context())
	rm.emitSnapshots(t.Context())

	snaps := decodeSnapshots(t, emitter.snapshot())
	require.Len(t, snaps, 3, "one MeterSnapshot per resource per tick")

	// Same resource, same interval bucket -> identical keys (retries dedup);
	// successive intervals would differ.
	k0 := snaps[0].GetUtilization().GetIdempotencyKey()
	assert.Equal(t, k0, snaps[1].GetUtilization().GetIdempotencyKey())
	assert.Equal(t, k0, snaps[2].GetUtilization().GetIdempotencyKey())
}

func TestRegister_UnregisterStopsSnapshots(t *testing.T) {
	emitter := &fakeEmitter{}
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		Enabled:          true,
		Emitter:          emitter,
		SnapshotInterval: time.Minute,
	})

	m := &fakeMeterable{identity: testIdentity}
	m.set([]SnapshotEntry{{Identity: testIdentity.WithResourceID("trigger-1"), Value: 1}})
	unregister := rm.Register(m)

	rm.emitSnapshots(t.Context())
	require.Len(t, decodeSnapshots(t, emitter.snapshot()), 1)

	unregister()
	// Idempotent: a second call is a no-op.
	require.NotPanics(t, unregister)

	rm.emitSnapshots(t.Context())
	require.Len(t, decodeSnapshots(t, emitter.snapshot()), 1, "unregistered Meterable must not be snapshotted")
}

func TestEmitSnapshot_FailureIsSwallowed(t *testing.T) {
	emitter := &fakeEmitter{err: errors.New("collector unavailable")}
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		Enabled:          true,
		Emitter:          emitter,
		SnapshotInterval: time.Minute,
	})

	m := &fakeMeterable{identity: testIdentity}
	m.set([]SnapshotEntry{{Identity: testIdentity.WithResourceID("trigger-1"), Value: 1}})
	unregister := rm.Register(m)
	defer unregister()

	require.NotPanics(t, func() { rm.emitSnapshots(t.Context()) })
	require.Len(t, emitter.snapshot(), 1)
}

// TestEmitSnapshot_BeholderObserver wires snapshots through the global beholder
// emitter. beholdertest.NewObserver is not parallel-safe; do not run this in
// parallel.
func TestEmitSnapshot_BeholderObserver(t *testing.T) {
	observer := beholdertest.NewObserver(t)
	clock := clockwork.NewFakeClockAt(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))

	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		Enabled:          true,
		Emitter:          beholder.GetEmitter(),
		SnapshotInterval: time.Minute,
		Clock:            clock,
	})

	m := &fakeMeterable{identity: testIdentity}
	m.set([]SnapshotEntry{{Identity: testIdentity.WithResourceID("trigger-1"), Value: 7}})
	unregister := rm.Register(m)
	defer unregister()

	servicetest.Run(t, rm)
	require.NoError(t, clock.BlockUntilContext(t.Context(), 1))
	clock.Advance(time.Minute)

	require.Eventually(t, func() bool {
		return len(observer.Messages(t, beholder.AttrKeyEntity, "metering.v1.MeterSnapshot")) == 1
	}, time.Second, time.Millisecond)

	msgs := observer.Messages(t, beholder.AttrKeyEntity, "metering.v1.MeterSnapshot")
	require.Len(t, msgs, 1)
	assert.Equal(t, "metering.v1.meter_snapshot", msgs[0].Attrs[beholder.AttrKeyDataSchema])

	var s meteringpb.MeterSnapshot
	require.NoError(t, proto.Unmarshal(msgs[0].Body, &s))
	assert.Equal(t, "trigger-1", s.GetIdentity().GetResourceId())
	assert.Equal(t, int64(7), s.GetUtilization().GetValue())
}

func TestIdempotencyKey(t *testing.T) {
	id := testIdentity.WithResourceID("trigger-1")
	base := IdempotencyKey(id, meteringpb.MeterAction_METER_ACTION_RESERVE, "event-1")

	t.Run("deterministic", func(t *testing.T) {
		assert.Equal(t, base, IdempotencyKey(id, meteringpb.MeterAction_METER_ACTION_RESERVE, "event-1"))
	})

	t.Run("format is sha256 hex", func(t *testing.T) {
		assert.Regexp(t, "^[0-9a-f]{64}$", base)
	})

	otherNode := id
	otherNode.NodeID = "node-2"
	otherResource := id
	otherResource.Resource = "log_filters"

	distinct := []struct {
		name string
		key  string
	}{
		{"action", IdempotencyKey(id, meteringpb.MeterAction_METER_ACTION_RELEASE, "event-1")},
		{"node_id", IdempotencyKey(otherNode, meteringpb.MeterAction_METER_ACTION_RESERVE, "event-1")},
		{"resource", IdempotencyKey(otherResource, meteringpb.MeterAction_METER_ACTION_RESERVE, "event-1")},
		{"resource ID", IdempotencyKey(id.WithResourceID("trigger-2"), meteringpb.MeterAction_METER_ACTION_RESERVE, "event-1")},
		{"event identity", IdempotencyKey(id, meteringpb.MeterAction_METER_ACTION_RESERVE, "event-2")},
	}
	for _, tt := range distinct {
		t.Run("distinct across "+tt.name, func(t *testing.T) {
			assert.NotEqual(t, base, tt.key)
		})
	}
}

func TestSnapshotIdempotencyKey(t *testing.T) {
	id := testIdentity.WithResourceID("trigger-1")
	base := SnapshotIdempotencyKey(id, 100)

	t.Run("format is sha256 hex", func(t *testing.T) {
		assert.Regexp(t, "^[0-9a-f]{64}$", base)
	})

	t.Run("stable within an interval bucket", func(t *testing.T) {
		assert.Equal(t, base, SnapshotIdempotencyKey(id, 100))
	})

	t.Run("differs per interval bucket", func(t *testing.T) {
		assert.NotEqual(t, base, SnapshotIdempotencyKey(id, 101))
	})

	t.Run("differs per resource_id", func(t *testing.T) {
		assert.NotEqual(t, base, SnapshotIdempotencyKey(id.WithResourceID("trigger-2"), 100))
	})

	t.Run("differs per node_id", func(t *testing.T) {
		other := id
		other.NodeID = "node-2"
		assert.NotEqual(t, base, SnapshotIdempotencyKey(other, 100))
	})

	t.Run("differs from a MeterRecord key", func(t *testing.T) {
		assert.NotEqual(t, base, IdempotencyKey(id, meteringpb.MeterAction_METER_ACTION_RESERVE, "1"))
	})
}

func TestNewUtilization(t *testing.T) {
	id := testIdentity.WithResourceID("filter-1")
	u := NewUtilization(id, meteringpb.MeterAction_METER_ACTION_UPDATE, 42, "event-9")

	assert.Equal(t, int64(42), u.GetValue())
	assert.Equal(t,
		IdempotencyKey(id, meteringpb.MeterAction_METER_ACTION_UPDATE, "event-9"),
		u.GetIdempotencyKey())
}

func TestResourceIdentity_WithResourceID(t *testing.T) {
	got := testIdentity.WithResourceID("rid-1")
	assert.Equal(t, "rid-1", got.ResourceID)
	assert.Empty(t, testIdentity.ResourceID, "receiver must be unchanged")
	// All other fields are copied through.
	assert.Equal(t, testIdentity.Service, got.Service)
	assert.Equal(t, testIdentity.NodeID, got.NodeID)
}
