package testutils

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

type Capability interface {
	// TODO config if needed for register and unregister

	Invoke(ctx context.Context, request *pb.CapabilityRequest) *pb.CapabilityResponse
	InvokeTrigger(ctx context.Context, request *pb.TriggerSubscriptionRequest) (*pb.Trigger, error)
	ID() string
}

type NoTriggerStub string

func (n NoTriggerStub) Error() string {
	return "Stub not implemented for trigger: " + string(n)
}

var _ error = NoTriggerStub("")
