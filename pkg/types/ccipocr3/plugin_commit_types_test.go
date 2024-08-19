package ccipocr3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommitPluginReport(t *testing.T) {
	t.Run("is empty", func(t *testing.T) {
		r := NewCommitPluginReport(nil, nil, nil)
		assert.True(t, r.IsEmpty())
	})

	t.Run("is not empty", func(t *testing.T) {
		r := NewCommitPluginReport(make([]MerkleRootChain, 1), nil, nil)
		assert.False(t, r.IsEmpty())

		r = NewCommitPluginReport(nil, make([]TokenPrice, 1), make([]GasPriceChain, 1))
		assert.False(t, r.IsEmpty())

		r = NewCommitPluginReport(make([]MerkleRootChain, 1), make([]TokenPrice, 1), make([]GasPriceChain, 1))
		assert.False(t, r.IsEmpty())
	})
}
