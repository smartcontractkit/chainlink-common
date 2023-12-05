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
	modifierCodecChainMod := codec.NewRenamer(map[string]string{"A": "B"})
	modByType, err := codec.NewModifierByType(map[reflect.Type]codec.Modifier{
		reflect.TypeOf(&modifierCodecChainType{}): modifierCodecChainMod,
	})
	require.NoError(t, err)
	t.Run("Uses modifier for the type", func(t *testing.T) {
		offChainType, err := modByType.RetypeForOffChain(reflect.TypeOf(&modifierCodecChainType{}))
		require.NoError(t, err)

		expectedType, err := modifierCodecChainMod.RetypeForOffChain(reflect.TypeOf(&modifierCodecChainType{}))
		require.NoError(t, err)
		assert.Equal(t, expectedType, offChainType)

		item := &modifierCodecChainType{A: 100}
		offChain, err := modByType.TransformForOffChain(item)
		require.NoError(t, err)
		actualOffChain, err := modifierCodecChainMod.TransformForOffChain(item)
		require.NoError(t, err)
		assert.Equal(t, actualOffChain, offChain)

		onChain, err := modByType.TransformForOnChain(offChain)
		require.NoError(t, err)
		assert.Equal(t, item, onChain)
	})

	t.Run("Returns error if modifier isn't found", func(t *testing.T) {
		_, err := modByType.RetypeForOffChain(reflect.TypeOf(&modifierCodecOffChainType{}))
		assert.Equal(t, types.ErrInvalidType, err)

		_, err = modByType.TransformForOnChain(&modifierCodecChainType{})
		assert.Equal(t, types.ErrInvalidType, err)

		_, err = modByType.TransformForOffChain(reflect.TypeOf(&modifierCodecOffChainType{}))
		assert.Equal(t, types.ErrInvalidType, err)
	})

	t.Run("New returns errors if type cannot be retyped", func(t *testing.T) {
		_, err := codec.NewModifierByType(map[reflect.Type]codec.Modifier{
			reflect.TypeOf(&modifierCodecChainType{}): codec.NewRenamer(map[string]string{"Azzz": "B"}),
		})
		require.True(t, errors.Is(err, types.ErrInvalidConfig))
	})
}
