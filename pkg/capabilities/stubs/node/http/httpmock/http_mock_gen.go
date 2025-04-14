// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc, DO NOT EDIT.

package httpmock

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/stubs/node/http"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/testutils"
)

// avoid unused imports
var _ = capabilities.CapabilityInfo{}
var _ = testutils.Registry{}

type ClientCapability struct {
	// TODO teardown with unrgister if register is needed, or allow setup and teardown
	// TODO register if needed...
	Fetch func(ctx context.Context, input *http.HttpFetchRequest /* TODO config? */) (*http.HttpFetchResponse, error)
}

func (cap *ClientCapability) Invoke(ctx context.Context, request *pb.CapabilityRequest) *pb.CapabilityResponse {
	capResp := &pb.CapabilityResponse{}
	switch request.Method {
	case "Fetch":
		input := &http.HttpFetchRequest{}
		if err := request.Payload.UnmarshalTo(input); err != nil {
			capResp.Response = &pb.CapabilityResponse_Error{Error: err.Error()}
			break
		}

		if cap.Fetch == nil {
			capResp.Response = &pb.CapabilityResponse_Error{Error: "no stub provided for Fetch"}
			break
		}
		resp, err := cap.Fetch(ctx, input)
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

func (cap *ClientCapability) InvokeTrigger(ctx context.Context, request *pb.TriggerSubscriptionRequest) (*pb.Trigger, error) {
	return nil, fmt.Errorf("method %s not found", request.Method)
}

func (cap *ClientCapability) ID() string {
	return "http@1.0.0"
}
