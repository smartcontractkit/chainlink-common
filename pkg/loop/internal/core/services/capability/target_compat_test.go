package capability

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/remote/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
)

// The registry protocol is mid-migration from a numeric broker handle to a gRPC
// target string. Hosts and plugin binaries are built from different versions of
// chainlink-common and there is no handshake that gates them (no HandshakeConfig
// sets ProtocolVersion), so every writer must populate both fields and every
// reader must cope with a peer that only sets one. These tests pin that contract.

func TestBrokerConnID_prefersTargetFallsBackToHandle(t *testing.T) {
	for _, tt := range []struct {
		name     string
		target   string
		legacyID uint32
		want     uint32
	}{
		{name: "new peer: target wins", target: net.BrokerTarget(7), legacyID: 7, want: 7},
		{name: "old peer: no target, handle used", target: "", legacyID: 42, want: 42},
		{name: "old peer: zero handle still honoured", target: "", legacyID: 0, want: 0},
		{name: "target wins even if handle disagrees", target: net.BrokerTarget(9), legacyID: 3, want: 9},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := brokerConnID(tt.target, tt.legacyID)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBrokerConnID_rejectsMalformedTarget(t *testing.T) {
	// A malformed target must not silently fall back to the legacy handle: the peer
	// said which connection it meant, and we could not understand it.
	_, err := brokerConnID("dns:///elsewhere:1234", 7)
	assert.Error(t, err)
}

func TestListBrokerConnIDs(t *testing.T) {
	t.Run("new peer: targets replace handles", func(t *testing.T) {
		res := &pb.ListReply{
			CapabilityID: []uint32{1, 2, 3},
			Targets:      []string{net.BrokerTarget(1), net.BrokerTarget(2), net.BrokerTarget(3)},
		}
		got, err := listBrokerConnIDs(res)
		require.NoError(t, err)
		assert.Equal(t, []uint32{1, 2, 3}, got)
	})

	t.Run("old peer: handles only", func(t *testing.T) {
		got, err := listBrokerConnIDs(&pb.ListReply{CapabilityID: []uint32{4, 5}})
		require.NoError(t, err)
		assert.Equal(t, []uint32{4, 5}, got)
	})

	t.Run("empty reply", func(t *testing.T) {
		got, err := listBrokerConnIDs(&pb.ListReply{})
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("order is preserved", func(t *testing.T) {
		res := &pb.ListReply{Targets: []string{net.BrokerTarget(3), net.BrokerTarget(1), net.BrokerTarget(2)}}
		got, err := listBrokerConnIDs(res)
		require.NoError(t, err)
		assert.Equal(t, []uint32{3, 1, 2}, got)
	})

	t.Run("malformed target fails rather than falling back", func(t *testing.T) {
		res := &pb.ListReply{
			CapabilityID: []uint32{1, 2},
			Targets:      []string{net.BrokerTarget(1), "nonsense"},
		}
		_, err := listBrokerConnIDs(res)
		assert.Error(t, err)
	})
}

// TestDualWrite_oldReaderContract pins the guarantee a pre-target peer relies on:
// messages this version writes still carry a usable numeric handle, and it agrees
// with the target. Constructed the way the writers in this package construct them.
func TestDualWrite_oldReaderContract(t *testing.T) {
	const id uint32 = 11

	t.Run("AddRequest", func(t *testing.T) {
		req := &pb.AddRequest{CapabilityID: id, Target: net.BrokerTarget(id)}
		assert.Equal(t, id, req.GetCapabilityID(), "old reader reads the handle")
		fromTarget, err := net.ParseBrokerTarget(req.GetTarget())
		require.NoError(t, err)
		assert.Equal(t, req.GetCapabilityID(), fromTarget, "both fields must agree")
	})

	t.Run("GetReply", func(t *testing.T) {
		reply := &pb.GetReply{CapabilityID: id, Target: net.BrokerTarget(id)}
		assert.Equal(t, id, reply.GetCapabilityID())
		fromTarget, err := net.ParseBrokerTarget(reply.GetTarget())
		require.NoError(t, err)
		assert.Equal(t, reply.GetCapabilityID(), fromTarget)
	})

	t.Run("ListReply", func(t *testing.T) {
		reply := &pb.ListReply{}
		for _, i := range []uint32{1, 2} {
			reply.CapabilityID = append(reply.CapabilityID, i)
			reply.Targets = append(reply.Targets, net.BrokerTarget(i))
		}
		assert.Equal(t, []uint32{1, 2}, reply.GetCapabilityID(), "old reader reads the handles")
		require.Len(t, reply.GetTargets(), len(reply.GetCapabilityID()), "dual-write must stay parallel")
	})
}
