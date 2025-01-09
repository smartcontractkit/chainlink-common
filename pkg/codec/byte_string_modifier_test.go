package codec_test

import (
	"encoding/hex"
	"errors"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/codec"
)

// MockAddressModifier is a mock implementation of the AddressModifier interface.
type MockAddressModifier struct {
	length int
}

func (m MockAddressModifier) EncodeAddress(bytes []byte) (string, error) {
	return "0x" + hex.EncodeToString(bytes), nil
}

func (m MockAddressModifier) DecodeAddress(str string) ([]byte, error) {
	if len(str) == 0 {
		return nil, errors.New("empty address")
	}
	return hex.DecodeString(str[2:]) // Skip the "0x" prefix for hex encoding
}

func (m MockAddressModifier) Length() int {
	return m.length
}

func TestAddressBytesToString(t *testing.T) {
	// Mocking AddressModifier for 20-byte addresses
	mockModifier := MockAddressModifier{length: 20}

	type concreteStruct struct {
		A string
		T [20]byte
	}

	type concreteStructWithLargeAddress struct {
		A string
		T [20]byte
	}

	type pointerStruct struct {
		A string
		T *[20]byte
	}

	type arrayStruct struct {
		A string
		T [2][20]byte
	}

	type sliceStruct struct {
		A string
		T [][20]byte
	}

	concretest := reflect.TypeOf(&concreteStruct{})
	concreteLargest := reflect.TypeOf(&concreteStructWithLargeAddress{})
	pointertst := reflect.TypeOf(&pointerStruct{})
	arrayst := reflect.TypeOf(&arrayStruct{})
	slicest := reflect.TypeOf(&sliceStruct{})

	type Bytes20AddressType [20]byte

	type otherIntegerType struct {
		A string
		T Bytes20AddressType
	}

	type pointerOtherIntegerType struct {
		A string
		T *Bytes20AddressType
	}
	oit := reflect.TypeOf(&otherIntegerType{})
	oitpt := reflect.TypeOf(&pointerOtherIntegerType{})

	testAddrBytes := [20]byte{}
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
			{"*typed address", oitpt},
		} {
			t.Run(test.name, func(t *testing.T) {
				converter := codec.NewAddressBytesToStringModifier([]string{"T"}, mockModifier)
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

	t.Run("RetypeToOffChain converts arrays of fixed length bytes to array of string", func(t *testing.T) {
		converter := codec.NewAddressBytesToStringModifier([]string{"T"}, mockModifier)

		convertedType, err := converter.RetypeToOffChain(arrayst, "")
		require.NoError(t, err)
		assert.Equal(t, reflect.Pointer, convertedType.Kind())
		convertedType = convertedType.Elem()

		require.Equal(t, 2, convertedType.NumField())
		assert.Equal(t, arrayst.Elem().Field(0), convertedType.Field(0))
		assert.Equal(t, reflect.TypeOf([2]string{}), convertedType.Field(1).Type)
	})

	t.Run("RetypeToOffChain converts slices of fixed length bytes to slices of string", func(t *testing.T) {
		converter := codec.NewAddressBytesToStringModifier([]string{"T"}, mockModifier)

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
			{"[20]byte", concretest, &concreteStruct{A: anyString, T: [20]byte{}}},
			{"*[20]byte", pointertst, &pointerStruct{A: anyString, T: &[20]byte{}}},
			{"typed address", oit, &otherIntegerType{A: anyString, T: Bytes20AddressType{}}},
			{"*typed address", oitpt, &pointerOtherIntegerType{A: anyString, T: &Bytes20AddressType{}}},
		} {
			t.Run(test.name, func(t *testing.T) {
				converter := codec.NewAddressBytesToStringModifier([]string{"T"}, mockModifier)
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
		converter := codec.NewAddressBytesToStringModifier([]string{"T"}, mockModifier)

		convertedType, err := converter.RetypeToOffChain(arrayst, "")
		require.NoError(t, err)

		rOffchain := reflect.New(convertedType.Elem())
		iOffChain := reflect.Indirect(rOffchain)

		arrayValue := [2]string{testAddrStr, testAddrStr}

		iOffChain.FieldByName("T").Set(reflect.ValueOf(arrayValue))

		actual, err := converter.TransformToOnChain(rOffchain.Interface(), "")
		require.NoError(t, err)

		expected := &arrayStruct{A: "", T: [2][20]byte{}}
		assert.Equal(t, expected, actual)
	})

	t.Run("TransformToOnChain converts string slice to slice of [length]byte", func(t *testing.T) {
		converter := codec.NewAddressBytesToStringModifier([]string{"T"}, mockModifier)

		convertedType, err := converter.RetypeToOffChain(slicest, "")
		require.NoError(t, err)

		rOffchain := reflect.New(convertedType.Elem())
		iOffChain := reflect.Indirect(rOffchain)

		iOffChain.FieldByName("T").Set(reflect.ValueOf([]string{testAddrStr, testAddrStr}))

		actual, err := converter.TransformToOnChain(rOffchain.Interface(), "")
		require.NoError(t, err)

		expected := &sliceStruct{
			A: "",
			T: [][20]byte{
				testAddrBytes,
				testAddrBytes,
			},
		}

		assert.Equal(t, expected, actual)
	})

	t.Run("TransformToOnChain returns error on invalid inputs", func(t *testing.T) {
		converter := codec.NewAddressBytesToStringModifier([]string{"T"}, mockModifier)

		tests := []struct {
			name       string
			addrStr    string
			structType reflect.Type
		}{
			{
				name:       "Invalid length input",
				addrStr:    "0x" + hex.EncodeToString([]byte{1, 2, 3}),
				structType: concretest,
			},
			{
				name:       "Larger than expected input",
				addrStr:    "0x" + hex.EncodeToString(make([]byte, 40)),
				structType: concreteLargest,
			},
			{
				name:       "Empty string input",
				addrStr:    "",
				structType: concretest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				convertedType, err := converter.RetypeToOffChain(tt.structType, "")
				require.NoError(t, err)

				rOffchain := reflect.New(convertedType.Elem())
				iOffChain := reflect.Indirect(rOffchain)
				iOffChain.FieldByName("A").SetString(anyString)
				iOffChain.FieldByName("T").Set(reflect.ValueOf(tt.addrStr))

				_, err = converter.TransformToOnChain(rOffchain.Interface(), "")
				require.Error(t, err)
			})
		}
	})

	t.Run("TransformToOffChain converts bytes to string", func(t *testing.T) {
		for _, test := range []struct {
			name     string
			t        reflect.Type
			offChain any
		}{
			{"[20]byte", concretest, &concreteStruct{A: anyString, T: [20]byte{}}},
			{"*[20]byte", pointertst, &pointerStruct{A: anyString, T: &[20]byte{}}},
			{"typed address", oit, &otherIntegerType{A: anyString, T: Bytes20AddressType{}}},
			{"*typed address", oitpt, &pointerOtherIntegerType{A: anyString, T: &Bytes20AddressType{}}},
		} {
			t.Run(test.name, func(t *testing.T) {
				converter := codec.NewAddressBytesToStringModifier([]string{"T"}, mockModifier)
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
		converter := codec.NewAddressBytesToStringModifier([]string{"T"}, mockModifier)

		convertedType, err := converter.RetypeToOffChain(arrayst, "")
		require.NoError(t, err)

		rOffchain := reflect.New(convertedType.Elem())
		iOffChain := reflect.Indirect(rOffchain)
		expectedAddrs := [2]string{testAddrStr, testAddrStr}
		iOffChain.FieldByName("T").Set(reflect.ValueOf(expectedAddrs))

		actual, err := converter.TransformToOffChain(&arrayStruct{A: anyString, T: [2][20]byte{}}, "")
		require.NoError(t, err)

		expected := reflect.New(convertedType.Elem())
		iExpected := reflect.Indirect(expected)
		iExpected.FieldByName("A").SetString(anyString)
		iExpected.FieldByName("T").Set(reflect.ValueOf(expectedAddrs))
		assert.Equal(t, expected.Interface(), actual)
	})

	t.Run("TransformToOffChain converts slice bytes to string slice", func(t *testing.T) {
		converter := codec.NewAddressBytesToStringModifier([]string{"T"}, mockModifier)

		convertedType, err := converter.RetypeToOffChain(slicest, "")
		require.NoError(t, err)

		rOffchain := reflect.New(convertedType.Elem())
		iOffChain := reflect.Indirect(rOffchain)
		expectedAddrs := []string{testAddrStr, testAddrStr}
		iOffChain.FieldByName("T").Set(reflect.ValueOf(expectedAddrs))

		actual, err := converter.TransformToOffChain(&sliceStruct{
			A: anyString,
			T: [][20]byte{testAddrBytes, testAddrBytes},
		}, "")
		require.NoError(t, err)

		expected := reflect.New(convertedType.Elem())
		iExpected := reflect.Indirect(expected)
		iExpected.FieldByName("A").SetString(anyString)
		iExpected.FieldByName("T").Set(reflect.ValueOf(expectedAddrs))
		assert.Equal(t, expected.Interface(), actual)
	})

	t.Run("Unsupported field type returns error", func(t *testing.T) {
		converter := codec.NewAddressBytesToStringModifier([]string{"T"}, mockModifier)

		unsupportedStruct := struct {
			A string
			T int // Unsupported type
		}{}

		// We expect RetypeToOffChain to return an error because 'T' is not a supported type.
		_, err := converter.RetypeToOffChain(reflect.TypeOf(&unsupportedStruct), "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot convert bytes for field T")
	})
}
