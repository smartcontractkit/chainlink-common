package aggregation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ByzantineQuorum(t *testing.T) {
	assert.Equal(t, 1, ByzantineQuorum(1, 1))

	// N = 3F + 1
	assert.Equal(t, 3, ByzantineQuorum(4, 1))
	assert.Equal(t, 5, ByzantineQuorum(7, 2))

	// N > 3F + 1
	assert.Equal(t, 4, ByzantineQuorum(5, 1))
	assert.Equal(t, 4, ByzantineQuorum(6, 1))
	assert.Equal(t, 6, ByzantineQuorum(9, 2))

}
