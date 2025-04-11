package testutils

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

type Capability interface {
	// TODO config if needed for register and unregister

	Invoke(ctx context.Context, request *pb.CapabilityRequest) (<-chan *pb.CapabilityResponse, error)
	ID() string
}

type Trigger interface {
	Trigger(ctx context.Context, request *pb.TriggerSubscriptionRequest) (*pb.Trigger, error)
	ID() string
	// TODO unregister
}
