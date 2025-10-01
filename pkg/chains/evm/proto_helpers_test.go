package evm_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	chain_common "github.com/smartcontractkit/chainlink-common/pkg/loop/chain-common"
	evmtypes "github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
)

func mkBytes(n int, fill byte) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = fill
	}
	return b
}

func TestAddressConverters(t *testing.T) {
	zero := make([]byte, evmtypes.AddressLength)

	t.Run("ConvertAddressFromProto", func(t *testing.T) {
		cases := []struct {
			name            string
			in              []byte
			wantErrParts    []string
			wantEqualsInput bool
		}{
			{"accepts 20-byte address", mkBytes(evmtypes.AddressLength, 0xAB), nil, true},
			{"rejects nil address", nil, []string{"address can't be nil"}, false},
			{"rejects wrong length and shows hex", mkBytes(evmtypes.AddressLength-1, 0x01), []string{"invalid address", "expected", "value=0x"}, false},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := evm.ConvertAddressFromProto(tc.in)
				if tc.wantErrParts != nil {
					if err == nil {
						t.Fatalf("expected error, got nil")
					}
					for _, p := range tc.wantErrParts {
						if !strings.Contains(err.Error(), p) {
							t.Fatalf("error missing %q, got: %v", p, err)
						}
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if tc.wantEqualsInput && !bytes.Equal(got[:], tc.in) {
					t.Fatalf("mismatch: got=%x want=%x", got[:], tc.in)
				}
			})
		}
	})

	t.Run("ConvertOptionalAddressFromProto", func(t *testing.T) {
		cases := []struct {
			name         string
			in           []byte
			wantZero     bool
			wantErrParts []string
		}{
			{"nil input becomes zero address", nil, true, nil},
			{"empty input becomes zero address", []byte{}, true, nil},
			{"rejects wrong length", mkBytes(evmtypes.AddressLength-2, 0xFF), false, []string{"invalid address"}},
			{"accepts 20-byte address", mkBytes(evmtypes.AddressLength, 0x11), false, nil},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := evm.ConvertOptionalAddressFromProto(tc.in)
				if tc.wantErrParts != nil {
					if err == nil {
						t.Fatalf("expected error, got nil")
					}
					for _, p := range tc.wantErrParts {
						if !strings.Contains(err.Error(), p) {
							t.Fatalf("error missing %q, got: %v", p, err)
						}
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if tc.wantZero {
					if !bytes.Equal(got[:], zero) {
						t.Fatalf("want zero address, got %x", got[:])
					}
				} else if !bytes.Equal(got[:], tc.in) {
					t.Fatalf("mismatch: got=%x want=%x", got[:], tc.in)
				}
			})
		}
	})

	t.Run("Roundtrip", func(t *testing.T) {
		addrs := []evmtypes.Address{
			evmtypes.Address(mkBytes(evmtypes.AddressLength, 0x01)),
			evmtypes.Address(mkBytes(evmtypes.AddressLength, 0x02)),
		}
		proto := evm.ConvertAddressesToProto(addrs)
		got, err := evm.ConvertAddressesFromProto(proto)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if len(got) != len(addrs) {
			t.Fatalf("len mismatch: got=%d want=%d", len(got), len(addrs))
		}
		for i := range addrs {
			if !bytes.Equal(got[i][:], addrs[i][:]) {
				t.Fatalf("idx %d mismatch: got=%x want=%x", i, got[i][:], addrs[i][:])
			}
		}
	})

	t.Run("AggregatesErrors", func(t *testing.T) {
		in := [][]byte{
			mkBytes(evmtypes.AddressLength, 0x01),
			mkBytes(evmtypes.AddressLength-1, 0x02),
			nil,
		}
		got, err := evm.ConvertAddressesFromProto(in)
		if got != nil {
			t.Fatalf("want nil result on aggregate error, got: %#v", got)
		}
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		msg := err.Error()
		if !strings.Contains(msg, "index 1") || !strings.Contains(msg, "index 2") {
			t.Fatalf("joined error should include both failing indices, got: %v", msg)
		}
	})

	t.Run("WrapsErrorsWithContext", func(t *testing.T) {
		in := [][]byte{mkBytes(evmtypes.AddressLength-1, 0x01)}
		_, err := evm.ConvertAddressesFromProto(in)
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), "failed to convert address at index 0") {
			t.Fatalf("missing context wrapper, got: %v", err)
		}
		if !errors.Is(err, err) {
			t.Fatalf("unexpected errors.Is behavior")
		}
	})
}

func TestHashConverters(t *testing.T) {
	zero := make([]byte, evmtypes.HashLength)

	t.Run("ConvertHashFromProto", func(t *testing.T) {
		cases := []struct {
			name            string
			in              []byte
			wantErrParts    []string
			wantEqualsInput bool
		}{
			{"accepts 32-byte hash", mkBytes(evmtypes.HashLength, 0xCD), nil, true},
			{"rejects nil hash", nil, []string{"hash can't be nil"}, false},
			{"rejects wrong length and shows hex", mkBytes(evmtypes.HashLength+1, 0x02), []string{"invalid hash", "expected", "value=0x"}, false},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := evm.ConvertHashFromProto(tc.in)
				if tc.wantErrParts != nil {
					if err == nil {
						t.Fatalf("expected error, got nil")
					}
					for _, p := range tc.wantErrParts {
						if !strings.Contains(err.Error(), p) {
							t.Fatalf("error missing %q, got: %v", p, err)
						}
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if tc.wantEqualsInput && !bytes.Equal(got[:], tc.in) {
					t.Fatalf("mismatch: got=%x want=%x", got[:], tc.in)
				}
			})
		}
	})

	t.Run("ConvertOptionalHashFromProto", func(t *testing.T) {
		cases := []struct {
			name         string
			in           []byte
			wantZero     bool
			wantErrParts []string
		}{
			{"nil input becomes zero hash", nil, true, nil},
			{"empty input becomes zero hash", []byte{}, true, nil},
			{"rejects wrong length", mkBytes(evmtypes.HashLength-1, 0xEE), false, []string{"invalid hash"}},
			{"accepts 32-byte hash", mkBytes(evmtypes.HashLength, 0xAA), false, nil},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				got, err := evm.ConvertOptionalHashFromProto(tc.in)
				if tc.wantErrParts != nil {
					if err == nil {
						t.Fatalf("expected error, got nil")
					}
					for _, p := range tc.wantErrParts {
						if !strings.Contains(err.Error(), p) {
							t.Fatalf("error missing %q, got: %v", p, err)
						}
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if tc.wantZero {
					if !bytes.Equal(got[:], zero) {
						t.Fatalf("want zero hash, got %x", got[:])
					}
				} else if !bytes.Equal(got[:], tc.in) {
					t.Fatalf("mismatch: got=%x want=%x", got[:], tc.in)
				}
			})
		}
	})

	t.Run("Roundtrip", func(t *testing.T) {
		hashes := []evmtypes.Hash{
			evmtypes.Hash(mkBytes(evmtypes.HashLength, 0xAA)),
			evmtypes.Hash(mkBytes(evmtypes.HashLength, 0xBB)),
		}
		proto := evm.ConvertHashesToProto(hashes)
		got, err := evm.ConvertHashesFromProto(proto)
		if err != nil {
			t.Fatalf("unexpected err: %v", err)
		}
		if len(got) != len(hashes) {
			t.Fatalf("len mismatch: got=%d want=%d", len(got), len(hashes))
		}
		for i := range hashes {
			if !bytes.Equal(got[i][:], hashes[i][:]) {
				t.Fatalf("idx %d mismatch: got=%x want=%x", i, got[i][:], hashes[i][:])
			}
		}
	})

	t.Run("AggregatesErrors", func(t *testing.T) {
		in := [][]byte{
			mkBytes(evmtypes.HashLength-1, 0x03),
			nil,
		}
		got, err := evm.ConvertHashesFromProto(in)
		if got != nil {
			t.Fatalf("want nil result on aggregate error, got: %#v", got)
		}
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "index 0") || !strings.Contains(err.Error(), "index 1") {
			t.Fatalf("joined error should include failing indices, got: %v", err)
		}
	})
}

func TestConvertExpressionsFromProto(t *testing.T) {
	testCases := []struct {
		Name           string
		In             []*evm.Expression
		ExpectedResult []query.Expression
		ExpectedError  string
	}{
		{
			Name:           "empty",
			ExpectedResult: []query.Expression{},
		},
		{
			Name: "Empty evaluator",
			In: []*evm.Expression{
				{},
			},
			ExpectedError: "rpc error: code = InvalidArgument desc = err to convert expr idx 0 err: unknown expression type: <nil>",
		},
		{
			Name: "Empty Expression_Primitive",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_Primitive{},
				},
			},
			ExpectedError: "rpc error: code = InvalidArgument desc = err to convert expr idx 0 err: unknown primitive type: <nil>",
		},
		{
			Name: "Empty Expression_BooleanExpression",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_BooleanExpression{},
				},
			},
			ExpectedResult: []query.Expression{
				{
					BoolExpression: query.BoolExpression{},
				},
			},
		},
		{
			Name: "Nested empty Expression",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_BooleanExpression{
						BooleanExpression: &evm.BooleanExpression{
							Expression: []*evm.Expression{
								{},
							},
						},
					},
				},
			},
			ExpectedError: "err to convert expr idx 0 err: failed to convert sub-expression 0: unknown expression type: <nil>",
		},
		{
			Name: "Empty Evaluator.Primitive",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_Primitive{
						Primitive: nil,
					},
				},
			},
			ExpectedError: "rpc error: code = InvalidArgument desc = err to convert expr idx 0 err: unknown primitive type: <nil>",
		},
		{
			Name: "Empty Evaluator.Primitive.Primitive",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_Primitive{
						Primitive: &evm.Primitive{},
					},
				},
			},
			ExpectedError: "rpc error: code = InvalidArgument desc = err to convert expr idx 0 err: unknown primitive type: <nil>",
		},
		{
			Name: "Empty Evaluator.Primitive.Primitive.GeneralPrimitive",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_Primitive{
						Primitive: &evm.Primitive{
							Primitive: &evm.Primitive_GeneralPrimitive{
								GeneralPrimitive: nil,
							},
						},
					},
				},
			},
			ExpectedError: "rpc error: code = InvalidArgument desc = err to convert expr idx 0 err: rpc error: code = InvalidArgument desc = primitive can not be nil",
		},
		{
			Name: "Empty Evaluator.Primitive.Primitive.GeneralPrimitive.Primitive",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_Primitive{
						Primitive: &evm.Primitive{
							Primitive: &evm.Primitive_GeneralPrimitive{
								GeneralPrimitive: &chain_common.Primitive{},
							},
						},
					},
				},
			},
			ExpectedError: "rpc error: code = InvalidArgument desc = err to convert expr idx 0 err: rpc error: code = InvalidArgument desc = unknown primitive type: <nil>",
		},
		{
			Name: "Empty Evaluator.Primitive.Primitive.GeneralPrimitive.Primitive.Comparator",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_Primitive{
						Primitive: &evm.Primitive{
							Primitive: &evm.Primitive_GeneralPrimitive{
								GeneralPrimitive: &chain_common.Primitive{
									Primitive: &chain_common.Primitive_Comparator{},
								},
							},
						},
					},
				},
			},
			ExpectedError: "comparator can not be nil",
		},
		{
			Name: "Empty Evaluator.Primitive.Primitive.GeneralPrimitive.Primitive.Comparator.ValueComparator",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_Primitive{
						Primitive: &evm.Primitive{
							Primitive: &evm.Primitive_GeneralPrimitive{
								GeneralPrimitive: &chain_common.Primitive{
									Primitive: &chain_common.Primitive_Comparator{
										Comparator: &chain_common.Comparator{
											ValueComparators: []*chain_common.ValueComparator{nil},
										},
									},
								},
							},
						},
					},
				},
			},
			ExpectedError: "unsupported primitive type: *evm.Primitive_GeneralPrimitive",
		},
		{
			Name: "Empty Evaluator.Primitive.Primitive.GeneralPrimitive.Primitive.Block",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_Primitive{
						Primitive: &evm.Primitive{
							Primitive: &evm.Primitive_GeneralPrimitive{
								GeneralPrimitive: &chain_common.Primitive{
									Primitive: &chain_common.Primitive_Block{},
								},
							},
						},
					},
				},
			},
			ExpectedError: "Block can not be nil",
		},
		{
			Name: "Empty Evaluator.Primitive.Primitive.GeneralPrimitive.Primitive.TxHash",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_Primitive{
						Primitive: &evm.Primitive{
							Primitive: &evm.Primitive_GeneralPrimitive{
								GeneralPrimitive: &chain_common.Primitive{
									Primitive: &chain_common.Primitive_TxHash{},
								},
							},
						},
					},
				},
			},
			ExpectedError: "TxHash can not be nil",
		},
		{
			Name: "Empty Evaluator.Primitive.Primitive.GeneralPrimitive.Primitive.Timestamp",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_Primitive{
						Primitive: &evm.Primitive{
							Primitive: &evm.Primitive_GeneralPrimitive{
								GeneralPrimitive: &chain_common.Primitive{
									Primitive: &chain_common.Primitive_Timestamp{},
								},
							},
						},
					},
				},
			},
			ExpectedError: "Timestamp can not be nil",
		},
		{
			Name: "Invalid Evaluator.Primitive.Primitive.ContractAddress",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_Primitive{
						Primitive: &evm.Primitive{
							Primitive: &evm.Primitive_ContractAddress{},
						},
					},
				},
			},
			ExpectedError: "address can't be nil",
		},
		{
			Name: "Invalid Evaluator.Primitive.Primitive.ContractAddress",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_Primitive{
						Primitive: &evm.Primitive{
							Primitive: &evm.Primitive_EventSig{},
						},
					},
				},
			},
			ExpectedError: "failed to convert event sig",
		},
		{
			Name: "Empty Evaluator.Primitive.Primitive.EventByTopic",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_Primitive{
						Primitive: &evm.Primitive{
							Primitive: &evm.Primitive_EventByTopic{},
						},
					},
				},
			},
			ExpectedError: "EventByTopic can not be nil",
		},
		{
			Name: "Invalid Evaluator.Primitive.Primitive.EventByTopic.HashedValueComparers",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_Primitive{
						Primitive: &evm.Primitive{
							Primitive: &evm.Primitive_EventByTopic{
								EventByTopic: &evm.EventByTopic{
									HashedValueComparers: []*evm.HashValueComparator{nil},
								},
							},
						},
					},
				},
			},
			ExpectedError: "failed to convert EventByTopic hashed value comparators: hashed value comparator can't be nil",
		},
		{
			Name: "Empty Evaluator.Primitive.Primitive.EventByWord",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_Primitive{
						Primitive: &evm.Primitive{
							Primitive: &evm.Primitive_EventByWord{
								EventByWord: nil,
							},
						},
					},
				},
			},
			ExpectedError: "EventByWord can not be nil",
		},
		{
			Name: "Invalid Empty Evaluator.Primitive.Primitive.EventByWord.HashedValueComparers",
			In: []*evm.Expression{
				{
					Evaluator: &evm.Expression_Primitive{
						Primitive: &evm.Primitive{
							Primitive: &evm.Primitive_EventByWord{
								EventByWord: &evm.EventByWord{
									HashedValueComparers: []*evm.HashValueComparator{nil},
								},
							},
						},
					},
				},
			},
			ExpectedError: "failed to convert EventByWord hashed value comparators: hashed value comparator can't be nil",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			got, err := evm.ConvertExpressionsFromProto(tc.In)
			if tc.ExpectedError != "" {
				require.ErrorContains(t, err, tc.ExpectedError)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.ExpectedResult, got)
		})
	}
}
