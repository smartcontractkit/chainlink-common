package resourcemanager

import (
	"context"
	"errors"
	"math/big"
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
	Product:         "cre-mainline",
	Tenant:          "mainline",
	NumericTenantID: "42",
	Environment:     "production",
	Zone:            "wf-zone-a",
	Don:             &DonIdentity{DonID: "don-1", NodeID: "node-1"},
	Service:         "cron-trigger",
	ResourcePool:    "trigger_registrations",
	ResourcePoolID:  "",
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

func attrMap(t *testing.T, attrKVs []any) map[string]any {
	t.Helper()
	require.Zero(t, len(attrKVs)%2)
	m := make(map[string]any, len(attrKVs)/2)
	for i := 0; i < len(attrKVs); i += 2 {
		key, ok := attrKVs[i].(string)
		require.True(t, ok)
		m[key] = attrKVs[i+1]
	}
	return m
}

func waitForSnapshotCount(t *testing.T, emitter *fakeEmitter, want int) {
	t.Helper()
	require.Eventually(t, func() bool {
		return len(decodeSnapshots(t, emitter.snapshot())) == want
	}, time.Second, time.Millisecond)
}

func TestEmitDelta_Gating(t *testing.T) {
	tests := []struct {
		name             string
		recordsEnabled   bool
		snapshotsEnabled bool
		emitter          *fakeEmitter
		wantEmits        int
	}{
		{name: "disabled does not emit", recordsEnabled: false, snapshotsEnabled: true, emitter: &fakeEmitter{}, wantEmits: 0},
		{name: "enabled emits", recordsEnabled: true, snapshotsEnabled: true, emitter: &fakeEmitter{}, wantEmits: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ResourceManagerConfig{
				MeterRecordsEnabled:   tt.recordsEnabled,
				MeterSnapshotsEnabled: tt.snapshotsEnabled,
				SnapshotInterval:      time.Minute,
			}
			if tt.emitter != nil {
				cfg.Emitter = tt.emitter
			}
			rm := NewResourceManager(logger.Test(t), cfg)
			rm.EmitDelta(t.Context(), testIdentity, "register:wf-1:trigger-1", 1, UtilizationFields{
				ResourceType: "operations",
				ResourceID:   "trigger-1",
				OrgID:        "org-1",
			})
			assert.Len(t, tt.emitter.calls, tt.wantEmits)
		})
	}
}

func TestEmitDelta_Success(t *testing.T) {
	emitter := &fakeEmitter{}
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		MeterRecordsEnabled: true,
		Emitter:             emitter,
	})

	rm.EmitDelta(t.Context(), testIdentity, "unregister:wf-1:trigger-1", -3, UtilizationFields{
		ResourceType: "operations",
		ResourceID:   "trigger-1",
		OrgID:        "org-1",
	})

	require.Len(t, emitter.calls, 1)
	call := emitter.calls[0]
	attrs := attrMap(t, call.attrKVs)
	assert.Equal(t, "metering.v1.meter_record", attrs[beholder.AttrKeyDataSchema])
	assert.Equal(t, "cll-meter", attrs[beholder.AttrKeyDomain])
	assert.Equal(t, "metering.v1.MeterRecord", attrs[beholder.AttrKeyEntity])

	var record meteringpb.MeterRecord
	require.NoError(t, proto.Unmarshal(call.body, &record))
	require.NotNil(t, record.GetIdentity())
	assert.Equal(t, testIdentity.Product, record.GetIdentity().GetProduct())
	assert.Equal(t, testIdentity.Tenant, record.GetIdentity().GetTenant())
	assert.Equal(t, testIdentity.NumericTenantID, record.GetIdentity().GetNumericTenantId())
	assert.Equal(t, testIdentity.DonID(), record.GetIdentity().GetDon().GetDonId())
	assert.Equal(t, testIdentity.NodeID(), record.GetIdentity().GetDon().GetNodeId())
	assert.Equal(t, testIdentity.ResourcePool, record.GetIdentity().GetResourcePool())
	assert.Equal(t, meteringpb.MeterAction_METER_ACTION_UPDATE, record.GetAction())
	require.Len(t, record.GetUtilizations(), 1)
	assert.Equal(t, "-3", record.GetUtilizations()[0].GetValue())
	assert.Equal(t, "operations", record.GetUtilizations()[0].GetResourceType())
	assert.Equal(t, "trigger-1", record.GetUtilizations()[0].GetResourceId())
	// event_id is the exact producer-supplied value, verbatim.
	assert.Equal(t, "unregister:wf-1:trigger-1", record.GetUtilizations()[0].GetEventId())
	assert.Equal(t, "org-1", record.GetUtilizations()[0].GetOrgId())
}

func TestEmitUsage_Action(t *testing.T) {
	emitter := &fakeEmitter{}
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		MeterRecordsEnabled: true,
		Emitter:             emitter,
	})

	rm.EmitUsage(t.Context(), testIdentity, "usage:req-1", 7, UtilizationFields{
		ResourceType: "operations",
		ResourceID:   "req-1",
		OrgID:        "org-1",
	})

	require.Len(t, emitter.calls, 1)
	var record meteringpb.MeterRecord
	require.NoError(t, proto.Unmarshal(emitter.calls[0].body, &record))
	assert.Equal(t, meteringpb.MeterAction_METER_ACTION_USAGE, record.GetAction())
	require.Len(t, record.GetUtilizations(), 1)
	assert.Equal(t, "7", record.GetUtilizations()[0].GetValue())
	assert.Equal(t, "usage:req-1", record.GetUtilizations()[0].GetEventId())
}

// TestEmitDelta_EventIDIdenticalAcrossNodes proves the core cross-node contract:
// two nodes fielding the SAME logical delta (identical eventID) emit the
// identical event_id, while a distinct request (distinct eventID) is distinct.
func TestEmitDelta_EventIDIdenticalAcrossNodes(t *testing.T) {
	emitter := &fakeEmitter{}
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		MeterRecordsEnabled: true,
		Emitter:             emitter,
	})
	fields := UtilizationFields{ResourceType: "operations", ResourceID: "trigger-1", OrgID: "org-1"}

	// Same triggering request, derived on two nodes -> identical eventID.
	sameID := EventID("register", "wf-1", "trigger-1")
	rm.EmitDelta(t.Context(), testIdentity, sameID, 1, fields)
	rm.EmitDelta(t.Context(), testIdentity, sameID, 1, fields)
	// A distinct request (e.g. re-register with a disambiguator) -> distinct.
	otherID := EventID("register", "wf-1", "trigger-1", "reactivation-2")
	rm.EmitDelta(t.Context(), testIdentity, otherID, 1, fields)

	require.Len(t, emitter.calls, 3)
	ids := make([]string, 3)
	for i := range emitter.calls {
		var r meteringpb.MeterRecord
		require.NoError(t, proto.Unmarshal(emitter.calls[i].body, &r))
		ids[i] = r.GetUtilizations()[0].GetEventId()
	}
	assert.Equal(t, sameID, ids[0])
	assert.Equal(t, ids[0], ids[1], "two nodes for the same request must emit identical event_id")
	assert.NotEqual(t, ids[0], ids[2], "distinct requests must emit distinct event_id")
}

// TestEmitDelta_EmptyEventID verifies the no-fallback rule: an empty eventID is
// logged + failure-counted and the emit is SKIPPED (no random id), while the
// call remains fail-open (no panic).
func TestEmitDelta_EmptyEventID(t *testing.T) {
	emitter := &fakeEmitter{}
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		MeterRecordsEnabled: true,
		Emitter:             emitter,
	})
	require.NotPanics(t, func() {
		rm.EmitDelta(t.Context(), testIdentity, "", 1, UtilizationFields{
			ResourceType: "operations",
			ResourceID:   "trigger-1",
			OrgID:        "org-1",
		})
	})
	assert.Empty(t, emitter.calls, "empty event_id must skip the emit, never generate a random id")
}

func TestEmitDelta_EmitFailureIsSwallowed(t *testing.T) {
	emitter := &fakeEmitter{err: errors.New("collector unavailable")}
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		MeterRecordsEnabled: true,
		Emitter:             emitter,
	})

	require.NotPanics(t, func() {
		rm.EmitDelta(t.Context(), testIdentity, "unregister:wf-1:trigger-1", -1, UtilizationFields{
			ResourceType: "operations",
			ResourceID:   "trigger-1",
			OrgID:        "org-1",
		})
	})
	require.Len(t, emitter.calls, 1)
}

func decodeSnapshots(t *testing.T, calls []emitCall) []*meteringpb.MeterSnapshot {
	t.Helper()
	var snaps []*meteringpb.MeterSnapshot
	for _, c := range calls {
		attrs := attrMap(t, c.attrKVs)
		if attrs[beholder.AttrKeyEntity] != "metering.v1.MeterSnapshot" {
			continue
		}
		// Snapshots route on the cll-meter beholder domain, same as records.
		assert.Equal(t, "cll-meter", attrs[beholder.AttrKeyDomain])
		var s meteringpb.MeterSnapshot
		require.NoError(t, proto.Unmarshal(c.body, &s))
		snaps = append(snaps, &s)
	}
	return snaps
}

func TestEmitSnapshot_OnePerResource(t *testing.T) {
	emitter := &fakeEmitter{}
	clock := clockwork.NewFakeClockAt(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		MeterRecordsEnabled:   true,
		MeterSnapshotsEnabled: true,
		Emitter:               emitter,
		SnapshotInterval:      30 * time.Second,
		Clock:                 clock,
	})

	m := &fakeMeterable{identity: testIdentity}
	m.set([]SnapshotEntry{
		{
			Identity: testIdentity,
			Utilizations: []*meteringpb.Utilization{
				NewUtilizationInt(3, UtilizationFields{
					ResourceType: "operations",
					ResourceID:   "trigger-1",
					OrgID:        "org-1",
				}),
			},
		},
		{
			Identity: testIdentity,
			Utilizations: []*meteringpb.Utilization{
				NewUtilizationInt(5, UtilizationFields{
					ResourceType: "operations",
					ResourceID:   "trigger-2",
					OrgID:        "org-2",
				}),
			},
		},
	})
	unregister := rm.Register(m)
	defer unregister()

	servicetest.Run(t, rm)
	require.NoError(t, clock.BlockUntilContext(t.Context(), 1))
	clock.Advance(30 * time.Second)
	waitForSnapshotCount(t, emitter, 2)

	snaps := decodeSnapshots(t, emitter.snapshot())
	require.Len(t, snaps, 2)

	bucketUnix := time.Date(2024, 1, 1, 0, 0, 30, 0, time.UTC).Unix()
	byID := map[string]*meteringpb.MeterSnapshot{}
	for _, s := range snaps {
		require.Len(t, s.GetUtilization(), 1)
		u := s.GetUtilization()[0]
		byID[u.GetResourceId()] = s
		// Snapshot timestamps are aligned to the interval boundary.
		assert.Equal(t, time.Date(2024, 1, 1, 0, 0, 30, 0, time.UTC), s.GetTimestamp().AsTime())
		// event_id is the RM-derived deterministic per-node/per-bucket key.
		wantID := SnapshotEventID(
			testIdentity.NodeID(), testIdentity.Service, testIdentity.ResourcePool,
			u.GetResourceType(), u.GetResourceId(), bucketUnix,
		)
		assert.Equal(t, wantID, u.GetEventId())
	}
	require.NotNil(t, byID["trigger-1"])
	assert.Equal(t, "3", byID["trigger-1"].GetUtilization()[0].GetValue())
	require.NotNil(t, byID["trigger-2"])
	assert.Equal(t, "5", byID["trigger-2"].GetUtilization()[0].GetValue())
}

func TestSnapshotsCannotRunWithoutRecords(t *testing.T) {
	emitter := &fakeEmitter{}
	clock := clockwork.NewFakeClockAt(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		MeterRecordsEnabled:   false,
		MeterSnapshotsEnabled: true,
		Emitter:               emitter,
		SnapshotInterval:      10 * time.Second,
		Clock:                 clock,
	})

	m := &fakeMeterable{identity: testIdentity}
	m.set([]SnapshotEntry{{
		Identity: testIdentity,
		Utilizations: []*meteringpb.Utilization{
			NewUtilizationInt(1, UtilizationFields{
				ResourceType: "operations",
				ResourceID:   "trigger-1",
			}),
		},
	}})
	rm.Register(m)
	servicetest.Run(t, rm)
	assert.Empty(t, decodeSnapshots(t, emitter.snapshot()))
}

func TestNewUtilizationVariants(t *testing.T) {
	fields := UtilizationFields{
		ResourceType: "operations",
		ResourceID:   "rid-1",
		OrgID:        "org-1",
	}

	uInt := NewUtilizationInt(42, fields)
	assert.Equal(t, "42", uInt.GetValue())
	assert.Equal(t, "operations", uInt.GetResourceType())
	assert.Equal(t, "rid-1", uInt.GetResourceId())
	// event_id is not a producer-supplied field; it is stamped at emit time.
	assert.Empty(t, uInt.GetEventId())
	assert.Equal(t, "org-1", uInt.GetOrgId())

	uNeg := NewUtilizationInt(-5, fields)
	assert.Equal(t, "-5", uNeg.GetValue())

	uBig := NewUtilizationBig(big.NewInt(0).Exp(big.NewInt(2), big.NewInt(80), nil), fields)
	assert.Equal(t, "1208925819614629174706176", uBig.GetValue())

	uFloat := NewUtilizationFloat(3.5, fields)
	assert.Equal(t, "3.5", uFloat.GetValue())
}

func TestEmitDelta_BeholderObserver(t *testing.T) {
	observer := beholdertest.NewObserver(t)
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		MeterRecordsEnabled: true,
		Emitter:             beholder.GetEmitter(),
	})
	rm.EmitDelta(t.Context(), testIdentity, "register:wf-1:trigger-1", 1, UtilizationFields{
		ResourceType: "operations",
		ResourceID:   "trigger-1",
		OrgID:        "org-1",
	})

	msgs := observer.Messages(t, beholder.AttrKeyEntity, "metering.v1.MeterRecord")
	require.Len(t, msgs, 1)
}
