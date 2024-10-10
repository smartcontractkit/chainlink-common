package ocr3_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/consensus/ocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/values/pb"
)

func Test_ValuesEncoder_Encode(t *testing.T) {
	t.Parallel()
	input := map[string]any{
		"foo": "bar",
		"baz": int64(42),
		"x":   map[string]any{"y": "z"},
	}
	inputWrapped, err := values.NewMap(input)
	require.NoError(t, err)

	expectedProto := &pb.Value{
		Value: &pb.Value_MapValue{
			MapValue: &pb.Map{
				Fields: map[string]*pb.Value{
					"foo": {Value: &pb.Value_StringValue{StringValue: "bar"}},
					"baz": {Value: &pb.Value_Int64Value{Int64Value: 42}},
					"x": {
						Value: &pb.Value_MapValue{
							MapValue: &pb.Map{
								Fields: map[string]*pb.Value{
									"y": {Value: &pb.Value_StringValue{StringValue: "z"}},
								},
							},
						},
					},
				},
			},
		},
	}

	encoder := ocr3.ValueMapEncoder{}
	actual, err := encoder.Encode(tests.Context(t), *inputWrapped)
	require.NoError(t, err)

	opts := proto.MarshalOptions{Deterministic: true}
	expected, err := opts.Marshal(expectedProto)
	require.NoError(t, err)

	assert.Equal(t, expected, actual)

	decoded := &pb.Value{}
	require.NoError(t, proto.Unmarshal(actual, decoded))

	val, err := values.FromProto(decoded)
	require.NoError(t, err)

	output := map[string]any{}
	require.NoError(t, val.UnwrapTo(&output))

	assert.Equal(t, input, output)
}
