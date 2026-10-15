package resourcemanager

import (
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services/servicetest"
	meteringpb "github.com/smartcontractkit/chainlink-protos/metering/go"
)

func collectUtilizationPoints(t *testing.T, reader *sdkmetric.ManualReader) []metricdata.DataPoint[int64] {
	t.Helper()
	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(t.Context(), &data))
	for _, sm := range data.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name == utilizationGaugeName {
				return m.Data.(metricdata.Gauge[int64]).DataPoints
			}
		}
	}
	return nil
}

func TestUtilizationGauge_ResourceIDLabel(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	gauge, err := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)).Meter("test").Int64Gauge(utilizationGaugeName)
	require.NoError(t, err)

	t.Run("usage records omit resource_id", func(t *testing.T) {
		rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{MeterRecordsEnabled: true, Emitter: &fakeEmitter{}})
		rm.utilization = gauge

		for _, exec := range []string{"exec-1", "exec-2"} {
			rm.EmitUsage(t.Context(), testIdentity, exec, 7, UtilizationFields{
				ResourceType: "cre:workflow:compute",
				ResourceID:   "wf-1:" + exec,
				OrgID:        "org-1",
			})
		}

		points := collectUtilizationPoints(t, reader)
		require.Len(t, points, 1)
		_, hasID := points[0].Attributes.Value("resource_id")
		assert.False(t, hasID)
		rt, _ := points[0].Attributes.Value("resource_type")
		assert.Equal(t, "cre:workflow:compute", rt.AsString())
	})

	t.Run("deltas keep resource_id", func(t *testing.T) {
		rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{MeterRecordsEnabled: true, Emitter: &fakeEmitter{}})
		rm.utilization = gauge

		for _, id := range []string{"trigger-1", "trigger-2"} {
			rm.EmitDelta(t.Context(), testIdentity, "register:wf-1:"+id, 1, UtilizationFields{
				ResourceType: "registrations",
				ResourceID:   id,
				OrgID:        "org-1",
			})
		}

		var ids []string
		for _, p := range collectUtilizationPoints(t, reader) {
			rt, _ := p.Attributes.Value("resource_type")
			if rt.AsString() != "registrations" {
				continue
			}
			id, hasID := p.Attributes.Value("resource_id")
			require.True(t, hasID)
			ids = append(ids, id.AsString())
		}
		assert.ElementsMatch(t, []string{"trigger-1", "trigger-2"}, ids)
	})

	t.Run("snapshots keep resource_id and share the delta series", func(t *testing.T) {
		emitter := &fakeEmitter{}
		clock := clockwork.NewFakeClockAt(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
		rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{
			MeterRecordsEnabled:   true,
			MeterSnapshotsEnabled: true,
			Emitter:               emitter,
			SnapshotInterval:      30 * time.Second,
			Clock:                 clock,
		})
		rm.utilization = gauge

		rm.EmitDelta(t.Context(), testIdentity, "register:wf-1:trigger-1", 1, UtilizationFields{
			ResourceType: "operations", ResourceID: "trigger-1", OrgID: "org-1",
		})
		deltaSet := attributeSetFor(t, collectUtilizationPoints(t, reader), "operations")

		m := &fakeMeterable{identity: testIdentity}
		m.set([]SnapshotEntry{{
			Identity: testIdentity,
			Utilizations: []*meteringpb.Utilization{
				NewUtilizationInt(3, UtilizationFields{ResourceType: "operations", ResourceID: "trigger-1", OrgID: "org-1"}),
			},
		}})
		unregister := rm.Register(m)
		defer unregister()

		servicetest.Run(t, rm)
		require.NoError(t, clock.BlockUntilContext(t.Context(), 1))
		clock.Advance(30 * time.Second)
		waitForSnapshotCount(t, emitter, 1)

		points := collectUtilizationPoints(t, reader)
		snapSet := attributeSetFor(t, points, "operations")
		id, hasID := snapSet.Value("resource_id")
		require.True(t, hasID)
		assert.Equal(t, "trigger-1", id.AsString())
		assert.True(t, deltaSet.Equals(&snapSet), "delta and snapshot must land on the same series")
	})
}

// attributeSetFor returns the attribute set of the single point with the given resource_type.
func attributeSetFor(t *testing.T, points []metricdata.DataPoint[int64], resourceType string) attribute.Set {
	t.Helper()
	var sets []attribute.Set
	for _, p := range points {
		if rt, _ := p.Attributes.Value("resource_type"); rt.AsString() == resourceType {
			sets = append(sets, p.Attributes)
		}
	}
	require.Len(t, sets, 1)
	return sets[0]
}
