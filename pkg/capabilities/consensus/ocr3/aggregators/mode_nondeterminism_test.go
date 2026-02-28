package aggregators_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/libocr/commontypes"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3/aggregators"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

// =============================================================================
// Bug: mode() tie-breaking via map iteration in reduce_aggregator.go:464-480
//
// The mode() function builds a map[sha256]*counter to count observation
// frequencies, then iterates the map twice:
//   1. Find the maximum count (lines 464-469)
//   2. Collect all values with that count into a slice (lines 471-476)
//
// It returns modes[0] — the first element of the collected slice. Because Go
// map iteration order is randomized, the slice ordering varies between calls,
// making the result non-deterministic when multiple values tie for highest
// frequency.
//
// In production (OCR3 consensus), every node independently runs Outcome() on
// the same observations. If mode() returns different values on different nodes,
// their outcome bytes diverge and PrepareSignature verification fails with:
//   "PrepareSignature failed to verify. This is commonly caused by
//    non-determinism in the ReportingPlugin"
//
// Run: go test -v -run TestMode -count=1
// =============================================================================

// TestModeTiedFrequencies creates a 2-way tie (equal frequency for two values)
// and verifies that repeated aggregation produces consistent results.
func TestModeTiedFrequencies(t *testing.T) {
	fields := []aggregators.AggregationField{
		{
			InputKey:   "data",
			OutputKey:  "data",
			Method:     "mode",
			ModeQuorum: "any", // don't enforce f+1 quorum so tie results are returned
		},
	}

	config := newModeTestConfig(t, fields)
	agg, err := aggregators.NewReduceAggregator(*config)
	require.NoError(t, err)

	// 3 nodes report "value_A", 3 nodes report "value_B" -- a perfect tie
	mkObs := func(val string) values.Value {
		m, err := values.WrapMap(map[string]any{"data": val})
		require.NoError(t, err)
		return m
	}

	observations := map[commontypes.OracleID][]values.Value{
		0: {mkObs("value_A")},
		1: {mkObs("value_A")},
		2: {mkObs("value_A")},
		3: {mkObs("value_B")},
		4: {mkObs("value_B")},
		5: {mkObs("value_B")},
	}

	// Run once to get a reference result
	firstOutcome, err := agg.Aggregate(logger.Nop(), nil, observations, 1)
	require.NoError(t, err)
	require.NotNil(t, firstOutcome)

	// Run 100 more times -- if mode() is deterministic, all results match.
	// If non-deterministic, at least one will differ.
	const iterations = 100
	mismatchCount := 0
	for i := 0; i < iterations; i++ {
		outcome, err := agg.Aggregate(logger.Nop(), nil, observations, 1)
		require.NoError(t, err)
		require.NotNil(t, outcome)

		if !proto.Equal(firstOutcome.EncodableOutcome, outcome.EncodableOutcome) {
			mismatchCount++
		}
	}

	assert.Zerof(t, mismatchCount,
		"mode() produced different outcomes in %d/%d iterations with tied frequencies -- "+
			"this confirms non-deterministic tie-breaking via map iteration in reduce_aggregator.go:464-480",
		mismatchCount, iterations)
}

// TestModeThreeWayTie uses a 3-way tie to further increase the probability
// of observing different map iteration orderings.
func TestModeThreeWayTie(t *testing.T) {
	fields := []aggregators.AggregationField{
		{
			InputKey:   "data",
			OutputKey:  "data",
			Method:     "mode",
			ModeQuorum: "any",
		},
	}

	config := newModeTestConfig(t, fields)
	agg, err := aggregators.NewReduceAggregator(*config)
	require.NoError(t, err)

	mkObs := func(val string) values.Value {
		m, err := values.WrapMap(map[string]any{"data": val})
		require.NoError(t, err)
		return m
	}

	// 3-way tie: 2 nodes each report "X", "Y", "Z"
	observations := map[commontypes.OracleID][]values.Value{
		0: {mkObs("X")},
		1: {mkObs("X")},
		2: {mkObs("Y")},
		3: {mkObs("Y")},
		4: {mkObs("Z")},
		5: {mkObs("Z")},
	}

	const iterations = 200
	seen := make(map[string]int)
	for i := 0; i < iterations; i++ {
		outcome, err := agg.Aggregate(logger.Nop(), nil, observations, 1)
		require.NoError(t, err)
		require.NotNil(t, outcome)

		m, err := values.FromMapValueProto(outcome.EncodableOutcome)
		require.NoError(t, err)

		reports := m.Underlying["Reports"]
		require.NotNil(t, reports)

		b, err := proto.MarshalOptions{Deterministic: true}.Marshal(values.Proto(reports))
		require.NoError(t, err)
		seen[string(b)]++
	}

	if len(seen) > 1 {
		t.Errorf("CONFIRMED: mode() non-determinism -- produced %d distinct outcomes over %d iterations "+
			"(reduce_aggregator.go:464-480 map iteration tie-breaking)",
			len(seen), iterations)
	}
}

// TestModeCrossNodeConsensusSimulation simulates an actual OCR3 round: multiple
// "nodes" independently run the same aggregation on the same observations.
// If the aggregation is deterministic all nodes produce identical outcome bytes.
// If non-deterministic, nodes disagree and PrepareSignature fails.
func TestModeCrossNodeConsensusSimulation(t *testing.T) {
	fields := []aggregators.AggregationField{
		{
			InputKey:   "price_source",
			OutputKey:  "price_source",
			Method:     "mode",
			ModeQuorum: "any",
		},
	}

	config := newModeTestConfig(t, fields)

	mkObs := func(val string) values.Value {
		m, err := values.WrapMap(map[string]any{"price_source": val})
		require.NoError(t, err)
		return m
	}

	// Volatile data source: 3 nodes saw "coinbase", 3 saw "binance" -- a tie
	observations := map[commontypes.OracleID][]values.Value{
		0: {mkObs("coinbase")},
		1: {mkObs("coinbase")},
		2: {mkObs("coinbase")},
		3: {mkObs("binance")},
		4: {mkObs("binance")},
		5: {mkObs("binance")},
	}

	// Each simulated node creates its own aggregator and runs independently
	const numNodes = 10
	outcomeBytes := make([][]byte, numNodes)
	for i := 0; i < numNodes; i++ {
		agg, err := aggregators.NewReduceAggregator(*config)
		require.NoError(t, err)

		outcome, err := agg.Aggregate(logger.Nop(), nil, observations, 1)
		require.NoError(t, err)
		require.NotNil(t, outcome)

		b, err := proto.MarshalOptions{Deterministic: true}.Marshal(outcome.EncodableOutcome)
		require.NoError(t, err)
		outcomeBytes[i] = b
	}

	// In a real OCR3 round, all nodes must produce identical outcome bytes
	allMatch := true
	for i := 1; i < numNodes; i++ {
		if string(outcomeBytes[i]) != string(outcomeBytes[0]) {
			allMatch = false
			break
		}
	}

	if !allMatch {
		t.Errorf("CONFIRMED: Cross-node consensus failure -- %d simulated nodes produced "+
			"different outcome bytes for the same observations. In production this causes "+
			"\"PrepareSignature failed to verify\" errors. Root cause: mode() tie-breaking "+
			"via map iteration in reduce_aggregator.go:464-480", numNodes)
	}
}

func newModeTestConfig(t *testing.T, fields []aggregators.AggregationField) *values.Map {
	t.Helper()
	unwrappedConfig := map[string]any{
		"fields":          fields,
		"outputFieldName": "Reports",
		"reportFormat":    "array",
	}
	config, err := values.NewMap(unwrappedConfig)
	require.NoError(t, err)
	return config
}
