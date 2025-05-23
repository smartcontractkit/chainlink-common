package registry

import (
	"context"
	"errors"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	capabilitesbp "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

type CapabilityWrapper struct {
	Capability
}

var _ capabilities.ExecutableAndTriggerCapability = (*CapabilityWrapper)(nil)

func (c *CapabilityWrapper) RegisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	ch := make(chan capabilities.TriggerResponse, 1)
	trigger, err := c.InvokeTrigger(ctx, &pb.TriggerSubscription{
		Id:      request.TriggerID,
		Payload: request.Payload,
		Method:  request.Method,
	})

	response := capabilities.TriggerResponse{}
	if err != nil {
		response.Err = err
	} else if trigger == nil {
		return nil, nil
	} else {
		response.Event = capabilities.TriggerEvent{
			TriggerType: request.TriggerID,
			Payload:     trigger.Payload,
		}
	}

	ch <- response
	close(ch)
	return ch, nil
}

func (c *CapabilityWrapper) UnregisterTrigger(_ context.Context, _ capabilities.TriggerRegistrationRequest) error {
	return nil
}

func (c *CapabilityWrapper) RegisterToWorkflow(_ context.Context, _ capabilities.RegisterToWorkflowRequest) error {
	return nil
}

func (c *CapabilityWrapper) UnregisterFromWorkflow(_ context.Context, _ capabilities.UnregisterFromWorkflowRequest) error {
	return nil
}

func (c *CapabilityWrapper) Execute(ctx context.Context, request capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	v1Request := capabilitesbp.CapabilityRequestToProto(request)
	v2Request := &pb.CapabilityRequest{
		Id:      v1Request.Metadata.ReferenceId,
		Payload: v1Request.Payload,
		Method:  v1Request.Method,
	}

	v2Response := c.Invoke(ctx, v2Request)
	switch r := v2Response.Response.(type) {
	case *pb.CapabilityResponse_Error:
		return capabilities.CapabilityResponse{}, errors.New(r.Error)
	case *pb.CapabilityResponse_Payload:
		return capabilities.CapabilityResponse{
			Payload: r.Payload,
		}, nil
	default:
		return capabilities.CapabilityResponse{}, fmt.Errorf("unknown capability response type: %T", r)
	}
}

func (c *CapabilityWrapper) Info(_ context.Context) (capabilities.CapabilityInfo, error) {
	return capabilities.NewCapabilityInfo(
		c.ID(), capabilities.CapabilityTypeCombined, fmt.Sprintf("Mock of capability %s", c.ID()))
}
