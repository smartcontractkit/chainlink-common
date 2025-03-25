package aggregation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ByzantineQuorum(t *testing.T) {
	assert.Equal(t, 1, ByzantineQuorum(1, 1))

	// N = 3F + 1
	assert.Equal(t, 3, ByzantineQuorum(1, 4))
	assert.Equal(t, 5, ByzantineQuorum(2, 7))

	// N > 3F + 1
	assert.Equal(t, 4, ByzantineQuorum(1, 5))
	assert.Equal(t, 4, ByzantineQuorum(1, 6))
	assert.Equal(t, 6, ByzantineQuorum(2, 9))
}
