package codec_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/codec"
)

func TestMultiModifier(t *testing.T) {
	testType := reflect.TypeOf(chainModifierTestStruct{})
	mod1 := codec.NewRenamer(map[string]string{"A": "B"})
	mod2 := codec.NewRenamer(map[string]string{"B": "C"})
	chainMod := codec.MultiModifier{mod1, mod2}
	t.Run("RetypeForOffChain chains modifiers", func(t *testing.T) {
		offChain, err := chainMod.RetypeForOffChain(testType, "")
		require.NoError(t, err)
		m1, err := mod1.RetypeForOffChain(testType, "")
		require.NoError(t, err)
		expected, err := mod2.RetypeForOffChain(m1, "")
		require.NoError(t, err)
		assert.Equal(t, expected, offChain)
	})

	t.Run("TransformForOffChain chains modifiers", func(t *testing.T) {
		_, err := chainMod.RetypeForOffChain(testType, "")
		require.NoError(t, err)

		input := chainModifierTestStruct{A: 100}
		actual, err := chainMod.TransformForOffChain(input, "")
		require.NoError(t, err)

		m1, err := mod1.TransformForOffChain(input, "")
		require.NoError(t, err)
		expected, err := mod2.TransformForOffChain(m1, "")
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("TransformForOnChain chains modifiers", func(t *testing.T) {
		offChainType, err := chainMod.RetypeForOffChain(testType, "")
		require.NoError(t, err)

		input := reflect.New(offChainType).Elem()
		input.FieldByName("C").SetInt(100)
		actual, err := chainMod.TransformForOnChain(input.Interface(), "")
		require.NoError(t, err)

		expected := chainModifierTestStruct{A: 100}
		assert.Equal(t, expected, actual)
	})
}

type chainModifierTestStruct struct {
	A int
}
