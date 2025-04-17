// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc, DO NOT EDIT.

package actionandtriggermock

import (
	"context"
	"fmt"
	"testing"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc/pkg/testdata/fixtures/capabilities/actionandtrigger"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2/pb"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/testutils"
)

// avoid unused imports
var _ = capabilities.CapabilityInfo{}
var _ = testutils.Registry{}

func NewBasicCapability(t testing.TB) (*BasicCapability, error) {
	c := &BasicCapability{}
	registry := testutils.GetRegistry(t)
	err := registry.RegisterCapability(c)
	return c, err
}

type BasicCapability struct {
	// TODO teardown with unrgister if register is needed, or allow setup and teardown
	// TODO register if needed...
	Action func(ctx context.Context, input *actionandtrigger.Input /* TODO config? */) (*actionandtrigger.Output, error)

	Trigger func(ctx context.Context, input *actionandtrigger.Config) (*actionandtrigger.TriggerEvent, error)
}

func (cap *BasicCapability) Invoke(ctx context.Context, request *pb.CapabilityRequest) *pb.CapabilityResponse {
	capResp := &pb.CapabilityResponse{}
	switch request.Method {
	case "Action":
		input := &actionandtrigger.Input{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			capResp.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
			break
		}

		if cap.Action == nil {
			capResp.Response = &pb.CapabilityResponse_Error{Error: "no stub provided for Action"}
			break
		}
		resp, err := cap.Action(ctx, input)
		if err != nil {
			capResp.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
		} else {
			payload, err := anypb.New(resp)
			if err == nil {
				capResp.Response = &pb.CapabilityResponse_Payload{Payload: payload}
			} else {
				capResp.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
			}
		}
	default:
		capResp.Response = &pb.CapabilityResponse_Error{Error: fmt.Sprintf("method %s not found", request.Method)}
	}
	return capResp
}

func (cap *BasicCapability) InvokeTrigger(ctx context.Context, request *pb.TriggerSubscriptionRequest) (*pb.Trigger, error) {
	trigger := &pb.Trigger{}
	switch request.Method {
	case "Trigger":
		input := &actionandtrigger.Config{}
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

func (cap *BasicCapability) ID() string {
	return "basic-test-action-trigger@1.0.0"
}
