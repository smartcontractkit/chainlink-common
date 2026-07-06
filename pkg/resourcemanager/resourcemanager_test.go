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

func TestEmitMeterRecord_Gating(t *testing.T) {
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
			u := NewUtilizationInt(1, UtilizationFields{
				ResourceType: "operations",
				ResourceID:   "trigger-1",
				EventID:      "event-1",
				OrgID:        "org-1",
			})
			rm.EmitMeterRecord(t.Context(), testIdentity, meteringpb.MeterAction_METER_ACTION_RESERVE, []*meteringpb.Utilization{u})
			assert.Len(t, tt.emitter.calls, tt.wantEmits)
		})
	}
}

func TestEmitMeterRecord_Success(t *testing.T) {
	emitter := &fakeEmitter{}
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		MeterRecordsEnabled: true,
		Emitter:             emitter,
	})

	u := NewUtilizationInt(1, UtilizationFields{
		ResourceType: "operations",
		ResourceID:   "trigger-1",
		EventID:      "event-1",
		OrgID:        "org-1",
	})
	rm.EmitMeterRecord(t.Context(), testIdentity, meteringpb.MeterAction_METER_ACTION_RESERVE, []*meteringpb.Utilization{u})

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
	assert.Equal(t, meteringpb.MeterAction_METER_ACTION_RESERVE, record.GetAction())
	require.Len(t, record.GetUtilizations(), 1)
	assert.Equal(t, "1", record.GetUtilizations()[0].GetValue())
	assert.Equal(t, "operations", record.GetUtilizations()[0].GetResourceType())
	assert.Equal(t, "trigger-1", record.GetUtilizations()[0].GetResourceId())
	assert.Equal(t, "event-1", record.GetUtilizations()[0].GetEventId())
	assert.Equal(t, "org-1", record.GetUtilizations()[0].GetOrgId())
}

func TestEmitMeterRecord_EmitFailureIsSwallowed(t *testing.T) {
	emitter := &fakeEmitter{err: errors.New("collector unavailable")}
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		MeterRecordsEnabled: true,
		Emitter:             emitter,
	})

	u := NewUtilizationInt(1, UtilizationFields{
		ResourceType: "operations",
		ResourceID:   "trigger-1",
		EventID:      "event-2",
		OrgID:        "org-1",
	})
	require.NotPanics(t, func() {
		rm.EmitMeterRecord(t.Context(), testIdentity, meteringpb.MeterAction_METER_ACTION_RELEASE, []*meteringpb.Utilization{u})
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

	byID := map[string]*meteringpb.MeterSnapshot{}
	for _, s := range snaps {
		require.Len(t, s.GetUtilization(), 1)
		byID[s.GetUtilization()[0].GetResourceId()] = s
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
		EventID:      "evt-1",
		OrgID:        "org-1",
	}

	uInt := NewUtilizationInt(42, fields)
	assert.Equal(t, "42", uInt.GetValue())
	assert.Equal(t, "operations", uInt.GetResourceType())
	assert.Equal(t, "rid-1", uInt.GetResourceId())
	assert.Equal(t, "evt-1", uInt.GetEventId())
	assert.Equal(t, "org-1", uInt.GetOrgId())

	uBig := NewUtilizationBig(big.NewInt(0).Exp(big.NewInt(2), big.NewInt(80), nil), fields)
	assert.Equal(t, "1208925819614629174706176", uBig.GetValue())

	uFloat := NewUtilizationFloat(3.5, fields)
	assert.Equal(t, "3.5", uFloat.GetValue())
}

func TestEmitMeterRecord_BeholderObserver(t *testing.T) {
	observer := beholdertest.NewObserver(t)
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
		MeterRecordsEnabled: true,
		Emitter:             beholder.GetEmitter(),
	})
	u := NewUtilizationInt(1, UtilizationFields{
		ResourceType: "operations",
		ResourceID:   "trigger-1",
		EventID:      "event-1",
		OrgID:        "org-1",
	})
	rm.EmitMeterRecord(t.Context(), testIdentity, meteringpb.MeterAction_METER_ACTION_RESERVE, []*meteringpb.Utilization{u})

	msgs := observer.Messages(t, beholder.AttrKeyEntity, "metering.v1.MeterRecord")
	require.Len(t, msgs, 1)
}
