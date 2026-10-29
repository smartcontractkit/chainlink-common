package resourcemanager

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"math"
	"math/big"
	"strings"

	meteringpb "github.com/smartcontractkit/chainlink-protos/metering/go"

	"github.com/smartcontractkit/chainlink-common/pkg/contexts"
	"github.com/smartcontractkit/chainlink-common/pkg/settings/cresettings"
)

// Usage fault kinds, as logged on the "Injected usage fault for billing reconciler canary" line.
const (
	faultDrop    = "drop"
	faultPerturb = "perturb"
)

// sampled reports whether key falls in the deterministic sample for rate:
// the first 8 bytes of sha256(key) mod 10000 are below rate in basis points.
// Same arithmetic as the workflow-worker DropDecider, so one offline helper
// recomputes both injectors' samples.
func sampled(rate float64, key string) bool {
	if !(rate > 0) { // also rejects NaN
		return false
	}
	h := sha256.Sum256([]byte(key))
	return binary.BigEndian.Uint64(h[:8])%10000 < uint64(math.Round(rate*10000))
}

// applyFault drops or perturbs usage records of orgs whose CRE settings
// (PerOrg.BillingFault*) enable it, so the billing reconcilers can measure
// their recall. It returns true when the record must not be emitted. Faults
// are logged as "Injected usage fault for billing reconciler canary", a contract line the reconcilers read.
func (rm *ResourceManager) applyFault(ctx context.Context, identity ResourceIdentity, eventID string, u *meteringpb.Utilization) (drop bool) {
	if rm.settings == nil || u == nil || u.GetOrgId() == "" {
		return false
	}
	cre := contexts.CREValue(ctx)
	cre.Org = u.GetOrgId() // the billed org is the settings tenant
	ctx = contexts.WithCRE(ctx, cre)
	s := &cresettings.Default.PerOrg
	dropRate, _ := s.BillingFaultDropRate.GetOrDefault(ctx, rm.settings)
	perturbRate, _ := s.BillingFaultPerturbRate.GetOrDefault(ctx, rm.settings)
	delta, _ := s.BillingFaultPerturbDelta.GetOrDefault(ctx, rm.settings)
	nodeRate, _ := s.BillingFaultNodeRate.GetOrDefault(ctx, rm.settings)

	mode := ""
	switch {
	case sampled(dropRate, faultDrop+"|"+eventID):
		mode = faultDrop // every node drops, so the record never reaches billing
	case delta != 0 && sampled(perturbRate, faultPerturb+"|"+eventID) && sampled(nodeRate, identity.NodeID()+"|"+eventID):
		mode = faultPerturb // a sample of nodes perturbs, so nodes disagree
	default:
		return false
	}
	workflowID, executionID, _ := strings.Cut(u.GetResourceId(), ":")
	fields := []any{
		"mode", mode,
		"resourceType", u.GetResourceType(),
		"eventID", eventID,
		"resourceId", u.GetResourceId(),
		"workflowID", workflowID,
		"executionID", executionID,
		"orgID", u.GetOrgId(),
		"nodeID", identity.NodeID(),
	}
	if mode == faultDrop {
		rm.lggr.Infow("Injected usage fault for billing reconciler canary", append(fields, "value", u.GetValue())...)
		return true
	}
	original := u.GetValue()
	v, ok := new(big.Int).SetString(original, 10)
	if !ok {
		return false
	}
	d := new(big.Int).Abs(big.NewInt(delta))
	if sampled(0.5, "sign|"+eventID) { // half the events under-report, half over-report
		d.Neg(d)
	}
	v.Add(v, d)
	if v.Sign() < 0 {
		v.SetInt64(0)
	}
	if v.String() == original { // a zero clamps to itself: nothing to detect, so no log
		return false
	}
	u.Value = v.String()
	rm.lggr.Infow("Injected usage fault for billing reconciler canary", append(fields, "value", u.Value, "originalValue", original)...)
	return false
}
