package resourcemanager

import (
	"fmt"
	"testing"

	meteringpb "github.com/smartcontractkit/chainlink-protos/metering/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestFaultSelected(t *testing.T) {
	assert.False(t, FaultSelected("s", 0, "a"))
	assert.True(t, FaultSelected("s", 1, "a"))
	assert.Equal(t, FaultSelected("s", 0.5, "a"), FaultSelected("s", 0.5, "a"), "deterministic")

	hits := 0
	for i := 0; i < 10000; i++ {
		if FaultSelected("s", 0.1, fmt.Sprint(i)) {
			hits++
		}
	}
	assert.InDelta(t, 1000, hits, 150, "rate is honoured")

	// Vectors shared with the reconcilers; do not change without updating them.
	for key, want := range map[string]bool{"a": false, "b": true, "c": false, "exec-1": true, "0xaa": true} {
		assert.Equal(t, want, FaultSelected("s", 0.5, key), key)
	}
	assert.False(t, FaultSelected("s|n1", 0.5, "exec-1"))
	assert.True(t, FaultSelected("s|n2", 0.5, "exec-1"))
}

func TestUsageFaultInjection_Validate(t *testing.T) {
	assert.NoError(t, UsageFaultInjection{}.Validate())
	assert.NoError(t, UsageFaultInjection{Mode: FaultModePerturb, Rate: 0.1, Delta: 1, NodeRate: 0.5}.Validate())
	assert.Error(t, UsageFaultInjection{Mode: "x", Rate: 0.1}.Validate())
	assert.Error(t, UsageFaultInjection{Mode: FaultModeDrop, Rate: 2}.Validate())
	assert.Error(t, UsageFaultInjection{Rate: 0.1}.Validate())
	assert.Error(t, UsageFaultInjection{Mode: FaultModeDrop, Rate: 0.1, NodeRate: -1}.Validate())
}

func emitUsageWithFault(t *testing.T, f UsageFaultInjection, nodeID, eventID string) *fakeEmitter {
	t.Helper()
	emitter := &fakeEmitter{}
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{MeterRecordsEnabled: true, Emitter: emitter, FaultInjection: f})
	id := testIdentity
	id.Don = &DonIdentity{DonID: "don-1", NodeID: nodeID}
	rm.EmitUsage(t.Context(), id, eventID, 100, UtilizationFields{
		ResourceType: ResourceTypeWorkflowCompute, ResourceID: "wf-1:" + eventID, OrgID: "org-1",
	})
	return emitter
}

func emittedValue(t *testing.T, emitter *fakeEmitter) string {
	t.Helper()
	require.Len(t, emitter.calls, 1)
	var record meteringpb.MeterRecord
	require.NoError(t, proto.Unmarshal(emitter.calls[0].body, &record))
	return record.GetUtilizations()[0].GetValue()
}

func TestUsageFaultInjection_Apply(t *testing.T) {
	t.Run("off emits unchanged", func(t *testing.T) {
		assert.Equal(t, "100", emittedValue(t, emitUsageWithFault(t, UsageFaultInjection{}, "n1", "exec-1")))
	})
	t.Run("drop skips the selected record", func(t *testing.T) {
		assert.Empty(t, emitUsageWithFault(t, UsageFaultInjection{Mode: FaultModeDrop, Rate: 1}, "n1", "exec-1").calls)
	})
	t.Run("unselected event is untouched", func(t *testing.T) {
		f := UsageFaultInjection{Mode: FaultModeDrop, Rate: 0.5, Salt: "s"}
		require.False(t, f.EventSelected("a"))
		assert.Equal(t, "100", emittedValue(t, emitUsageWithFault(t, f, "n1", "a")))
	})
	t.Run("perturb adds delta on every node by default", func(t *testing.T) {
		assert.Equal(t, "107", emittedValue(t, emitUsageWithFault(t, UsageFaultInjection{Mode: FaultModePerturb, Rate: 1, Delta: 7}, "n1", "exec-1")))
		assert.Equal(t, "93", emittedValue(t, emitUsageWithFault(t, UsageFaultInjection{Mode: FaultModePerturb, Rate: 1, Delta: -7}, "n2", "exec-1")))
	})
	t.Run("perturb honours the node rate", func(t *testing.T) {
		f := UsageFaultInjection{Mode: FaultModePerturb, Rate: 1, Delta: 7, Salt: "s", NodeRate: 0.5}
		var perturbed, untouched int
		for i := 0; i < 20; i++ {
			switch emittedValue(t, emitUsageWithFault(t, f, fmt.Sprintf("node-%d", i), "exec-1")) {
			case "107":
				perturbed++
			case "100":
				untouched++
			}
		}
		assert.Positive(t, perturbed)
		assert.Positive(t, untouched)
		assert.Equal(t, 20, perturbed+untouched)
	})
	t.Run("durable deltas are never faulted", func(t *testing.T) {
		emitter := &fakeEmitter{}
		rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{MeterRecordsEnabled: true, Emitter: emitter,
			FaultInjection: UsageFaultInjection{Mode: FaultModeDrop, Rate: 1}})
		rm.EmitDelta(t.Context(), testIdentity, "delta-1", 1, UtilizationFields{ResourceType: "operations", ResourceID: "r"})
		assert.Len(t, emitter.calls, 1)
	})
}
