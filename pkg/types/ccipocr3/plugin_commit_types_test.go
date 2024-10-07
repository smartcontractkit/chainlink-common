package ccipocr3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommitPluginReport(t *testing.T) {
	t.Run("is empty", func(t *testing.T) {
		r := CommitPluginReport{}
		assert.True(t, r.IsEmpty())
	})

	t.Run("is not empty", func(t *testing.T) {
		r := CommitPluginReport{
			MerkleRoots: make([]MerkleRootChain, 1),
		}
		assert.False(t, r.IsEmpty())

		r = CommitPluginReport{
			PriceUpdates: PriceUpdates{
				TokenPriceUpdates: make([]TokenPrice, 1),
				GasPriceUpdates:   make([]GasPriceChain, 1),
			},
		}
		assert.False(t, r.IsEmpty())

		r = CommitPluginReport{
			RMNSignatures: make([]RMNECDSASignature, 1),
		}
		assert.False(t, r.IsEmpty())

		r = CommitPluginReport{
			MerkleRoots: make([]MerkleRootChain, 1),
			PriceUpdates: PriceUpdates{
				TokenPriceUpdates: make([]TokenPrice, 1),
				GasPriceUpdates:   make([]GasPriceChain, 1),
			},
			RMNSignatures: make([]RMNECDSASignature, 1),
		}
		assert.False(t, r.IsEmpty())
	})
}
