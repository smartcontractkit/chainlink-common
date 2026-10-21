package resourcemanager

import (
	"testing"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"google.golang.org/protobuf/proto"

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

func TestUtilizationGauge_NoResourceIDLabel(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	gauge, err := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader)).Meter("test").Int64Gauge(utilizationGaugeName)
	require.NoError(t, err)

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

	for _, exec := range []string{"exec-1", "exec-2"} {
		rm.EmitUsage(t.Context(), testIdentity, exec, 7, UtilizationFields{
			ResourceType: "cre:workflow:compute", ResourceID: "wf-1:" + exec, OrgID: "org-1",
		})
	}
	for _, id := range []string{"trigger-1", "trigger-2"} {
		rm.EmitDelta(t.Context(), testIdentity, "register:wf-1:"+id, 1, UtilizationFields{
			ResourceType: "registrations", ResourceID: id, OrgID: "org-1",
		})
	}
	m := &fakeMeterable{identity: testIdentity}
	m.set([]SnapshotEntry{{
		Identity: testIdentity,
		Utilizations: []*meteringpb.Utilization{
			NewUtilizationInt(3, UtilizationFields{ResourceType: "operations", ResourceID: "wf-1", OrgID: "org-1"}),
		},
	}})
	unregister := rm.Register(m)
	defer unregister()
	servicetest.Run(t, rm)
	require.NoError(t, clock.BlockUntilContext(t.Context(), 1))
	clock.Advance(30 * time.Second)
	waitForSnapshotCount(t, emitter, 1)

	points := collectUtilizationPoints(t, reader)
	byType := map[string]int{}
	for _, p := range points {
		_, hasID := p.Attributes.Value("resource_id")
		assert.False(t, hasID, "resource_id must never be a metric label")
		rt, _ := p.Attributes.Value("resource_type")
		byType[rt.AsString()]++
	}
	// Distinct resource ids collapse onto one series per resource_type.
	assert.Equal(t, map[string]int{"cre:workflow:compute": 1, "registrations": 1, "operations": 1}, byType)

	// The durable events still carry the resource id.
	ids := map[string]bool{}
	for _, call := range emitter.snapshot() {
		var record meteringpb.MeterRecord
		if proto.Unmarshal(call.body, &record) != nil {
			continue
		}
		for _, u := range record.GetUtilizations() {
			ids[u.GetResourceId()] = true
		}
	}
	assert.True(t, ids["wf-1:exec-1"] && ids["wf-1:exec-2"] && ids["trigger-1"] && ids["trigger-2"])
}
