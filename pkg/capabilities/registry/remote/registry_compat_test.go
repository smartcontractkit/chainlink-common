package remote

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	p2ptypes "github.com/smartcontractkit/libocr/ragep2p/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"

	registrypb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/remote/pb"
)

// List carries locators as two parallel repeated fields, one of which predates the
// other. These tests pin how the pair is written and read, which is what lets a peer
// built before targets existed keep working.

func TestLocatorsFromListReply(t *testing.T) {
	t.Run("new peer: targets are authoritative", func(t *testing.T) {
		got := locatorsFromListReply(&registrypb.ListReply{
			CapabilityID: []uint32{1, 2, 3},
			Targets:      []string{"a", "b", "c"},
		})
		require.Len(t, got, 3)
		assert.Equal(t, []Locator{
			{Target: "a", LegacyHandle: 1},
			{Target: "b", LegacyHandle: 2},
			{Target: "c", LegacyHandle: 3},
		}, got)
	})

	t.Run("old peer: handles only", func(t *testing.T) {
		got := locatorsFromListReply(&registrypb.ListReply{CapabilityID: []uint32{4, 5}})
		assert.Equal(t, []Locator{{LegacyHandle: 4}, {LegacyHandle: 5}}, got)
	})

	t.Run("empty reply", func(t *testing.T) {
		assert.Empty(t, locatorsFromListReply(&registrypb.ListReply{}))
	})

	t.Run("order is preserved", func(t *testing.T) {
		got := locatorsFromListReply(&registrypb.ListReply{Targets: []string{"c", "a", "b"}})
		assert.Equal(t, []string{"c", "a", "b"}, []string{got[0].Target, got[1].Target, got[2].Target})
	})

	t.Run("targets shorter than handles does not panic", func(t *testing.T) {
		got := locatorsFromListReply(&registrypb.ListReply{
			CapabilityID: []uint32{1},
			Targets:      []string{"a", "b"},
		})
		require.Len(t, got, 2)
		assert.Equal(t, Locator{Target: "a", LegacyHandle: 1}, got[0])
		assert.Equal(t, Locator{Target: "b"}, got[1], "no handle to pair with")
	})
}

func TestListReplyFor(t *testing.T) {
	t.Run("names every capability both ways", func(t *testing.T) {
		reply := listReplyFor([]Locator{
			{Target: "a", LegacyHandle: 1},
			{Target: "b", LegacyHandle: 2},
		})
		assert.Equal(t, []uint32{1, 2}, reply.GetCapabilityID(), "an old peer reads the handles")
		assert.Equal(t, []string{"a", "b"}, reply.GetTargets())
	})

	t.Run("omits targets entirely when any is missing", func(t *testing.T) {
		// A reader treats targets as replacing handles wholesale, so a partial list
		// would be read as the complete set and the unnamed capability would vanish.
		reply := listReplyFor([]Locator{
			{Target: "a", LegacyHandle: 1},
			{LegacyHandle: 2},
		})
		assert.Equal(t, []uint32{1, 2}, reply.GetCapabilityID())
		assert.Empty(t, reply.GetTargets())
	})

	t.Run("empty", func(t *testing.T) {
		reply := listReplyFor(nil)
		assert.Empty(t, reply.GetCapabilityID())
		assert.Empty(t, reply.GetTargets())
	})

	t.Run("round trips through the reader", func(t *testing.T) {
		locs := []Locator{{Target: "a", LegacyHandle: 1}, {Target: "b", LegacyHandle: 2}}
		assert.Equal(t, locs, locatorsFromListReply(listReplyFor(locs)))
	})
}

func TestToDON(t *testing.T) {
	don := &registrypb.DON{
		Id:   0,
		Name: "test-don",
		Members: [][]byte{
			{0: 4, 31: 0},
			{0: 5, 31: 0},
		},
		F:             2,
		ConfigVersion: 1,
		Families:      []string{"a"},
		Config:        []byte("test-config"),
	}

	expected := capabilities.DON{
		ID:   0,
		Name: "test-don",
		Members: []p2ptypes.PeerID{
			[32]byte{0: 4},
			[32]byte{0: 5},
		},
		F:             2,
		ConfigVersion: 1,
		Families:      []string{"a"},
		Config:        []byte("test-config"),
	}

	actual := toDON(don)

	require.Equal(t, expected, actual)
}
