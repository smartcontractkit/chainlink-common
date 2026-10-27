package resourcemanager

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"math/big"
	"strings"

	meteringpb "github.com/smartcontractkit/chainlink-protos/metering/go"
)

// Usage fault kinds, as logged on the "Injected usage fault" line.
const (
	// FaultModeDrop skips emission of selected usage records on every node.
	FaultModeDrop = "drop"
	// FaultModePerturb adds Delta to the value of selected usage records on selected nodes.
	FaultModePerturb = "perturb"
)

// UsageFaultInjection drops and perturbs two independent deterministic
// samples of usage records so both billing reconcilers can measure their
// recall at the same time. Only records of allowlisted orgs (the canaries)
// are touched; the zero value is off.
type UsageFaultInjection struct {
	// Orgs are the org ids whose records may be faulted; required when any rate is set.
	Orgs []string
	// DropRate is the fraction of usage events dropped on every node, keyed on the event id.
	DropRate float64
	// PerturbRate is the fraction of usage events perturbed, keyed on the event id; drop wins when both select.
	PerturbRate float64
	// Delta is added to the value in a perturb; negative only, so a leak can never overbill.
	Delta int64
	// NodeRate is the fraction of nodes that apply a perturb; 1 (or 0) means every node.
	NodeRate float64
}

// Enabled reports whether any fault is configured.
func (f UsageFaultInjection) Enabled() bool { return f.DropRate > 0 || f.PerturbRate > 0 }

// Validate rejects unusable settings.
func (f UsageFaultInjection) Validate() error {
	var errs []error
	for name, v := range map[string]float64{"drop": f.DropRate, "perturb": f.PerturbRate, "node": f.NodeRate} {
		if math.IsNaN(v) || v < 0 || v > 1 { // NaN compares false everywhere, so test it explicitly
			errs = append(errs, fmt.Errorf("fault injection %s rate %v: want 0..1", name, v))
		}
	}
	if f.Enabled() && len(f.orgSet()) == 0 {
		errs = append(errs, errors.New("fault injection rate set without an org allowlist"))
	}
	if f.PerturbRate > 0 && f.Delta >= 0 {
		errs = append(errs, fmt.Errorf("fault injection delta %d: want negative", f.Delta))
	}
	return errors.Join(errs...)
}

func (f UsageFaultInjection) orgSet() map[string]struct{} {
	set := make(map[string]struct{}, len(f.Orgs))
	for _, o := range f.Orgs {
		if o = strings.TrimSpace(o); o != "" {
			set[o] = struct{}{}
		}
	}
	return set
}

// OrgAllowed reports whether records of orgID may be faulted. An empty org
// id (unresolved) is never allowed.
func (f UsageFaultInjection) OrgAllowed(orgID string) bool {
	if orgID == "" {
		return false
	}
	_, ok := f.orgSet()[strings.TrimSpace(orgID)]
	return ok
}

// FaultSelected reports whether key falls in the deterministic sample for
// rate. FNV-1a hashes the key; fmix64 spreads near-identical keys apart; the
// top 53 bits become a uniform fraction in [0, 1) compared against rate.
// Reconcilers recompute the sample from rate alone.
func FaultSelected(rate float64, key string) bool {
	if rate <= 0 {
		return false
	}
	if rate >= 1 {
		return true
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(key))
	return float64(fmix64(h.Sum64())>>11)/float64(uint64(1)<<53) < rate // 53 mantissa bits, no rounding bias
}

// fmix64 is the MurmurHash3 finalizer: xor-shift and multiply by its published
// constants so every input bit affects every output bit (FNV alone clusters short keys).
func fmix64(h uint64) uint64 {
	h ^= h >> 33
	h *= 0xff51afd7ed558ccd
	h ^= h >> 33
	h *= 0xc4ceb9fe1a85ec53
	h ^= h >> 33
	return h
}

// DropSelected reports whether the usage event is in the drop sample.
func (f UsageFaultInjection) DropSelected(eventID string) bool {
	return FaultSelected(f.DropRate, FaultModeDrop+"|"+eventID)
}

// PerturbSelected reports whether the usage event is in the perturb sample.
// The two samples are keyed apart; an event in both is dropped.
func (f UsageFaultInjection) PerturbSelected(eventID string) bool {
	return !f.DropSelected(eventID) && FaultSelected(f.PerturbRate, FaultModePerturb+"|"+eventID)
}

// NodeSelected reports whether this node applies a perturb for the event.
func (f UsageFaultInjection) NodeSelected(eventID, nodeID string) bool {
	if f.NodeRate <= 0 || f.NodeRate >= 1 {
		return true
	}
	return FaultSelected(f.NodeRate, nodeID+"|"+eventID)
}

// applyFault applies the configured fault to a usage utilization. It returns
// true when the record must not be emitted. Faults are logged as
// "Injected usage fault", a contract line consumed by the reconcilers.
func (rm *ResourceManager) applyFault(identity ResourceIdentity, eventID string, u *meteringpb.Utilization) (drop bool) {
	f := rm.faultInjection
	if u == nil || !f.OrgAllowed(u.GetOrgId()) {
		return false
	}
	mode := ""
	switch {
	case f.DropSelected(eventID):
		mode = FaultModeDrop
	case f.PerturbSelected(eventID) && f.NodeSelected(eventID, identity.NodeID()):
		mode = FaultModePerturb
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
	if mode == FaultModeDrop {
		rm.lggr.Warnw("Injected usage fault", append(fields, "value", u.GetValue())...)
		return true
	}
	original := u.GetValue()
	v, ok := new(big.Int).SetString(original, 10)
	if !ok {
		return false
	}
	v.Add(v, big.NewInt(f.Delta))
	if v.Sign() < 0 {
		v.SetInt64(0)
	}
	u.Value = v.String()
	rm.lggr.Warnw("Injected usage fault", append(fields, "value", u.Value, "originalValue", original)...)
	return false
}
