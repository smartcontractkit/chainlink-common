package testutils

import (
	"errors"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

type runtime[T any] struct {
	*runner[T]
}

func (r runtime[T]) IsNodeRuntime() {}

func (r runtime[T]) CallCapability(request *pb.CapabilityRequest) sdk.Promise[*pb.CapabilityResponse] {
	capability, ok := r.runner.registry.capabilities[request.Id]
	if !ok {
		return sdk.PromiseFromResult((*pb.CapabilityResponse)(nil), fmt.Errorf("capability %s not found", request.Id))
	}

	response := make(chan *pb.CapabilityResponse, 1)
	go func() {
		response <- capability.Invoke(r.ctx, request)
	}()

	return sdk.NewBasicPromise(func() (*pb.CapabilityResponse, error) {
		return <-response, nil
	})
}

func (r runtime[T]) RunInNodeModeWithBuiltInConsensus(fn func(nodeRuntime sdk.NodeRuntime) *pb.BuiltInConsensusRequest) sdk.Promise[values.Value] {
	result := fn(r.nodeRunner().runtime)

	observation := result.Observation
	if observation == nil && result.DefaultValue != nil {
		return sdk.PromiseFromResult(values.Value(nil), nil)
	}

	switch o := observation.(type) {
	case *pb.BuiltInConsensusRequest_Value:
		value, err := values.FromProto(o.Value)
		return sdk.PromiseFromResult(value, err)
	case *pb.BuiltInConsensusRequest_Error:
		return sdk.PromiseFromResult(values.Value(nil), errors.New(o.Error))
	}

	return sdk.PromiseFromResult(values.Value(nil), errors.New("should not get here"))
}

var _ sdk.DonRuntime = &runtime[sdk.DonRuntime]{}
var _ sdk.NodeRuntime = &runtime[sdk.NodeRuntime]{}
