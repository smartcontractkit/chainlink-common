package net

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrokerTarget_roundTrip(t *testing.T) {
	for _, id := range []uint32{0, 1, 7, 65535, 4294967295} {
		target := BrokerTarget(id)
		got, err := ParseBrokerTarget(target)
		require.NoError(t, err, "target %q", target)
		assert.Equal(t, id, got)
	}
}

func TestBrokerTarget_format(t *testing.T) {
	assert.Equal(t, "broker://7", BrokerTarget(7))
}

func TestParseBrokerTarget_rejects(t *testing.T) {
	for _, target := range []string{
		"",                    // no target at all
		"7",                   // bare id, no scheme
		"dns:///host:1234",    // a different scheme
		"broker://",           // scheme but no id
		"broker://abc",        // non-numeric id
		"broker://-1",         // negative
		"broker://4294967296", // overflows uint32
	} {
		_, err := ParseBrokerTarget(target)
		assert.Error(t, err, "expected %q to be rejected", target)
	}
}
