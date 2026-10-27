package resourcemanager

import (
	"fmt"
	"math"
	"testing"

	meteringpb "github.com/smartcontractkit/chainlink-protos/metering/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

func TestFaultSelected(t *testing.T) {
	assert.False(t, FaultSelected(0, "a"))
	assert.True(t, FaultSelected(1, "a"))
	assert.Equal(t, FaultSelected(0.5, "a"), FaultSelected(0.5, "a"), "deterministic")

	hits := 0
	for i := 0; i < 10000; i++ {
		if FaultSelected(0.1, fmt.Sprint(i)) {
			hits++
		}
	}
	assert.InDelta(t, 1000, hits, 150, "rate is honoured")

	// Vectors shared with the reconcilers; do not change without updating them.
	for key, want := range map[string]bool{"a": false, "b": true, "c": true, "exec-1": false, "exec-2": true, "0xaa": false} {
		assert.Equal(t, want, FaultSelected(0.5, key), key)
	}
	assert.True(t, FaultSelected(0.5, "n2|exec-2"))
	assert.False(t, FaultSelected(0.5, "n1|exec-2"))
}

func TestUsageFaultInjection_Validate(t *testing.T) {
	orgs := []string{"org-canary"}
	assert.NoError(t, UsageFaultInjection{}.Validate())
	assert.NoError(t, UsageFaultInjection{Orgs: orgs, DropRate: 0.01, PerturbRate: 0.1, Delta: -1, NodeRate: 0.5}.Validate())
	assert.NoError(t, UsageFaultInjection{Orgs: orgs, DropRate: 0.1}.Validate())
	assert.Error(t, UsageFaultInjection{Orgs: orgs, DropRate: 2}.Validate())
	assert.Error(t, UsageFaultInjection{Orgs: orgs, PerturbRate: -0.1, Delta: -1}.Validate())
	assert.Error(t, UsageFaultInjection{Orgs: orgs, DropRate: 0.1, NodeRate: -1}.Validate())
	assert.Error(t, UsageFaultInjection{Orgs: orgs, DropRate: math.NaN()}.Validate(), "NaN drop rate")
	assert.Error(t, UsageFaultInjection{Orgs: orgs, PerturbRate: math.NaN(), Delta: -1}.Validate(), "NaN perturb rate")
	assert.Error(t, UsageFaultInjection{Orgs: orgs, DropRate: 0.1, NodeRate: math.NaN()}.Validate(), "NaN node rate")
	assert.Error(t, UsageFaultInjection{DropRate: 0.1}.Validate(), "no org allowlist")
	assert.Error(t, UsageFaultInjection{Orgs: []string{" "}, DropRate: 0.1}.Validate(), "blank org allowlist")
	assert.Error(t, UsageFaultInjection{Orgs: orgs, PerturbRate: 0.1, Delta: 1}.Validate(), "positive delta")
	assert.Error(t, UsageFaultInjection{Orgs: orgs, PerturbRate: 0.1, Delta: 0}.Validate(), "zero delta")
}

func TestUsageFaultInjection_Selection(t *testing.T) {
	f := UsageFaultInjection{Orgs: []string{"org-canary"}, DropRate: 0.5, PerturbRate: 0.5, Delta: -1}
	// Vectors shared with the reconcilers; do not change without updating them.
	assert.False(t, f.DropSelected("a"))
	assert.True(t, f.PerturbSelected("a"))
	assert.True(t, f.DropSelected("exec-2"))
	assert.False(t, f.PerturbSelected("exec-2"), "an event in both samples is dropped")
	assert.False(t, f.DropSelected("exec-3"))
	assert.False(t, f.PerturbSelected("exec-3"))
	assert.True(t, f.DropSelected("0xabc123"))
	assert.False(t, f.PerturbSelected("0xabc123"))
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
	orgs := []string{"org-1"}
	t.Run("off emits unchanged", func(t *testing.T) {
		assert.Equal(t, "100", emittedValue(t, emitUsageWithFault(t, UsageFaultInjection{}, "n1", "exec-1")))
	})
	t.Run("drop skips the selected record", func(t *testing.T) {
		assert.Empty(t, emitUsageWithFault(t, UsageFaultInjection{Orgs: orgs, DropRate: 1}, "n1", "exec-1").calls)
	})
	t.Run("unlisted org is never touched", func(t *testing.T) {
		assert.Equal(t, "100", emittedValue(t, emitUsageWithFault(t, UsageFaultInjection{Orgs: []string{"org-other"}, DropRate: 1}, "n1", "exec-1")))
		assert.Equal(t, "100", emittedValue(t, emitUsageWithFault(t, UsageFaultInjection{DropRate: 1}, "n1", "exec-1")), "no allowlist")
	})
	t.Run("unselected event is untouched", func(t *testing.T) {
		f := UsageFaultInjection{Orgs: orgs, DropRate: 0.5}
		require.False(t, f.DropSelected("a"))
		assert.Equal(t, "100", emittedValue(t, emitUsageWithFault(t, f, "n1", "a")))
	})
	t.Run("drop and perturb run together on independent samples", func(t *testing.T) {
		f := UsageFaultInjection{Orgs: orgs, DropRate: 0.5, PerturbRate: 0.5, Delta: -7}
		assert.Empty(t, emitUsageWithFault(t, f, "n1", "exec-2").calls, "in the drop sample")
		assert.Equal(t, "93", emittedValue(t, emitUsageWithFault(t, f, "n1", "a")), "in the perturb sample only")
		assert.Equal(t, "100", emittedValue(t, emitUsageWithFault(t, f, "n1", "exec-3")), "in neither")
	})
	t.Run("perturb subtracts delta on every node by default", func(t *testing.T) {
		assert.Equal(t, "93", emittedValue(t, emitUsageWithFault(t, UsageFaultInjection{Orgs: orgs, PerturbRate: 1, Delta: -7}, "n1", "exec-1")))
		assert.Equal(t, "93", emittedValue(t, emitUsageWithFault(t, UsageFaultInjection{Orgs: orgs, PerturbRate: 1, Delta: -7}, "n2", "exec-1")))
	})
	t.Run("perturb clamps at zero", func(t *testing.T) {
		assert.Equal(t, "0", emittedValue(t, emitUsageWithFault(t, UsageFaultInjection{Orgs: orgs, PerturbRate: 1, Delta: -1000}, "n1", "exec-1")))
	})
	t.Run("perturb honours the node rate", func(t *testing.T) {
		f := UsageFaultInjection{Orgs: orgs, PerturbRate: 1, Delta: -7, NodeRate: 0.5}
		var perturbed, untouched int
		for i := 0; i < 20; i++ {
			switch emittedValue(t, emitUsageWithFault(t, f, fmt.Sprintf("node-%d", i), "exec-1")) {
			case "93":
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
			FaultInjection: UsageFaultInjection{Orgs: orgs, DropRate: 1}})
		rm.EmitDelta(t.Context(), testIdentity, "delta-1", 1, UtilizationFields{ResourceType: "operations", ResourceID: "r"})
		assert.Len(t, emitter.calls, 1)
	})
}
