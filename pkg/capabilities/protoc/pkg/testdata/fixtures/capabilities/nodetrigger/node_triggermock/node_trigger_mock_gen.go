// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc, DO NOT EDIT.

package nodetriggermock

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/nodetrigger"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testutils"
)

// avoid unused imports
var _ = capabilities.CapabilityInfo{}
var _ = testutils.Registry{}

type NodeEventCapability struct {
	// TODO teardown with unrgister if register is needed, or allow setup and teardown

	Trigger func(ctx context.Context, input *nodetrigger.Config) (*nodetrigger.Outputs, error)
}

func (cap *NodeEventCapability) Invoke(ctx context.Context, request *pb.CapabilityRequest) *pb.CapabilityResponse {
	capResp := &pb.CapabilityResponse{}
	capResp.Response = &pb.CapabilityResponse_Error{Error: fmt.Sprintf("method %s not found", request.Method)}
	return capResp
}

func (cap *NodeEventCapability) InvokeTrigger(ctx context.Context, request *pb.TriggerSubscriptionRequest) (*pb.Trigger, error) {
	trigger := &pb.Trigger{}
	switch request.Method {
	case "Trigger":
		input := &nodetrigger.Config{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			return nil, err
		}

		if cap.Trigger == nil {
			return nil, testutils.NoTriggerStub("Trigger")
		}

		resp, err := cap.Trigger(ctx, input)
		if err != nil {
			return nil, err
		} else {
			if resp == nil {
				return nil, nil
			}

			payload, err := anypb.New(resp)
			if err != nil {
				return nil, err
			}
			trigger.Payload = payload
			trigger.Id = "mock"
		}
	default:
		return nil, fmt.Errorf("method %s not found", request.Method)
	}
	return trigger, nil
}

func (cap *NodeEventCapability) ID() string {
	return "basic-test-trigger@1.0.0"
}
