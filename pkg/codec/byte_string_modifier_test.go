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
	type concreteStruct struct {
		A string
		T [codec.Byte20Address]byte
	}

	type pointerStruct struct {
		A string
		T *[codec.Byte20Address]byte
	}

	type arrayStruct struct {
		A string
		T [2][codec.Byte20Address]byte
	}

	type sliceStruct struct {
		A string
		T [][codec.Byte20Address]byte
	}

	concretest := reflect.TypeOf(&concreteStruct{})
	pointertst := reflect.TypeOf(&pointerStruct{})
	arrayst := reflect.TypeOf(&arrayStruct{})
	slicest := reflect.TypeOf(&sliceStruct{})

	type otherIntegerType struct {
		A string
		T addressType
	}
	oit := reflect.TypeOf(&otherIntegerType{})

	testAddrBytes := [codec.Byte20Address]byte{}
	testAddrStr := "0x" + hex.EncodeToString(testAddrBytes[:])
	anyString := "test"

	t.Run("RetypeToOffChain converts fixed length bytes to string", func(t *testing.T) {
		for _, test := range []struct {
			name string
			tp   reflect.Type
		}{
			{"[20]byte", concretest},
			{"typed address", oit},
			{"[20]byte pointer", pointertst},
		} {
			t.Run(test.name, func(t *testing.T) {
				converter := codec.NewAddressBytesToStringModifier(codec.Byte20Address, codec.NoChecksum, []string{"T"})
				convertedType, err := converter.RetypeToOffChain(test.tp, "")

				require.NoError(t, err)
				assert.Equal(t, reflect.Pointer, convertedType.Kind())
				convertedType = convertedType.Elem()

				require.Equal(t, 2, convertedType.NumField())
				assert.Equal(t, test.tp.Elem().Field(0), convertedType.Field(0))
				assert.Equal(t, test.tp.Elem().Field(1).Name, convertedType.Field(1).Name)
				assert.Equal(t, reflect.TypeOf(""), convertedType.Field(1).Type)
			})
		}
	})

	t.Run("RetypeToOffChain converts converts arrays of fixed length bytes to array of string", func(t *testing.T) {
		converter := codec.NewAddressBytesToStringModifier(codec.Byte20Address, codec.NoChecksum, []string{"T"})

		convertedType, err := converter.RetypeToOffChain(arrayst, "")

		require.NoError(t, err)
		assert.Equal(t, reflect.Pointer, convertedType.Kind())
		convertedType = convertedType.Elem()

		require.Equal(t, 2, convertedType.NumField())
		assert.Equal(t, arrayst.Elem().Field(0), convertedType.Field(0))
		assert.Equal(t, reflect.TypeOf([2]string{}), convertedType.Field(1).Type)
	})

	t.Run("RetypeToOffChain converts slices of fixed length bytes to slices of string", func(t *testing.T) {
		converter := codec.NewAddressBytesToStringModifier(codec.Byte20Address, codec.NoChecksum, []string{"T"})

		convertedType, err := converter.RetypeToOffChain(slicest, "")

		require.NoError(t, err)
		assert.Equal(t, reflect.Pointer, convertedType.Kind())
		convertedType = convertedType.Elem()

		require.Equal(t, 2, convertedType.NumField())
		assert.Equal(t, slicest.Elem().Field(0), convertedType.Field(0))
		assert.Equal(t, reflect.TypeOf([]string{}), convertedType.Field(1).Type)
	})

	t.Run("TransformToOnChain converts string to bytes", func(t *testing.T) {
		for _, test := range []struct {
			name     string
			t        reflect.Type
			expected any
		}{
			{"[20]byte", concretest, &concreteStruct{A: anyString, T: [codec.Byte20Address]byte{}}},
			{"typed address", oit, &otherIntegerType{A: anyString, T: addressType{}}},
			{"[20]byte pointer", pointertst, &pointerStruct{A: anyString, T: &[codec.Byte20Address]byte{}}},
		} {
			t.Run(test.name, func(t *testing.T) {
				converter := codec.NewAddressBytesToStringModifier(codec.Byte20Address, codec.NoChecksum, []string{"T"})
				convertedType, err := converter.RetypeToOffChain(test.t, "")
				require.NoError(t, err)

				rOffchain := reflect.New(convertedType.Elem())
				iOffChain := reflect.Indirect(rOffchain)
				iOffChain.FieldByName("A").SetString(anyString)
				iOffChain.FieldByName("T").Set(reflect.ValueOf(testAddrStr))

				actual, err := converter.TransformToOnChain(rOffchain.Interface(), "")
				require.NoError(t, err)

				assert.Equal(t, test.expected, actual)
			})
		}
	})

	t.Run("TransformToOnChain converts string array to array of fixed length bytes", func(t *testing.T) {
		converter := codec.NewAddressBytesToStringModifier(codec.Byte20Address, codec.NoChecksum, []string{"T"})

		convertedType, err := converter.RetypeToOffChain(arrayst, "")
		require.NoError(t, err)

		rOffchain := reflect.New(convertedType.Elem())
		iOffChain := reflect.Indirect(rOffchain)

		arrayValue := [2]string{testAddrStr, testAddrStr}

		iOffChain.FieldByName("T").Set(reflect.ValueOf(arrayValue))

		actual, err := converter.TransformToOnChain(rOffchain.Interface(), "")
		require.NoError(t, err)

		expected := &arrayStruct{A: "", T: [2][codec.Byte20Address]byte{}}
		assert.Equal(t, expected, actual)
	})

	t.Run("TransformToOnChain converts string slice to slice of [20]byte", func(t *testing.T) {
		converter := codec.NewAddressBytesToStringModifier(codec.Byte20Address, codec.NoChecksum, []string{"T"})

		convertedType, err := converter.RetypeToOffChain(slicest, "")
		require.NoError(t, err)

		rOffchain := reflect.New(convertedType.Elem())
		iOffChain := reflect.Indirect(rOffchain)

		iOffChain.FieldByName("T").Set(reflect.ValueOf([]string{testAddrStr, testAddrStr}))

		actual, err := converter.TransformToOnChain(rOffchain.Interface(), "")
		require.NoError(t, err)

		expected := &sliceStruct{
			A: "",
			T: [][codec.Byte20Address]byte{
				testAddrBytes,
				testAddrBytes,
			},
		}

		assert.Equal(t, expected, actual)
	})

	t.Run("TransformToOffChain converts bytes to string", func(t *testing.T) {
		for _, test := range []struct {
			name     string
			t        reflect.Type
			offChain any
		}{
			{"[20]byte", concretest, &concreteStruct{A: anyString, T: [codec.Byte20Address]byte{}}},
			{"typed address", oit, &otherIntegerType{A: anyString, T: addressType{}}},
			{"[20]byte pointer", pointertst, &pointerStruct{A: anyString, T: &[codec.Byte20Address]byte{}}},
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
				iOffChain.FieldByName("T").Set(reflect.ValueOf(testAddrStr))
				assert.Equal(t, expected.Interface(), actual)
			})
		}
	})

	t.Run("TransformToOffChain converts array of bytes to string array", func(t *testing.T) {
		converter := codec.NewAddressBytesToStringModifier(codec.Byte20Address, codec.NoChecksum, []string{"T"})

		convertedType, err := converter.RetypeToOffChain(arrayst, "")
		require.NoError(t, err)

		rOffchain := reflect.New(convertedType.Elem())
		iOffChain := reflect.Indirect(rOffchain)
		expectedAddrs := [2]string{testAddrStr, testAddrStr}
		iOffChain.FieldByName("T").Set(reflect.ValueOf(expectedAddrs))

		actual, err := converter.TransformToOffChain(&arrayStruct{A: anyString, T: [2][codec.Byte20Address]byte{}}, "")
		require.NoError(t, err)

		expected := reflect.New(convertedType.Elem())
		iExpected := reflect.Indirect(expected)
		iExpected.FieldByName("A").SetString(anyString)
		iExpected.FieldByName("T").Set(reflect.ValueOf(expectedAddrs))
		assert.Equal(t, expected.Interface(), actual)
	})

	t.Run("TransformToOffChain converts slice bytes to string slice", func(t *testing.T) {
		converter := codec.NewAddressBytesToStringModifier(codec.Byte20Address, codec.NoChecksum, []string{"T"})

		convertedType, err := converter.RetypeToOffChain(slicest, "")
		require.NoError(t, err)

		rOffchain := reflect.New(convertedType.Elem())
		iOffChain := reflect.Indirect(rOffchain)
		expectedAddrs := []string{testAddrStr, testAddrStr}
		iOffChain.FieldByName("T").Set(reflect.ValueOf(expectedAddrs))

		actual, err := converter.TransformToOffChain(&sliceStruct{
			A: anyString,
			T: [][codec.Byte20Address]byte{testAddrBytes, testAddrBytes},
		}, "")
		require.NoError(t, err)

		expected := reflect.New(convertedType.Elem())
		iExpected := reflect.Indirect(expected)
		iExpected.FieldByName("A").SetString(anyString)
		iExpected.FieldByName("T").Set(reflect.ValueOf(expectedAddrs))
		assert.Equal(t, expected.Interface(), actual)
	})
}
