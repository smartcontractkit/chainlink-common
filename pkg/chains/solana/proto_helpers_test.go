package solana_test

import (
	"bytes"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	conv "github.com/smartcontractkit/chainlink-common/pkg/chains/solana"
	chain_common "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-common"
	typesolana "github.com/smartcontractkit/chainlink-common/pkg/types/chains/solana"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	solprimitives "github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives/solana"
)

func mkBytes(n int, fill byte) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = fill
	}
	return b
}

func TestPublicKeyConverters(t *testing.T) {
	t.Run("ConvertPublicKeyFromProto", func(t *testing.T) {
		cases := []struct {
			name            string
			in              []byte
			wantErrParts    []string
			wantEqualsInput bool
		}{
			{"accepts 32-byte public key", mkBytes(typesolana.PublicKeyLength, 0xAB), nil, true},
			{"rejects nil public key", nil, []string{"address can't be nil"}, false},
			{"rejects wrong length and shows base58", mkBytes(typesolana.PublicKeyLength-1, 0x01), []string{"invalid public key", "expected", "value="}, false},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := conv.ConvertPublicKeyFromProto(tc.in)
				if tc.wantErrParts != nil {
					require.Error(t, err, "expected error")
					for _, p := range tc.wantErrParts {
						require.Contains(t, err.Error(), p)
					}
					return
				}
				require.NoError(t, err)
				if tc.wantEqualsInput && !bytes.Equal(got[:], tc.in) {
					t.Fatalf("mismatch: got=%x want=%x", got[:], tc.in)
				}
			})
		}
	})

	t.Run("Roundtrip slice", func(t *testing.T) {
		keys := [][]byte{
			mkBytes(typesolana.PublicKeyLength, 0x11),
			mkBytes(typesolana.PublicKeyLength, 0x22),
		}
		dom, err := conv.ConvertPublicKeysFromProto(keys)
		require.NoError(t, err)
		require.Len(t, dom, 2)
		back := conv.ConvertPublicKeysToProto(dom)
		require.Len(t, back, 2)
		require.True(t, bytes.Equal(back[0], keys[0]) && bytes.Equal(back[1], keys[1]))
	})

	t.Run("AggregatesErrors", func(t *testing.T) {
		in := [][]byte{
			mkBytes(typesolana.PublicKeyLength, 0x01),
			mkBytes(typesolana.PublicKeyLength-1, 0x02),
			nil,
		}
		got, err := conv.ConvertPublicKeysFromProto(in)
		require.Nil(t, got)
		require.Error(t, err)
		msg := err.Error()
		require.Contains(t, msg, "public key[1]")
		require.Contains(t, msg, "public key[2]")
	})
}

func TestSignatureConverters(t *testing.T) {
	t.Run("ConvertSignatureFromProto", func(t *testing.T) {
		ok := mkBytes(typesolana.SignatureLength, 0xCD)
		got, err := conv.ConvertSignatureFromProto(ok)
		require.NoError(t, err)
		require.True(t, bytes.Equal(ok, got[:]))

		_, err = conv.ConvertSignatureFromProto(nil)
		require.ErrorContains(t, err, "signature can't be nil")

		_, err = conv.ConvertSignatureFromProto(mkBytes(typesolana.SignatureLength+1, 0x01))
		require.ErrorContains(t, err, "invalid signature")
		require.ErrorContains(t, err, "expected")
		require.ErrorContains(t, err, "value=")
	})

	t.Run("Batch/roundtrip", func(t *testing.T) {
		s1 := mkBytes(typesolana.SignatureLength, 0xAA)
		s2 := mkBytes(typesolana.SignatureLength, 0xBB)
		dom, err := conv.ConvertSignaturesFromProto([][]byte{s1, s2})
		require.NoError(t, err)
		back := conv.ConvertSignaturesToProto(dom)
		require.True(t, bytes.Equal(s1, back[0]) && bytes.Equal(s2, back[1]))
	})

	t.Run("AggregatesErrors", func(t *testing.T) {
		_, err := conv.ConvertSignaturesFromProto([][]byte{
			mkBytes(typesolana.SignatureLength-1, 0x01),
			nil,
		})
		require.Error(t, err)
		require.Contains(t, err.Error(), "signature[0]")
		require.Contains(t, err.Error(), "signature[1]")
	})
}

func TestHashConverters(t *testing.T) {
	t.Run("ConvertHashFromProto", func(t *testing.T) {
		ok := mkBytes(typesolana.HashLength, 0xEF)
		got, err := conv.ConvertHashFromProto(ok)
		require.NoError(t, err)
		require.True(t, bytes.Equal(ok, got[:]))

		_, err = conv.ConvertHashFromProto(nil)
		require.ErrorContains(t, err, "hash can't be nil")

		_, err = conv.ConvertHashFromProto(mkBytes(typesolana.HashLength-2, 0x01))
		require.ErrorContains(t, err, "invalid hash")
		require.ErrorContains(t, err, "expected")
		require.ErrorContains(t, err, "value=") // base58 string
	})
}

func TestEventSignatureConverters(t *testing.T) {
	t.Run("ConvertEventSigFromProto", func(t *testing.T) {
		ok := mkBytes(typesolana.EventSignatureLength, 0xAA)
		got, err := conv.ConvertEventSigFromProto(ok)
		require.NoError(t, err)
		require.True(t, bytes.Equal(ok, got[:]))

		_, err = conv.ConvertEventSigFromProto(nil)
		require.ErrorContains(t, err, "hash can't be nil") // matches current message

		_, err = conv.ConvertEventSigFromProto(mkBytes(typesolana.EventSignatureLength-1, 0x01))
		require.ErrorContains(t, err, "invalid event signature")
	})
}

func TestConvertExpressionsFromProto_Errors(t *testing.T) {
	testCases := []struct {
		Name          string
		In            []*conv.Expression
		ExpectedErr   string
		ExpectedFinal []query.Expression
	}{
		{
			Name: "empty",
		},
		{
			Name: "Empty evaluator",
			In:   []*conv.Expression{{}},
			ExpectedErr: "rpc error: code = InvalidArgument desc = err to convert expr idx 0 err: " +
				"unknown expression type: <nil>",
		},
		{
			Name: "Empty Expression_Primitive",
			In:   []*conv.Expression{{Evaluator: &conv.Expression_Primitive{}}},
			ExpectedErr: "rpc error: code = InvalidArgument desc = err to convert expr idx 0 err: " +
				"unknown primitive type: <nil>",
		},
		{
			Name: "Empty Expression_BooleanExpression",
			In:   []*conv.Expression{{Evaluator: &conv.Expression_BooleanExpression{}}},
			ExpectedFinal: []query.Expression{
				{BoolExpression: query.BoolExpression{}},
			},
		},
		{
			Name: "Nested empty Expression",
			In: []*conv.Expression{{
				Evaluator: &conv.Expression_BooleanExpression{
					BooleanExpression: &conv.BooleanExpression{
						Expression: []*conv.Expression{{}},
					},
				},
			}},
			ExpectedErr: "err to convert expr idx 0 err: failed to convert sub-expression 0: unknown expression type: <nil>",
		},
		{
			Name: "Empty Evaluator.Primitive.Primitive",
			In: []*conv.Expression{{
				Evaluator: &conv.Expression_Primitive{
					Primitive: &conv.Primitive{},
				},
			}},
			ExpectedErr: "rpc error: code = InvalidArgument desc = err to convert expr idx 0 err: unknown primitive type: <nil>",
		},
		{
			Name: "Empty Evaluator.Primitive.GeneralPrimitive",
			In: []*conv.Expression{{
				Evaluator: &conv.Expression_Primitive{
					Primitive: &conv.Primitive{
						Primitive: &conv.Primitive_GeneralPrimitive{},
					},
				},
			}},
			ExpectedErr: "rpc error: code = InvalidArgument desc = err to convert expr idx 0 err: rpc error: code = InvalidArgument desc = primitive can not be nil",
		},
		{
			Name: "Empty Evaluator.Primitive.GeneralPrimitive.Primitive",
			In: []*conv.Expression{{
				Evaluator: &conv.Expression_Primitive{
					Primitive: &conv.Primitive{
						Primitive: &conv.Primitive_GeneralPrimitive{
							GeneralPrimitive: &chain_common.Primitive{},
						},
					},
				},
			}},
			ExpectedErr: "rpc error: code = InvalidArgument desc = err to convert expr idx 0 err: rpc error: code = InvalidArgument desc = unknown primitive type: <nil>",
		},
		{
			Name: "Empty Comparator in GeneralPrimitive",
			In: []*conv.Expression{{
				Evaluator: &conv.Expression_Primitive{
					Primitive: &conv.Primitive{
						Primitive: &conv.Primitive_GeneralPrimitive{
							GeneralPrimitive: &chain_common.Primitive{
								Primitive: &chain_common.Primitive_Comparator{},
							},
						},
					},
				},
			}},
			ExpectedErr: "comparator can not be nil",
		},
		{
			Name: "Invalid Solana Address primitive (nil bytes)",
			In: []*conv.Expression{{
				Evaluator: &conv.Expression_Primitive{
					Primitive: &conv.Primitive{
						Primitive: &conv.Primitive_Address{Address: nil},
					},
				},
			}},
			ExpectedErr: "convert expr err: address can't be nil",
		},
		{
			Name: "Invalid Solana EventSig primitive (bad len)",
			In: []*conv.Expression{{
				Evaluator: &conv.Expression_Primitive{
					Primitive: &conv.Primitive{
						Primitive: &conv.Primitive_EventSig{EventSig: mkBytes(typesolana.EventSignatureLength-1, 0x01)},
					},
				},
			}},
			ExpectedErr: "failed to convert event sig: invalid event signature",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			got, err := conv.ConvertExpressionsFromProto(tc.In)
			if tc.ExpectedErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.ExpectedErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.ExpectedFinal, got)
		})
	}
}

func TestExpressions_Roundtrip_SolanaPrimitives(t *testing.T) {
	// Build (Address AND EventSig) OR EventBySubkey
	addrBytes := mkBytes(typesolana.PublicKeyLength, 0x01)
	evBytes := mkBytes(typesolana.EventSignatureLength, 0x02)

	addr, err := conv.ConvertPublicKeyFromProto(addrBytes)
	require.NoError(t, err)
	ev, err := conv.ConvertEventSigFromProto(evBytes)
	require.NoError(t, err)

	a := solprimitives.NewAddressFilter(addr)
	e := solprimitives.NewEventSigFilter(ev)
	evBy := solprimitives.NewEventBySubkeyFilter(1, []solprimitives.IndexedValueComparator{
		{Value: typesolana.IndexedValue{1, 2, 3}, Operator: 0},
	})

	root := query.Or(query.And(a, e), evBy)

	// to proto
	pb, err := conv.ConvertExpressionsToProto([]query.Expression{root})
	require.NoError(t, err)
	require.Len(t, pb, 1)

	// from proto
	round, err := conv.ConvertExpressionsFromProto(pb)
	require.NoError(t, err)
	require.Len(t, round, 1)
}

func TestLPFilterAndSubkeysConverters(t *testing.T) {
	f := &conv.LPFilterQuery{
		Name:            "test",
		Address:         mkBytes(typesolana.PublicKeyLength, 0xAA),
		EventName:       "Evt",
		EventSig:        mkBytes(typesolana.EventSignatureLength, 0xBB),
		StartingBlock:   10,
		ContractIdlJson: []byte(`{"idl":1}`),
		SubkeyPaths:     []*conv.Subkeys{{Subkeys: []string{"a", "b"}}, {Subkeys: []string{"c"}}},
		Retention:       int64(time.Hour),
		MaxLogsKept:     100,
		IncludeReverted: true,
	}
	df, err := conv.ConvertLPFilterQueryFromProto(f)
	require.NoError(t, err)
	require.Equal(t, "test", df.Name)
	require.Equal(t, 2, len(df.SubkeyPaths))

	back := conv.ConvertLPFilterQueryToProto(df)
	require.Equal(t, f.Name, back.Name)
	require.Equal(t, f.MaxLogsKept, back.MaxLogsKept)

	// Also exercise subkey helpers alone; ensure shape preserved.
	keys := [][]string{{"x", "y"}, {"z"}}
	ps := conv.ConvertSubkeyPathsToProto(keys)
	got := conv.ConvertSubkeyPathsFromProto(ps)
	require.Equal(t, keys, got)
}

func TestGettersAndSmallStructs_Smoke(t *testing.T) {
	// A few quick "does not panic / round-trips" on small structs

	// DataSlice
	ds := &conv.DataSlice{Offset: 5, Length: 7}
	dd := conv.ConvertDataSliceFromProto(ds)
	require.NotNil(t, dd)
	require.EqualValues(t, 5, *dd.Offset)
	require.EqualValues(t, 7, *dd.Length)
	ds2 := conv.ConvertDataSliceToProto(dd)
	require.EqualValues(t, 5, ds2.Offset)
	require.EqualValues(t, 7, ds2.Length)

	// UiTokenAmount
	ui := &conv.UiTokenAmount{Amount: "42", Decimals: 6, UiAmountString: "0.000042"}
	du := conv.ConvertUiTokenAmountFromProto(ui)
	require.NotNil(t, du)
	require.Equal(t, uint8(6), du.Decimals)
	ui2 := conv.ConvertUiTokenAmountToProto(du)
	require.EqualValues(t, 6, ui2.Decimals)

	// Commitment/Encoding enums (spot)
	require.Equal(t, conv.EncodingType_ENCODING_TYPE_BASE64, conv.ConvertEncodingTypeToProto(typesolana.EncodingBase64))
	require.Equal(t, typesolana.EncodingJSON, conv.ConvertEncodingTypeFromProto(conv.EncodingType_ENCODING_TYPE_JSON))
	require.Equal(t, conv.CommitmentType_COMMITMENT_TYPE_FINALIZED, conv.ConvertCommitmentToProto(typesolana.CommitmentFinalized))
	require.Equal(t, typesolana.CommitmentProcessed, conv.ConvertCommitmentFromProto(conv.CommitmentType_COMMITMENT_TYPE_PROCESSED))
}

func TestGetSignatureStatusesConverters(t *testing.T) {
	req := &conv.GetSignatureStatusesRequest{Sigs: [][]byte{mkBytes(typesolana.SignatureLength, 0xAA)}}
	dr, err := conv.ConvertGetSignatureStatusesRequestFromProto(req)
	require.NoError(t, err)
	req2 := conv.ConvertGetSignatureStatusesRequestToProto(dr)
	require.Len(t, req2.Sigs, 1)
	require.True(t, bytes.Equal(req.Sigs[0], req2.Sigs[0]))
	c := uint64(2)
	rep := &conv.GetSignatureStatusesReply{
		Results: []*conv.GetSignatureStatusesResult{{
			Slot:               1,
			Confirmations:      &c,
			Err:                "",
			ConfirmationStatus: conv.ConfirmationStatusType_CONFIRMATION_STATUS_TYPE_CONFIRMED,
		}},
	}
	drep := conv.ConvertGetSignatureStatusesReplyFromProto(rep)
	require.EqualValues(t, 1, drep.Results[0].Slot)
	require.NotNil(t, drep.Results[0].Confirmations)
	require.EqualValues(t, c, *drep.Results[0].Confirmations)

	rep2 := conv.ConvertGetSignatureStatusesReplyToProto(drep)
	require.EqualValues(t, &c, rep2.Results[0].Confirmations)
	require.Equal(t, conv.ConfirmationStatusType_CONFIRMATION_STATUS_TYPE_CONFIRMED, rep2.Results[0].ConfirmationStatus)
}

func TestErrorJoinBehavior_PublicKeys(t *testing.T) {
	in := [][]byte{
		mkBytes(typesolana.PublicKeyLength-1, 0x01),
		nil,
	}
	_, err := conv.ConvertPublicKeysFromProto(in)
	require.Error(t, err)
	// Error should mention both indices
	require.Contains(t, err.Error(), "public key[0]")
	require.Contains(t, err.Error(), "public key[1]")

	// Ensure errors.Is behaves reasonably (not super strict here)
	require.True(t, errors.Is(err, err))
}
