package ocr3

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
)

// TestPickEncoderDeterministic_BugFix pins the fix for the OCR3 map-iteration
// tiebreak bug at reporting_plugin.go:396-407. Under OCR3 this only caused
// intermittent outcome divergence that the outer consensus masked; under
// OCR3_1 each node writes the result to its local KV, so divergence never
// self-heals. This test runs the tiebreak many times with distinct map
// insertion orders to ensure the result is byte-identical.
func TestPickEncoderDeterministic_BugFix(t *testing.T) {
	cfgA := encoderConfig{name: "enc-a", config: &pb.Map{}}
	cfgB := encoderConfig{name: "enc-b", config: &pb.Map{}}
	cfgC := encoderConfig{name: "enc-c", config: &pb.Map{}}

	shaToEncoder := map[string]encoderConfig{
		"sha-a": cfgA,
		"sha-b": cfgB,
		"sha-c": cfgC,
	}

	t.Run("picks winner by count desc", func(t *testing.T) {
		counts := map[string]int{
			"sha-a": 4,
			"sha-b": 7, // winner
			"sha-c": 5,
		}
		got := pickEncoderDeterministic(counts, shaToEncoder, 5)
		require.NotNil(t, got)
		assert.Equal(t, "enc-b", got.name)
	})

	t.Run("ties broken by sha asc (deterministic across runs)", func(t *testing.T) {
		counts := map[string]int{
			"sha-a": 7,
			"sha-b": 7,
			"sha-c": 7,
		}
		// Run many times; Go map iteration order is randomized, but our
		// sort must make the outcome identical every time.
		for i := 0; i < 50; i++ {
			got := pickEncoderDeterministic(counts, shaToEncoder, 5)
			require.NotNil(t, got)
			assert.Equal(t, "enc-a", got.name, "iteration %d", i)
		}
	})

	t.Run("nil when nothing reaches threshold", func(t *testing.T) {
		counts := map[string]int{
			"sha-a": 3,
			"sha-b": 4,
			"sha-c": 2,
		}
		got := pickEncoderDeterministic(counts, shaToEncoder, 5)
		assert.Nil(t, got)
	})

	t.Run("skips sha with no matching encoder", func(t *testing.T) {
		counts := map[string]int{
			"sha-ghost": 10, // top count but no entry in shaToEncoder
			"sha-a":     5,
		}
		got := pickEncoderDeterministic(counts, shaToEncoder, 5)
		require.NotNil(t, got)
		assert.Equal(t, "enc-a", got.name)
	})

	t.Run("empty input returns nil", func(t *testing.T) {
		assert.Nil(t, pickEncoderDeterministic(nil, shaToEncoder, 5))
		assert.Nil(t, pickEncoderDeterministic(map[string]int{}, shaToEncoder, 5))
	})
}

// TestContainsString is trivial but pinned because containsString is on the
// hot prune path — a silent rewrite of it breaks KV pruning correctness.
func TestContainsString(t *testing.T) {
	assert.True(t, containsString([]string{"a", "b", "c"}, "b"))
	assert.False(t, containsString([]string{"a", "b", "c"}, "d"))
	assert.False(t, containsString(nil, "a"))
	assert.False(t, containsString([]string{}, "a"))
}
