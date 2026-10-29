package resourcemanager

import (
	"fmt"
	"math"
	"testing"

	meteringpb "github.com/smartcontractkit/chainlink-protos/metering/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

func TestSampled(t *testing.T) {
	assert.False(t, sampled(0, "a"))
	assert.False(t, sampled(math.NaN(), "a"))
	assert.False(t, sampled(-1, "a"))
	assert.True(t, sampled(1, "a"))
	assert.Equal(t, sampled(0.5, "a"), sampled(0.5, "a"), "deterministic")

	hits := 0
	for i := 0; i < 10000; i++ {
		if sampled(0.1, fmt.Sprint(i)) {
			hits++
		}
	}
	assert.InDelta(t, 1000, hits, 150, "rate is honoured")

	// Vectors shared with the reconcilers; do not change without updating them.
	for key, want := range map[string]bool{"a": true, "b": false, "c": true, "exec-1": true, "exec-2": false, "0xaa": true} {
		assert.Equal(t, want, sampled(0.5, key), key)
	}
}

// faultSettings builds a getter with PerOrg.BillingFault* overrides for org.
func faultSettings(t *testing.T, org string, kv map[string]string) settings.Getter {
	t.Helper()
	cfg := fmt.Sprintf("[org.%s.PerOrg]\n", org)
	for k, v := range kv {
		cfg += fmt.Sprintf("%s = %q\n", k, v)
	}
	g, err := settings.NewTOMLGetter([]byte(cfg))
	require.NoError(t, err)
	return g
}

func emitUsageWithFault(t *testing.T, g settings.Getter, nodeID, eventID string) *fakeEmitter {
	t.Helper()
	emitter := &fakeEmitter{}
	rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{MeterRecordsEnabled: true, Emitter: emitter, Settings: g})
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

func TestApplyFault(t *testing.T) {
	t.Run("no settings emits unchanged", func(t *testing.T) {
		assert.Equal(t, "100", emittedValue(t, emitUsageWithFault(t, nil, "n1", "exec-1")))
	})
	t.Run("defaults are off", func(t *testing.T) {
		g := faultSettings(t, "org-1", nil)
		assert.Equal(t, "100", emittedValue(t, emitUsageWithFault(t, g, "n1", "exec-1")))
	})
	t.Run("drop skips the selected record", func(t *testing.T) {
		g := faultSettings(t, "org-1", map[string]string{"BillingFaultDropRate": "1"})
		assert.Empty(t, emitUsageWithFault(t, g, "n1", "exec-1").calls)
	})
	t.Run("another org's settings do not apply", func(t *testing.T) {
		g := faultSettings(t, "org-other", map[string]string{"BillingFaultDropRate": "1"})
		assert.Equal(t, "100", emittedValue(t, emitUsageWithFault(t, g, "n1", "exec-1")))
	})
	t.Run("record without an org is never faulted", func(t *testing.T) {
		emitter := &fakeEmitter{}
		rm := NewResourceManager(logger.Test(t), ResourceManagerConfig{MeterRecordsEnabled: true, Emitter: emitter,
			Settings: faultSettings(t, "org-1", map[string]string{"BillingFaultDropRate": "1"})})
		rm.EmitUsage(t.Context(), testIdentity, "exec-1", 100, UtilizationFields{ResourceType: ResourceTypeWorkflowCompute, ResourceID: "wf-1:exec-1"})
		assert.Len(t, emitter.calls, 1)
	})
	t.Run("drop and perturb run together on independent samples", func(t *testing.T) {
		g := faultSettings(t, "org-1", map[string]string{"BillingFaultDropRate": "0.5", "BillingFaultPerturbRate": "0.5", "BillingFaultPerturbDelta": "-7"})
		// Vectors shared with the reconcilers; do not change without updating them.
		require.True(t, sampled(0.5, "drop|exec-1"))
		require.False(t, sampled(0.5, "drop|exec-17"))
		require.True(t, sampled(0.5, "perturb|exec-17"))
		require.False(t, sampled(0.5, "drop|exec-9"))
		require.False(t, sampled(0.5, "perturb|exec-9"))
		require.True(t, sampled(0.5, "sign|exec-17"), "subtracts")
		require.False(t, sampled(0.5, "sign|exec-2"), "adds")
		assert.Empty(t, emitUsageWithFault(t, g, "n1", "exec-1").calls, "in the drop sample")
		assert.Equal(t, "93", emittedValue(t, emitUsageWithFault(t, g, "n1", "exec-17")), "in the perturb sample only")
		assert.Equal(t, "100", emittedValue(t, emitUsageWithFault(t, g, "n1", "exec-9")), "in neither")
	})
	t.Run("perturb direction is per event and shared by every node", func(t *testing.T) {
		g := faultSettings(t, "org-1", map[string]string{"BillingFaultPerturbRate": "1", "BillingFaultPerturbDelta": "7"})
		assert.Equal(t, "93", emittedValue(t, emitUsageWithFault(t, g, "n1", "exec-17")))
		assert.Equal(t, "93", emittedValue(t, emitUsageWithFault(t, g, "n2", "exec-17")))
		assert.Equal(t, "107", emittedValue(t, emitUsageWithFault(t, g, "n1", "exec-2")))
		assert.Equal(t, "107", emittedValue(t, emitUsageWithFault(t, g, "n2", "exec-2")))
	})
	t.Run("perturb uses the delta magnitude", func(t *testing.T) {
		g := faultSettings(t, "org-1", map[string]string{"BillingFaultPerturbRate": "1", "BillingFaultPerturbDelta": "-7"})
		assert.Equal(t, "107", emittedValue(t, emitUsageWithFault(t, g, "n1", "exec-2")))
	})
	t.Run("perturb needs a nonzero delta", func(t *testing.T) {
		g := faultSettings(t, "org-1", map[string]string{"BillingFaultPerturbRate": "1", "BillingFaultPerturbDelta": "0"})
		assert.Equal(t, "100", emittedValue(t, emitUsageWithFault(t, g, "n1", "exec-1")))
	})
	t.Run("perturb clamps at zero", func(t *testing.T) {
		g := faultSettings(t, "org-1", map[string]string{"BillingFaultPerturbRate": "1", "BillingFaultPerturbDelta": "1000"})
		assert.Equal(t, "0", emittedValue(t, emitUsageWithFault(t, g, "n1", "exec-17")))
	})
	t.Run("perturb leaves a zero value and logs nothing", func(t *testing.T) {
		g := faultSettings(t, "org-1", map[string]string{"BillingFaultPerturbRate": "1", "BillingFaultPerturbDelta": "7"})
		emitter := &fakeEmitter{}
		lggr, observed := logger.TestObserved(t, zapcore.InfoLevel)
		rm := NewResourceManager(lggr, ResourceManagerConfig{MeterRecordsEnabled: true, Emitter: emitter, Settings: g})
		rm.EmitUsage(t.Context(), testIdentity, "exec-17", 0, UtilizationFields{ResourceType: ResourceTypeWorkflowCompute, ResourceID: "wf-1:exec-17", OrgID: "org-1"})
		assert.Equal(t, "0", emittedValue(t, emitter))
		assert.Empty(t, observed.FilterMessage("Injected usage fault for billing reconciler canary").All())
	})
	t.Run("perturb honours the node rate", func(t *testing.T) {
		g := faultSettings(t, "org-1", map[string]string{"BillingFaultPerturbRate": "1", "BillingFaultPerturbDelta": "7", "BillingFaultNodeRate": "0.5"})
		var perturbed, untouched int
		for i := 0; i < 20; i++ {
			switch emittedValue(t, emitUsageWithFault(t, g, fmt.Sprintf("node-%d", i), "exec-1")) {
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
			Settings: faultSettings(t, "org-1", map[string]string{"BillingFaultDropRate": "1"})})
		rm.EmitDelta(t.Context(), testIdentity, "delta-1", 1, UtilizationFields{ResourceType: "operations", ResourceID: "r", OrgID: "org-1"})
		assert.Len(t, emitter.calls, 1)
	})
}
