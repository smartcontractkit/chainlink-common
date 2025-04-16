package testutils

import (
	"errors"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

type runtime[T any] struct {
	*runner[T]
	callErr error
}

func (r *runtime[T]) IsNodeRuntime() {}

func (r *runtime[T]) CallCapability(request *pb.CapabilityRequest) sdk.Promise[*pb.CapabilityResponse] {
	if r.callErr != nil {
		return sdk.PromiseFromResult((*pb.CapabilityResponse)(nil), r.callErr)
	}

	capability, ok := r.runner.registry.capabilities[request.Id]
	if !ok {
		return sdk.PromiseFromResult((*pb.CapabilityResponse)(nil), NoCapability(request.Id))
	}

	response := make(chan *pb.CapabilityResponse, 1)
	go func() {
		response <- capability.Invoke(r.ctx, request)
	}()

	return sdk.NewBasicPromise(func() (*pb.CapabilityResponse, error) {
		return <-response, nil
	})
}

func (r *runtime[T]) RunInNodeModeWithBuiltInConsensus(fn func(nodeRuntime sdk.NodeRuntime) *pb.BuiltInConsensusRequest) sdk.Promise[values.Value] {
	r.callErr = sdk.DonModeCallInNodeMode()
	nrt := r.nodeRunner().runtime
	result := fn(nrt)
	nrt.(*runtime[sdk.NodeRuntime]).callErr = sdk.NodeModeCallInDonMode()
	r.callErr = nil

	observation := result.Observation
	switch o := observation.(type) {
	case *pb.BuiltInConsensusRequest_Value:
		return sdk.PromiseFromResult(values.FromProto(o.Value))
	case *pb.BuiltInConsensusRequest_Error:
		if result.DefaultValue.Value == nil {
			return sdk.PromiseFromResult(values.Value(nil), errors.New(o.Error))
		}

		return sdk.PromiseFromResult(values.FromProto(result.DefaultValue))
	}

	return sdk.PromiseFromResult(values.Value(nil), errors.New("should not get here"))
}

var _ sdk.DonRuntime = &runtime[sdk.DonRuntime]{}
var _ sdk.NodeRuntime = &runtime[sdk.NodeRuntime]{}
