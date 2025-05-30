// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc, DO NOT EDIT.

package nodetriggermock

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/nodetrigger"

	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/testutils/registry"
)

// avoid unused imports
var _ = registry.Registry{}

func NewNodeEventCapability(t testing.TB) (*NodeEventCapability, error) {
	c := &NodeEventCapability{}
	reg := registry.GetRegistry(t)
	err := reg.RegisterCapability(c)
	return c, err
}

type NodeEventCapability struct {
	Trigger func(ctx context.Context, input *nodetrigger.Config) (*nodetrigger.Outputs, error)
}

func (cap *NodeEventCapability) Invoke(ctx context.Context, request *sdkpb.CapabilityRequest) *sdkpb.CapabilityResponse {
	capResp := &sdkpb.CapabilityResponse{}
	capResp.Response = &sdkpb.CapabilityResponse_Error{Error: fmt.Sprintf("method %s not found", request.Method)}
	return capResp
}

func (cap *NodeEventCapability) InvokeTrigger(ctx context.Context, request *sdkpb.TriggerSubscription) (*sdkpb.Trigger, error) {
	trigger := &sdkpb.Trigger{}
	switch request.Method {
	case "Trigger":
		input := &nodetrigger.Config{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			return nil, err
		}

		if cap.Trigger == nil {
			return nil, registry.ErrNoTriggerStub("Trigger")
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
		}
	default:
		return nil, fmt.Errorf("method %s not found", request.Method)
	}
	return trigger, nil
}

func (cap *NodeEventCapability) ID() string {
	return "basic-test-node-trigger@1.0.0"
}
