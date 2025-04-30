package testutils

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

func NewFakeWrapper(tb testing.TB, capability capabilities.BaseCapability) (*FakeWrapper, error) {
	wrapper := &FakeWrapper{
		BaseCapability: capability,
		tb:             tb,
	}
	err := GetRegistry(tb).RegisterCapability(wrapper)
	return wrapper, err
}

type FakeWrapper struct {
	capabilities.BaseCapability
	tb testing.TB
}

func (f *FakeWrapper) Invoke(ctx context.Context, request *pb.CapabilityRequest) *pb.CapabilityResponse {
	executable, ok := f.BaseCapability.(capabilities.ExecutableCapability)
	if !ok {
		return &pb.CapabilityResponse{
			Response: &pb.CapabilityResponse_Error{
				Error: fmt.Sprintf("method %s not found", request.Method),
			},
		}
	}

	response, err := executable.Execute(ctx, capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{
			WorkflowExecutionID: request.ExecutionId,
		},
		Payload: request.Payload,
		// TODO: https://smartcontract-it.atlassian.net/browse/CAPPL-799 add the default to the call
		ConfigPayload: nil,
		Method:        request.Method,
	})

	if err != nil {
		return &pb.CapabilityResponse{Response: &pb.CapabilityResponse_Error{Error: err.Error()}}
	}

	return &pb.CapabilityResponse{Response: &pb.CapabilityResponse_Payload{Payload: response.Payload}}
}

func (f *FakeWrapper) InvokeTrigger(ctx context.Context, request *pb.TriggerSubscription) (*pb.Trigger, error) {
	trigger, ok := f.BaseCapability.(capabilities.TriggerCapability)
	if !ok {
		return nil, fmt.Errorf("method %s not found", request.Method)
	}

	register := capabilities.TriggerRegistrationRequest{
		TriggerID: request.Id,
		Metadata: capabilities.RequestMetadata{
			WorkflowExecutionID: request.ExecId,
		},
		Payload: request.Payload,
		Method:  request.Method,
	}
	ch, err := trigger.RegisterTrigger(ctx, register)
	if err != nil {
		return nil, err
	}

	response, ok := <-ch

	// Fake isn't returning a trigger to run
	if !ok {
		return nil, nil
	}

	if response.Err != nil {
		return nil, response.Err
	}

	return &pb.Trigger{
		// TODO: https://smartcontract-it.atlassian.net/browse/CAPPL-791 multiple of the same trigger registered
		Id:      request.Id,
		Payload: request.Payload,
	}, nil
}

func (f *FakeWrapper) ID() string {
	info, err := f.Info(f.tb.Context())
	require.NoError(f.tb, err)
	return info.ID
}

var _ Capability = (*FakeWrapper)(nil)
