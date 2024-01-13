package interfacetests

import (
	"math/big"
	"testing"

	"github.com/smartcontractkit/libocr/commontypes"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

func RunCodecInterfaceFuzzTests(f *testing.F, tester CodecInterfaceTester) {
	// A flattening of the TestStruct, replacing the Account/Accounts with just an int that can be used to make an account
	f.Add(int32(0), "foo", uint8(0), []byte{} /*[]int8 is not supported, byte is the same size*/, 0, 1, int64(1), []byte{} /*fixed size isn't allowed*/, 1, "foo")
	f.Fuzz(func(t *testing.T, field int32, differentField string, oracleId uint8, oracleIds []byte, accountSeed, accountsSeed int, bigField int64, fixedBytes []byte, i int, s string) {
		t.Run("Encode and decode produce the same result", func(t *testing.T) {
			tester.Setup(t)
			oids := [32]commontypes.OracleID{}
			for index, id := range oracleIds {
				if index == len(oids) {
					break
				}
				oids[index] = commontypes.OracleID(id)
			}

			fb := [2]byte{}
			for index := 0; index < 2 && index < len(fixedBytes); index++ {
				fb[index] = fixedBytes[index]
			}

			testStruct := &TestStruct{
				Field:          field,
				DifferentField: differentField,
				OracleID:       commontypes.OracleID(oracleId),
				OracleIDs:      oids,
				Account:        tester.GetAccountBytes(accountSeed),
				Accounts:       [][]byte{tester.GetAccountBytes(accountsSeed + 1), tester.GetAccountBytes(accountsSeed + 2)},
				BigField:       big.NewInt(bigField),
				NestedStruct: MidLevelTestStruct{
					FixedBytes: fb,
					Inner: InnerTestStruct{
						I: i,
						S: s,
					},
				},
			}
			codec := tester.GetCodec(t)
			ctx := tests.Context(t)

			encoded, err := codec.Encode(ctx, testStruct, TestItemType)
			require.NoError(t, err)
			decoded := &TestStruct{}
			err = codec.Decode(ctx, encoded, decoded, TestItemType)
			require.NoError(t, err)
			// big.Int can represent zero in many ways, make it the same
			if testStruct.BigField.Cmp(big.NewInt(0)) == 0 {
				require.Equal(t, 0, decoded.BigField.Cmp(big.NewInt(0)))
				decoded.BigField = testStruct.BigField
			}
			require.Equal(t, testStruct, decoded)
		})
	})
}
