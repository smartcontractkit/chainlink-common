package ccipocr3

import (
	"testing"

	chainsel "github.com/smartcontractkit/chain-selectors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtraDataCodec_DecodeExtraArgs(t *testing.T) {
	extraDataCodec := ExtraDataCodec{
		chainsel.FamilySolana: nil,
	}

	t.Run("empty extraArgs returns nil", func(t *testing.T) {
		result, err := extraDataCodec.DecodeExtraArgs(Bytes{}, ChainSelector(1))
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("invalid chain selector", func(t *testing.T) {
		extraArgs := Bytes{1, 2, 3}

		result, err := extraDataCodec.DecodeExtraArgs(extraArgs, ChainSelector(999999))

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get chain family")
	})

	t.Run("unsupported family", func(t *testing.T) {
		extraArgs := Bytes{4, 5, 6}

		result, err := extraDataCodec.DecodeExtraArgs(extraArgs, ChainSelector(222782988166878823))

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unsupported family for extra args type")
	})
}

func TestExtraDataCodec_DecodeTokenAmountDestExecData(t *testing.T) {
	extraDataCodec := ExtraDataCodec{
		chainsel.FamilySolana: nil,
	}

	t.Run("empty destExecData returns nil", func(t *testing.T) {
		result, err := extraDataCodec.DecodeTokenAmountDestExecData(Bytes{}, ChainSelector(1))
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("invalid chain selector", func(t *testing.T) {
		destExecData := Bytes{4, 5, 6}

		result, err := extraDataCodec.DecodeTokenAmountDestExecData(destExecData, ChainSelector(999999))

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get chain family")
	})

	t.Run("unsupported family", func(t *testing.T) {
		destExecData := Bytes{7, 8, 9}

		result, err := extraDataCodec.DecodeTokenAmountDestExecData(destExecData, ChainSelector(222782988166878823))

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "unsupported family for dest exec data type")
	})
}
