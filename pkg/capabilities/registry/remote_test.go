// External test package: keeps this file matching the integration suite that exercises the remote
// registry against a real registry server (in the capabilities repo's crecore/registry package, the
// sole consumer of that production code).
package registry_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry"
)

// --- conversions ---

func TestCapabilityTypeConvertersRoundTrip(t *testing.T) {
	// Round-tripping must be lossless in both directions; the wire type is the
	// shared capabilities.pb enum, so there is no collapsing step to undo.
	for _, in := range []capabilities.CapabilityType{
		capabilities.CapabilityTypeTrigger,
		capabilities.CapabilityTypeAction,
		capabilities.CapabilityTypeConsensus,
		capabilities.CapabilityTypeTarget,
		capabilities.CapabilityTypeCombined,
		capabilities.CapabilityTypeUnknown,
	} {
		pbType, err := capabilitiespb.CapabilityTypeToProto(in)
		require.NoError(t, err)

		back, err := capabilitiespb.CapabilityTypeFromProto(pbType)
		require.NoError(t, err)
		assert.Equal(t, in, back)
	}
}

func TestDONFromProto_Nil(t *testing.T) {
	assert.Equal(t, capabilities.DON{}, registry.DONFromProto(nil))
}
