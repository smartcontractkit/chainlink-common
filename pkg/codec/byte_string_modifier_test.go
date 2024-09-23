package codec_test

import (
	"encoding/hex"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/codec"
)

type addressType [codec.Byte20Address]byte

func TestAddressBytesToString(t *testing.T) {
	type testStruct struct {
		A string
		T [codec.Byte20Address]byte
	}
	tst := reflect.TypeOf(&testStruct{})

	type otherIntegerType struct {
		A string
		T addressType
	}
	oit := reflect.TypeOf(&otherIntegerType{})

	empty := [codec.Byte20Address]byte{}
	testAddr := "0x" + hex.EncodeToString(empty[:])

	t.Run("RetypeToOffChain converts fix length bytes types", func(t *testing.T) {
		for _, test := range []struct {
			name string
			t    reflect.Type
		}{
			{"[20]byte", tst},
			{"typed address", oit},
		} {
			t.Run(test.name, func(t *testing.T) {
				converter := codec.NewAddressBytesToStringModifier(codec.Byte20Address, codec.NoChecksum, []string{"T"})
				convertedType, err := converter.RetypeToOffChain(test.t, "")

				require.NoError(t, err)
				assert.Equal(t, reflect.Pointer, convertedType.Kind())
				convertedType = convertedType.Elem()

				require.Equal(t, 2, convertedType.NumField())
				assert.Equal(t, tst.Elem().Field(0), convertedType.Field(0))
				assert.Equal(t, tst.Elem().Field(1).Name, convertedType.Field(1).Name)
				assert.Equal(t, reflect.TypeOf(""), convertedType.Field(1).Type)
			})
		}
	})

	t.Run("TransformToOnChain converts time to integer types", func(t *testing.T) {
		anyString := "test"
		for _, test := range []struct {
			name     string
			t        reflect.Type
			expected any
		}{
			{"[20]byte", tst, &testStruct{A: anyString, T: [codec.Byte20Address]byte{}}},
			{"typed address", oit, &otherIntegerType{A: anyString, T: addressType{}}},
		} {
			t.Run(test.name, func(t *testing.T) {
				converter := codec.NewAddressBytesToStringModifier(codec.Byte20Address, codec.NoChecksum, []string{"T"})
				convertedType, err := converter.RetypeToOffChain(test.t, "")
				require.NoError(t, err)

				rOffchain := reflect.New(convertedType.Elem())
				iOffChain := reflect.Indirect(rOffchain)
				iOffChain.FieldByName("A").SetString(anyString)
				iOffChain.FieldByName("T").Set(reflect.ValueOf(testAddr))

				actual, err := converter.TransformToOnChain(rOffchain.Interface(), "")
				require.NoError(t, err)

				assert.Equal(t, test.expected, actual)
			})
		}
	})

	t.Run("TransformToOffChain converts bytes to string", func(t *testing.T) {
		anyString := "test"
		for _, test := range []struct {
			name     string
			t        reflect.Type
			offChain any
		}{
			{"[20]byte", tst, &testStruct{A: anyString, T: [codec.Byte20Address]byte{}}},
			{"typed address", oit, &otherIntegerType{A: anyString, T: addressType{}}},
		} {
			t.Run(test.name, func(t *testing.T) {
				converter := codec.NewAddressBytesToStringModifier(codec.Byte20Address, codec.NoChecksum, []string{"T"})
				convertedType, err := converter.RetypeToOffChain(test.t, "")
				require.NoError(t, err)

				actual, err := converter.TransformToOffChain(test.offChain, "")
				require.NoError(t, err)

				expected := reflect.New(convertedType.Elem())
				iOffChain := reflect.Indirect(expected)
				iOffChain.FieldByName("A").SetString(anyString)
				iOffChain.FieldByName("T").Set(reflect.ValueOf(testAddr))
				assert.Equal(t, expected.Interface(), actual)
			})
		}
	})
}
