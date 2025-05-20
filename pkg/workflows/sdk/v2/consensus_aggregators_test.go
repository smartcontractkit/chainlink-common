package sdk_test

import (
	"math/big"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestConsensusMedianAggregation(t *testing.T) {
	descriptor := sdk.ConsensusMedianAggregation[int]()
	require.NoError(t, descriptor.Err())
	assert.Equal(t, descriptor.Descriptor().GetAggregation(), pb.AggregationType_MEDIAN)
}

func TestConsensusIdenticalAggregation(t *testing.T) {
	t.Run("valid types", func(t *testing.T) {
		descriptor := sdk.ConsensusIdenticalAggregation[int]()
		require.NoError(t, descriptor.Err())
		assert.Equal(t, descriptor.Descriptor().GetAggregation(), pb.AggregationType_IDENTICAL)
	})

	t.Run("invalid types", func(t *testing.T) {
		descriptor := sdk.ConsensusIdenticalAggregation[chan int]()
		require.Error(t, descriptor.Err())
	})
}

func TestConsensusCommonPrefixAggregation(t *testing.T) {
	t.Run("valid primitive types", func(t *testing.T) {
		descriptor, err := sdk.ConsensusCommonPrefixAggregation[string]()()
		require.NoError(t, err)
		assert.Equal(t, descriptor.Descriptor().GetAggregation(), pb.AggregationType_COMMON_PREFIX)
	})

	t.Run("invalid primitive types", func(t *testing.T) {
		_, err := sdk.ConsensusCommonPrefixAggregation[[]chan int]()()
		require.Error(t, err)
	})
}

func TestConsensusCommonSuffixAggregation(t *testing.T) {
	t.Run("valid primitive types", func(t *testing.T) {
		descriptor, err := sdk.ConsensusCommonSuffixAggregation[string]()()
		require.NoError(t, err)
		assert.Equal(t, descriptor.Descriptor().GetAggregation(), pb.AggregationType_COMMON_SUFFIX)
	})

	t.Run("invalid primitive types", func(t *testing.T) {
		_, err := sdk.ConsensusCommonSuffixAggregation[[]chan int]()()
		require.Error(t, err)
	})
}

func TestConsensusAggregationFromTags(t *testing.T) {
	t.Run("valid median - all numeric types", func(t *testing.T) {
		t.Run("int", func(t *testing.T) { testMedianField[int](t) })
		t.Run("int8", func(t *testing.T) { testMedianField[int8](t) })
		t.Run("int16", func(t *testing.T) { testMedianField[int16](t) })
		t.Run("int32", func(t *testing.T) { testMedianField[int32](t) })
		t.Run("int64", func(t *testing.T) { testMedianField[int64](t) })
		t.Run("uint", func(t *testing.T) { testMedianField[uint](t) })
		t.Run("uint8", func(t *testing.T) { testMedianField[uint8](t) })
		t.Run("uint16", func(t *testing.T) { testMedianField[uint16](t) })
		t.Run("uint32", func(t *testing.T) { testMedianField[uint32](t) })
		t.Run("uint64", func(t *testing.T) { testMedianField[uint64](t) })
		t.Run("float32", func(t *testing.T) { testMedianField[float32](t) })
		t.Run("float64", func(t *testing.T) { testMedianField[float64](t) })
		t.Run("*big.Int", func(t *testing.T) { testMedianField[*big.Int](t) })
		t.Run("decimal", func(t *testing.T) { testMedianField[decimal.Decimal](t) })
	})

	t.Run("valid identical", func(t *testing.T) {
		type S struct {
			Val   string    `consensus:"identical"`
			PVal  *string   `consensus:"identical"`
			Slice []string  `consensus:"identical"`
			Array [2]string `consensus:"identical"`
		}
		desc := sdk.ConsensusAggregationFromTags[S]()
		require.NoError(t, desc.Err())
		expected := &pb.ConsensusDescriptor{
			Descriptor_: &pb.ConsensusDescriptor_FieldsMap{
				FieldsMap: &pb.FieldsMap{
					Fields: map[string]*pb.ConsensusDescriptor{
						"Val": {
							Descriptor_: &pb.ConsensusDescriptor_Aggregation{
								Aggregation: pb.AggregationType_IDENTICAL,
							},
						},
						"PVal": {
							Descriptor_: &pb.ConsensusDescriptor_Aggregation{
								Aggregation: pb.AggregationType_IDENTICAL,
							},
						},
						"Slice": {
							Descriptor_: &pb.ConsensusDescriptor_Aggregation{
								Aggregation: pb.AggregationType_IDENTICAL,
							},
						},
						"Array": {
							Descriptor_: &pb.ConsensusDescriptor_Aggregation{
								Aggregation: pb.AggregationType_IDENTICAL,
							},
						},
					},
				},
			},
		}
		require.True(t, proto.Equal(desc.Descriptor(), expected))
	})

	t.Run("valid common prefix", func(t *testing.T) {
		type S struct {
			Val []string `consensus:"common_prefix"`
		}
		desc := sdk.ConsensusAggregationFromTags[S]()
		require.NoError(t, desc.Err())
		expected := &pb.ConsensusDescriptor{
			Descriptor_: &pb.ConsensusDescriptor_FieldsMap{
				FieldsMap: &pb.FieldsMap{
					Fields: map[string]*pb.ConsensusDescriptor{
						"Val": {
							Descriptor_: &pb.ConsensusDescriptor_Aggregation{
								Aggregation: pb.AggregationType_COMMON_PREFIX,
							},
						},
					},
				},
			},
		}
		require.True(t, proto.Equal(desc.Descriptor(), expected))
	})

	t.Run("valid common suffix", func(t *testing.T) {
		type S struct {
			Val [2]string `consensus:"common_suffix"`
		}
		desc := sdk.ConsensusAggregationFromTags[S]()
		require.NoError(t, desc.Err())
		expected := &pb.ConsensusDescriptor{
			Descriptor_: &pb.ConsensusDescriptor_FieldsMap{
				FieldsMap: &pb.FieldsMap{
					Fields: map[string]*pb.ConsensusDescriptor{
						"Val": {
							Descriptor_: &pb.ConsensusDescriptor_Aggregation{
								Aggregation: pb.AggregationType_COMMON_SUFFIX,
							},
						},
					},
				},
			},
		}
		require.True(t, proto.Equal(desc.Descriptor(), expected))
	})

	t.Run("valid nested", func(t *testing.T) {
		type Inner struct {
			Score int32 `consensus:"median"`
		}
		type Outer struct {
			In Inner `consensus:"nested"`
		}
		desc := sdk.ConsensusAggregationFromTags[Outer]()
		require.NoError(t, desc.Err())
		expected := &pb.ConsensusDescriptor{
			Descriptor_: &pb.ConsensusDescriptor_FieldsMap{
				FieldsMap: &pb.FieldsMap{
					Fields: map[string]*pb.ConsensusDescriptor{
						"In": {
							Descriptor_: &pb.ConsensusDescriptor_FieldsMap{
								FieldsMap: &pb.FieldsMap{
									Fields: map[string]*pb.ConsensusDescriptor{
										"Score": {
											Descriptor_: &pb.ConsensusDescriptor_Aggregation{
												Aggregation: pb.AggregationType_MEDIAN,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}
		require.True(t, proto.Equal(desc.Descriptor(), expected))
	})

	t.Run("valid identical nested", func(t *testing.T) {
		type Inner struct {
			Score int32
		}

		type Outer struct {
			In  Inner  `consensus:"identical"`
			PIn *Inner `consensus:"identical"`
		}
		desc := sdk.ConsensusAggregationFromTags[Outer]()
		require.NoError(t, desc.Err())
		expected := &pb.ConsensusDescriptor{
			Descriptor_: &pb.ConsensusDescriptor_FieldsMap{
				FieldsMap: &pb.FieldsMap{
					Fields: map[string]*pb.ConsensusDescriptor{
						"In": {
							Descriptor_: &pb.ConsensusDescriptor_Aggregation{
								Aggregation: pb.AggregationType_IDENTICAL,
							},
						},
						"PIn": {
							Descriptor_: &pb.ConsensusDescriptor_Aggregation{
								Aggregation: pb.AggregationType_IDENTICAL,
							},
						},
					},
				},
			},
		}
		require.True(t, proto.Equal(desc.Descriptor(), expected))
	})

	t.Run("invalid identical nested", func(t *testing.T) {
		type Inner struct {
			Ch chan int32 `consensus:"identical"`
		}

		type Outer struct {
			In Inner `consensus:"identical"`
		}
		desc := sdk.ConsensusAggregationFromTags[Outer]()
		require.Error(t, desc.Err())
	})

	t.Run("invalid nested field", func(t *testing.T) {
		type Inner struct {
			Ch chan int `consensus:"median"`
		}
		type Outer struct {
			In Inner `consensus:"nested"`
		}
		desc := sdk.ConsensusAggregationFromTags[Outer]()
		require.Error(t, desc.Err())
	})

	t.Run("valid pointer", func(t *testing.T) {
		type S struct {
			Val string `consensus:"identical"`
		}

		desc := sdk.ConsensusAggregationFromTags[*S]()
		require.NoError(t, desc.Err())
		expected := &pb.ConsensusDescriptor{
			Descriptor_: &pb.ConsensusDescriptor_FieldsMap{
				FieldsMap: &pb.FieldsMap{
					Fields: map[string]*pb.ConsensusDescriptor{
						"Val": {
							Descriptor_: &pb.ConsensusDescriptor_Aggregation{
								Aggregation: pb.AggregationType_IDENTICAL,
							},
						},
					},
				},
			},
		}
		require.True(t, proto.Equal(desc.Descriptor(), expected))
	})

	t.Run("invalid median", func(t *testing.T) {
		type S struct {
			Val string `consensus:"median"`
		}
		desc := sdk.ConsensusAggregationFromTags[S]()
		require.ErrorContains(t, desc.Err(), "not a numeric type")
	})

	t.Run("invalid not a struct", func(t *testing.T) {
		desc := sdk.ConsensusAggregationFromTags[int]()
		require.ErrorContains(t, desc.Err(), "expects a struct type")
	})

	t.Run("ignore fields", func(t *testing.T) {
		type S struct {
			Val                  string `consensus:"identical"`
			IgnoredField         string `consensus:"ignore"`
			IgnoredImplicitField string
		}
		desc := sdk.ConsensusAggregationFromTags[S]()
		require.NoError(t, desc.Err())
		expected := &pb.ConsensusDescriptor{
			Descriptor_: &pb.ConsensusDescriptor_FieldsMap{
				FieldsMap: &pb.FieldsMap{
					Fields: map[string]*pb.ConsensusDescriptor{
						"Val": {
							Descriptor_: &pb.ConsensusDescriptor_Aggregation{
								Aggregation: pb.AggregationType_IDENTICAL,
							},
						},
					},
				},
			},
		}
		require.True(t, proto.Equal(desc.Descriptor(), expected))
	})

	t.Run("invalid identical", func(t *testing.T) {
		t.Run("channel", func(t *testing.T) { testInvalidIdenticalField[chan string](t) })
		t.Run("non string key map", func(t *testing.T) { testInvalidIdenticalField[map[int]int](t) })
	})

	t.Run("common prefix for valid types", func(t *testing.T) {
		descriptor := sdk.ConsensusAggregationFromTags[struct {
			Val []int `consensus:"common_prefix"`
		}]()

		require.NoError(t, descriptor.Err())
		expected := &pb.ConsensusDescriptor{
			Descriptor_: &pb.ConsensusDescriptor_FieldsMap{
				FieldsMap: &pb.FieldsMap{
					Fields: map[string]*pb.ConsensusDescriptor{
						"Val": {
							Descriptor_: &pb.ConsensusDescriptor_Aggregation{
								Aggregation: pb.AggregationType_COMMON_PREFIX,
							},
						},
					},
				},
			},
		}

		require.True(t, proto.Equal(descriptor.Descriptor(), expected))
	})

	t.Run("common prefix invalid types", func(t *testing.T) {
		desc := sdk.ConsensusAggregationFromTags[struct {
			Val chan int `consensus:"common_prefix"`
		}]()

		require.Error(t, desc.Err())
	})

	t.Run("common suffix for valid types", func(t *testing.T) {
		descriptor := sdk.ConsensusAggregationFromTags[struct {
			Val []int `consensus:"common_suffix"`
		}]()

		require.NoError(t, descriptor.Err())
		expected := &pb.ConsensusDescriptor{
			Descriptor_: &pb.ConsensusDescriptor_FieldsMap{
				FieldsMap: &pb.FieldsMap{
					Fields: map[string]*pb.ConsensusDescriptor{
						"Val": {
							Descriptor_: &pb.ConsensusDescriptor_Aggregation{
								Aggregation: pb.AggregationType_COMMON_SUFFIX,
							},
						},
					},
				},
			},
		}

		require.True(t, proto.Equal(descriptor.Descriptor(), expected))
	})

	t.Run("common suffix invalid types", func(t *testing.T) {
		desc := sdk.ConsensusAggregationFromTags[struct {
			Val chan int `consensus:"common_suffix"`
		}]()

		require.Error(t, desc.Err())
	})

	t.Run("invalid tag", func(t *testing.T) {
		type Invalid struct {
			In int `consensus:"not_real"`
		}
		desc := sdk.ConsensusAggregationFromTags[Invalid]()
		require.Error(t, desc.Err())
	})
}

func testMedianField[T any](t *testing.T) {
	t.Helper()
	desc := sdk.ConsensusAggregationFromTags[struct {
		Val  T  `consensus:"median"`
		PVal *T `consensus:"median"`
	}]()
	require.NoError(t, desc.Err())
	expected := &pb.ConsensusDescriptor{
		Descriptor_: &pb.ConsensusDescriptor_FieldsMap{
			FieldsMap: &pb.FieldsMap{
				Fields: map[string]*pb.ConsensusDescriptor{
					"Val": {
						Descriptor_: &pb.ConsensusDescriptor_Aggregation{
							Aggregation: pb.AggregationType_MEDIAN,
						},
					},
					"PVal": {
						Descriptor_: &pb.ConsensusDescriptor_Aggregation{
							Aggregation: pb.AggregationType_MEDIAN,
						},
					},
				},
			},
		},
	}
	require.True(t, proto.Equal(desc.Descriptor(), expected))
}

func testInvalidIdenticalField[T any](t *testing.T) {
	t.Helper()
	testInvalidIdenticalFieldHelper[T](t)
	testInvalidIdenticalFieldHelper[*T](t)
}

func testInvalidIdenticalFieldHelper[T any](t *testing.T) {
	t.Helper()
	desc := sdk.ConsensusAggregationFromTags[struct {
		Val T `consensus:"identical"`
	}]()
	require.ErrorContains(t, desc.Err(), "field Val marked as identical but is not a valid type")
}
