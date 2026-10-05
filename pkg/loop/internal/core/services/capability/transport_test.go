package capability

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/remote"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
)

// The registry protocol is mid-migration from a numeric broker handle to a gRPC
// target string. Hosts and plugin binaries are built from different versions of
// chainlink-common and no handshake gates them (no HandshakeConfig sets
// ProtocolVersion), so this transport must cope with a peer that names a connection
// either way. These tests pin that contract.

func TestBrokerConnID_prefersTargetFallsBackToHandle(t *testing.T) {
	for _, tt := range []struct {
		name string
		loc  remote.Locator
		want uint32
	}{
		{
			name: "new peer: target wins",
			loc:  remote.Locator{Target: net.BrokerTarget(7), LegacyHandle: 7},
			want: 7,
		},
		{
			name: "old peer: no target, handle used",
			loc:  remote.Locator{LegacyHandle: 42},
			want: 42,
		},
		{
			name: "old peer: zero handle still honoured",
			loc:  remote.Locator{},
			want: 0,
		},
		{
			name: "target wins even if the handle disagrees",
			loc:  remote.Locator{Target: net.BrokerTarget(9), LegacyHandle: 3},
			want: 9,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := brokerConnID(tt.loc)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBrokerConnID_rejectsMalformedTarget(t *testing.T) {
	// A target we cannot parse must not silently fall back to the handle: the peer
	// said which connection it meant and we could not understand it, so dialling the
	// handle could reach something else entirely.
	_, err := brokerConnID(remote.Locator{Target: "dns:///elsewhere:1234", LegacyHandle: 7})
	assert.Error(t, err)
}
