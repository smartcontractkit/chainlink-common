package testutils

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

type Capability interface {
	Invoke(ctx context.Context, request *pb.CapabilityRequest) *pb.CapabilityResponse
	InvokeTrigger(ctx context.Context, request *pb.TriggerSubscription) (*pb.Trigger, error)
	ID() string
}
