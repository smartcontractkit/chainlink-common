package aggregators_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/aggregators"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

// =============================================================================
// Bug: identicalAggregator.collectHighestCounts() map iteration tie-breaking
//      in identical.go:73-103
//
// The collectHighestCounts function iterates a map[sha256]*counter with a
// strict ">" comparison (line 80):
//
//   for _, counter := range shaToCounter {
//       if counter.count > highestCount {   // <-- strict greater-than
//           highestCount = counter.count
//           highestObservation = counter.fullObservation
//       }
//   }
//
// When two distinct observation values have the SAME count, the first entry
// encountered in map iteration order wins (since "count > highestCount" is
// false for equal counts). Go randomizes map iteration order, so different
// processes (nodes) may pick different winners.
//
// This is exploitable when n >= 2*(2f+1), allowing two groups to each reach
// the 2f+1 quorum threshold. Example: f=2, n=10, 2f+1=5 -- two groups of 5.
//
// In production, nodes disagree on the outcome, causing:
//   "PrepareSignature failed to verify. This is commonly caused by
//    non-determinism in the ReportingPlugin"
//
// Run: go test -v -run TestIdentical -count=1
// =============================================================================

// TestIdenticalTiedCounts creates two observation groups of equal size, both
// meeting the 2f+1 threshold, and verifies whether the aggregator produces
// consistent results.
func TestIdenticalTiedCounts(t *testing.T) {
	config := newIdenticalTestConfig(t, nil)
	agg, err := aggregators.NewIdenticalAggregator(*config)
	require.NoError(t, err)

	// f=2, 2f+1=5. With 10 nodes: 5 report "alpha", 5 report "beta".
	// Both groups meet the quorum threshold. The > comparison means
	// whichever map entry is iterated first sets highestCount=5, and the
	// other entry (also count=5) fails the > check. The winner depends
	// on map iteration order.
	observations := map[commontypes.OracleID][]values.Value{
		0: {values.NewString("alpha")},
		1: {values.NewString("alpha")},
		2: {values.NewString("alpha")},
		3: {values.NewString("alpha")},
		4: {values.NewString("alpha")},
		5: {values.NewString("beta")},
		6: {values.NewString("beta")},
		7: {values.NewString("beta")},
		8: {values.NewString("beta")},
		9: {values.NewString("beta")},
	}

	const iterations = 200
	seen := make(map[string]int)
	for i := 0; i < iterations; i++ {
		outcome, err := agg.Aggregate(logger.Nop(), nil, observations, 2) // f=2
		require.NoError(t, err)
		require.NotNil(t, outcome)

		m, err := values.FromMapValueProto(outcome.EncodableOutcome)
		require.NoError(t, err)

		val := m.Underlying["0"]
		require.NotNil(t, val)

		b, err := proto.MarshalOptions{Deterministic: true}.Marshal(values.Proto(val))
		require.NoError(t, err)
		seen[string(b)]++
	}

	if len(seen) > 1 {
		t.Errorf("CONFIRMED: identicalAggregator non-determinism -- produced %d distinct outcomes "+
			"over %d iterations (identical.go:79 map iteration tie-breaking with equal counts)",
			len(seen), iterations)
	}
}

// TestIdenticalCrossNodeSimulation simulates multiple nodes running the
// identical aggregator on the same observations. All nodes should agree
// on the same outcome.
func TestIdenticalCrossNodeSimulation(t *testing.T) {
	config := newIdenticalTestConfig(t, nil)

	// f=2, 2f+1=5. Two groups of 5 -- both meet quorum.
	observations := map[commontypes.OracleID][]values.Value{
		0: {values.NewString("alpha")},
		1: {values.NewString("alpha")},
		2: {values.NewString("alpha")},
		3: {values.NewString("alpha")},
		4: {values.NewString("alpha")},
		5: {values.NewString("beta")},
		6: {values.NewString("beta")},
		7: {values.NewString("beta")},
		8: {values.NewString("beta")},
		9: {values.NewString("beta")},
	}

	const numNodes = 10
	outcomeBytes := make([][]byte, numNodes)
	for i := 0; i < numNodes; i++ {
		agg, err := aggregators.NewIdenticalAggregator(*config)
		require.NoError(t, err)

		outcome, err := agg.Aggregate(logger.Nop(), nil, observations, 2) // f=2
		require.NoError(t, err)
		require.NotNil(t, outcome)

		b, err := proto.MarshalOptions{Deterministic: true}.Marshal(outcome.EncodableOutcome)
		require.NoError(t, err)
		outcomeBytes[i] = b
	}

	allMatch := true
	for i := 1; i < numNodes; i++ {
		if string(outcomeBytes[i]) != string(outcomeBytes[0]) {
			allMatch = false
			break
		}
	}

	if !allMatch {
		t.Errorf("CONFIRMED: Cross-node consensus failure -- %d simulated nodes produced "+
			"different outcome bytes. Root cause: identicalAggregator map iteration "+
			"tie-breaking in identical.go:79", numNodes)
	}
}

// TestIdenticalQuorumEnforcement is a sanity check that verifies the aggregator
// correctly rejects observations when neither group reaches the 2f+1 threshold.
func TestIdenticalQuorumEnforcement(t *testing.T) {
	config := newIdenticalTestConfig(t, nil)
	agg, err := aggregators.NewIdenticalAggregator(*config)
	require.NoError(t, err)

	// f=1, 2f+1=3. With 4 nodes: 2 report "A", 2 report "B".
	// Neither group meets the quorum of 3.
	observations := map[commontypes.OracleID][]values.Value{
		0: {values.NewString("A")},
		1: {values.NewString("A")},
		2: {values.NewString("B")},
		3: {values.NewString("B")},
	}

	outcome, err := agg.Aggregate(logger.Nop(), nil, observations, 1) // f=1
	require.Error(t, err)
	require.Nil(t, outcome)
	require.Contains(t, err.Error(), "can't reach consensus")
}

func newIdenticalTestConfig(t *testing.T, overrideKeys []string) *values.Map {
	t.Helper()
	unwrappedConfig := map[string]any{
		"expectedObservationsLen": len(overrideKeys),
		"keyOverrides":            overrideKeys,
	}
	config, err := values.NewMap(unwrappedConfig)
	require.NoError(t, err)
	return config
}
