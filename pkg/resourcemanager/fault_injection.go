package resourcemanager

import (
	"errors"
	"fmt"
	"hash/fnv"
	"math/big"
	"strings"

	meteringpb "github.com/smartcontractkit/chainlink-protos/metering/go"
)

// Usage fault injection modes.
const (
	// FaultModeDrop skips emission of selected usage records on every node.
	FaultModeDrop = "drop"
	// FaultModePerturb adds Delta to the value of selected usage records on selected nodes.
	FaultModePerturb = "perturb"
)

// UsageFaultInjection drops or perturbs a deterministic sample of usage
// records so billing reconcilers can measure their own recall. Test
// environments only; the zero value is off.
type UsageFaultInjection struct {
	// Mode is "" (off), "drop" or "perturb".
	Mode string
	// Rate is the fraction of usage events selected, keyed on the event id.
	Rate float64
	// Delta is added to the value in perturb mode (may be negative).
	Delta int64
	// NodeRate is the fraction of nodes that apply a perturb; 1 (or 0) means every node.
	NodeRate float64
}

// Enabled reports whether any fault is configured.
func (f UsageFaultInjection) Enabled() bool { return f.Mode != "" && f.Rate > 0 }

// Validate rejects unusable settings.
func (f UsageFaultInjection) Validate() error {
	var errs []error
	switch f.Mode {
	case "", FaultModeDrop, FaultModePerturb:
	default:
		errs = append(errs, fmt.Errorf("fault injection mode %q: want drop or perturb", f.Mode))
	}
	if f.Rate < 0 || f.Rate > 1 {
		errs = append(errs, fmt.Errorf("fault injection rate %v: want 0..1", f.Rate))
	}
	if f.NodeRate < 0 || f.NodeRate > 1 {
		errs = append(errs, fmt.Errorf("fault injection node rate %v: want 0..1", f.NodeRate))
	}
	if f.Rate > 0 && f.Mode == "" {
		errs = append(errs, errors.New("fault injection rate set without a mode"))
	}
	return errors.Join(errs...)
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

// EventSelected reports whether the usage event is in the sample.
func (f UsageFaultInjection) EventSelected(eventID string) bool {
	return f.Enabled() && FaultSelected(f.Rate, eventID)
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
	if u == nil || !f.EventSelected(eventID) {
		return false
	}
	workflowID, executionID, _ := strings.Cut(u.GetResourceId(), ":")
	fields := []any{
		"mode", f.Mode,
		"resourceType", u.GetResourceType(),
		"eventID", eventID,
		"resourceId", u.GetResourceId(),
		"workflowID", workflowID,
		"executionID", executionID,
		"orgID", u.GetOrgId(),
		"nodeID", identity.NodeID(),
	}
	switch f.Mode {
	case FaultModeDrop:
		rm.lggr.Warnw("Injected usage fault", append(fields, "value", u.GetValue())...)
		return true
	case FaultModePerturb:
		if !f.NodeSelected(eventID, identity.NodeID()) {
			return false
		}
		original := u.GetValue()
		v, ok := new(big.Int).SetString(original, 10)
		if !ok {
			return false
		}
		u.Value = v.Add(v, big.NewInt(f.Delta)).String()
		rm.lggr.Warnw("Injected usage fault", append(fields, "value", u.Value, "originalValue", original)...)
	}
	return false
}
