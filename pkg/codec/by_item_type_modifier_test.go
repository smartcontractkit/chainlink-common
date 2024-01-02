package codec_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/codec"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

func TestByTypeModifier(t *testing.T) {
	anyitemType := "anything"
	modifierCodecChainMod := codec.NewRenamer(map[string]string{"A": "B"})
	mod, err := codec.NewByItemTypeModifier(map[string]codec.Modifier{anyitemType: modifierCodecChainMod})
	require.NoError(t, err)
	t.Run("Uses modifier for the type", func(t *testing.T) {
		offChainType, err := mod.RetypeForOffChain(reflect.TypeOf(&modifierCodecChainType{}), anyitemType)
		require.NoError(t, err)

		expectedType, err := modifierCodecChainMod.RetypeForOffChain(reflect.TypeOf(&modifierCodecChainType{}), anyitemType)
		require.NoError(t, err)
		assert.Equal(t, expectedType, offChainType)

		item := &modifierCodecChainType{A: 100}
		offChain, err := mod.TransformForOffChain(item, anyitemType)
		require.NoError(t, err)
		actualOffChain, err := modifierCodecChainMod.TransformForOffChain(item, anyitemType)
		require.NoError(t, err)
		assert.Equal(t, actualOffChain, offChain)

		onChain, err := mod.TransformForOnChain(offChain, anyitemType)
		require.NoError(t, err)
		assert.Equal(t, item, onChain)
	})

	t.Run("Returns error if modifier isn't found", func(t *testing.T) {
		_, err := mod.RetypeForOffChain(reflect.TypeOf(&modifierCodecOffChainType{}), "different")
		assert.True(t, errors.Is(err, types.ErrInvalidType))

		_, err = mod.TransformForOnChain(&modifierCodecChainType{}, "different")
		assert.True(t, errors.Is(err, types.ErrInvalidType))

		_, err = mod.TransformForOffChain(reflect.TypeOf(&modifierCodecOffChainType{}), "different")
		assert.True(t, errors.Is(err, types.ErrInvalidType))
	})
}
