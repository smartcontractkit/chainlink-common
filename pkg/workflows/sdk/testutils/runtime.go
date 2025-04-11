package testutils

import (
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

type runtime[T any] struct {
	*runner[T]
}

func (r runtime[T]) IsNodeRuntime() {}

func (r runtime[T]) CallCapability(request *pb.CapabilityRequest) sdk.Promise[*pb.CapabilityResponse] {
	panic("implement me")
}

func (r runtime[T]) Config() []byte {
	//TODO implement me
	panic("implement me")
}

func (r runtime[T]) RunInNodeModeWithBuiltInConsensus(fn func(nodeRuntime sdk.NodeRuntime) *pb.BuiltInConsensusRequest) sdk.Promise[values.Value] {
	//TODO implement me
	panic("implement me")
}

var _ sdk.DonRuntime = &runtime[sdk.DonRuntime]{}
var _ sdk.NodeRuntime = &runtime[sdk.NodeRuntime]{}
