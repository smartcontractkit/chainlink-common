package sdk_test

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

func TestRunInNodeMode_SimpleConsensusType(t *testing.T) {
	runtime := &mockDonRuntime{}

	p := sdk.RunInNodeMode(runtime, func(nr sdk.NodeRuntime) (int, error) {
		return 42, nil
	}, pb.SimpleConsensusType_MEDIAN)

	val, err := p.Await()
	require.NoError(t, err)
	assert.Equal(t, 42, val)
}

func TestRunInNodeMode_PointerTypes(t *testing.T) {
	runtime := &mockDonRuntime{}
	p := sdk.RunInNodeMode(runtime, func(nr sdk.NodeRuntime) (*int, error) {
		val := 42
		return &val, nil
	}, pb.SimpleConsensusType_MEDIAN)

	val, err := p.Await()
	require.NoError(t, err)
	assert.Equal(t, 42, *val)
}

func TestRunInNodeMode_PrimitiveConsensusWithDefault(t *testing.T) {
	runtime := &mockDonRuntime{}

	p := sdk.RunInNodeMode(runtime, func(nr sdk.NodeRuntime) (int, error) {
		return 99, nil
	}, &sdk.PrimitiveConsensusWithDefault[int]{
		SimpleConsensusType: pb.SimpleConsensusType_IDENTICAL,
		DefaultValue:        123,
	})

	val, err := p.Await()
	require.NoError(t, err)
	assert.Equal(t, 99, val)
}

func TestRunInNodeMode_ErrorFromFunction(t *testing.T) {
	runtime := &mockDonRuntime{}

	p := sdk.RunInNodeMode(runtime, func(nr sdk.NodeRuntime) (int, error) {
		return 0, errors.New("some error")
	}, pb.SimpleConsensusType_MEDIAN)

	_, err := p.Await()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "some error")
}

func TestRunInNodeMode_ErrorWrappingResult(t *testing.T) {
	runtime := &mockDonRuntime{}

	type unsupported struct {
		Test chan int
	}
	p := sdk.RunInNodeMode(runtime, func(nr sdk.NodeRuntime) (*unsupported, error) {
		return &unsupported{Test: make(chan int)}, nil
	}, pb.SimpleConsensusType_MEDIAN)

	_, err := p.Await()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not wrap into value:")
}

func TestRunInNodeMode_ErrorWrappingDefault(t *testing.T) {
	runtime := &mockDonRuntime{}

	type unsupported struct {
		Test chan int
	}

	p := sdk.RunInNodeMode(runtime, func(nr sdk.NodeRuntime) (*unsupported, error) {
		return nil, errors.New("some error")
	}, &sdk.PrimitiveConsensusWithDefault[*unsupported]{
		SimpleConsensusType: pb.SimpleConsensusType_MEDIAN,
		DefaultValue:        &unsupported{Test: make(chan int)},
	})

	_, err := p.Await()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not wrap into value:")
}

// mockNodeRuntime implements NodeRuntime for testing.
type mockNodeRuntime struct{}

func (m mockNodeRuntime) CallCapability(request *pb.CapabilityRequest) sdk.Promise[*pb.CapabilityResponse] {
	panic("unused in tests")
}

func (m mockNodeRuntime) Config() []byte {
	panic("unused in tests")
}

func (m mockNodeRuntime) LogWriter() io.Writer {
	panic("unused in tests")
}

func (m mockNodeRuntime) IsNodeRuntime() {}

type mockDonRuntime struct {
	requests []*pb.BuiltInConsensusRequest
}

func (m *mockDonRuntime) RunInNodeMode(fn func(sdk.NodeRuntime) *pb.BuiltInConsensusRequest) sdk.Promise[values.Value] {
	req := fn(mockNodeRuntime{})
	m.requests = append(m.requests, req)

	if errObs, ok := req.Observation.(*pb.BuiltInConsensusRequest_Error); ok {
		return sdk.PromiseFromResult[values.Value](nil, errors.New(errObs.Error))
	}
	val, _ := values.FromProto(req.Observation.(*pb.BuiltInConsensusRequest_Value).Value)
	return sdk.PromiseFromResult(val, nil)
}

func (m *mockDonRuntime) CallCapability(*pb.CapabilityRequest) sdk.Promise[*pb.CapabilityResponse] {
	panic("not used in test")
}
func (m *mockDonRuntime) Config() []byte       { return nil }
func (m *mockDonRuntime) LogWriter() io.Writer { return nil }
