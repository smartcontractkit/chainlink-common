package aggregators

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
)

// TestMode_twoWayTieIsDeterministicAcrossRepeatedCalls targets a bimodal tie: two
// distinct values each appear the same number of times (here, 3× "tie-a" and 3× "tie-b").
// The winner must not depend on map iteration order, so repeated calls with the same
// multiset must return an observation equal under protobuf semantics every time.
func TestMode_twoWayTieIsDeterministicAcrossRepeatedCalls(t *testing.T) {
	t.Parallel()

	a, err := values.Wrap("tie-a")
	require.NoError(t, err)
	b, err := values.Wrap("tie-b")
	require.NoError(t, err)
	items := []values.Value{a, a, a, b, b, b}

	var baseline *pb.Value
	for i := range 400 {
		got, maxCount, err := mode(items)
		require.NoError(t, err)
		require.Equal(t, 3, maxCount)

		gotProto := values.Proto(got)
        // set baseline for comparison on first iteration, all subsequent must match
		if i == 0 {
			baseline = proto.Clone(gotProto).(*pb.Value)
			continue
		}
		require.True(t, proto.Equal(baseline, gotProto),
			"iteration %d: mode() must pick the same tied winner on every call", i)
	}
}
