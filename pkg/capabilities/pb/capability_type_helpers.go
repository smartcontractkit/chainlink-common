package pb

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

// CapabilityTypeToProto converts a capability type to its wire form.
//
// Unknown is a legal value on both sides: a capability may genuinely not declare
// a type, and only callers that need one (e.g. registering a capability, which
// must know which services to serve) should reject it.
func CapabilityTypeToProto(t capabilities.CapabilityType) (CapabilityType, error) {
	switch t {
	case capabilities.CapabilityTypeTrigger:
		return CapabilityTypeTrigger, nil
	case capabilities.CapabilityTypeAction:
		return CapabilityTypeAction, nil
	case capabilities.CapabilityTypeConsensus:
		return CapabilityTypeConsensus, nil
	case capabilities.CapabilityTypeTarget:
		return CapabilityTypeTarget, nil
	case capabilities.CapabilityTypeCombined:
		return CapabilityTypeCombined, nil
	case capabilities.CapabilityTypeUnknown:
		return CapabilityTypeUnknown, nil
	default:
		return CapabilityTypeUnknown, fmt.Errorf("unknown capability type %q", t)
	}
}

// CapabilityTypeFromProto converts a wire capability type to the Go type.
func CapabilityTypeFromProto(t CapabilityType) (capabilities.CapabilityType, error) {
	switch t {
	case CapabilityTypeTrigger:
		return capabilities.CapabilityTypeTrigger, nil
	case CapabilityTypeAction:
		return capabilities.CapabilityTypeAction, nil
	case CapabilityTypeConsensus:
		return capabilities.CapabilityTypeConsensus, nil
	case CapabilityTypeTarget:
		return capabilities.CapabilityTypeTarget, nil
	case CapabilityTypeCombined:
		return capabilities.CapabilityTypeCombined, nil
	case CapabilityTypeUnknown:
		return capabilities.CapabilityTypeUnknown, nil
	default:
		return capabilities.CapabilityTypeUnknown, fmt.Errorf("unknown capability type %s", t)
	}
}
