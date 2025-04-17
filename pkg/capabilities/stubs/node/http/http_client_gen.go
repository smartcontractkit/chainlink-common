// Code generated by github.com/smartcontractkit/chainlink-common/pkg/capabilities/protoc, DO NOT EDIT.

package http

import (
	"errors"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2/pb"
)

type Client struct {
	// TODO config types (optional)
	// TODO capability interfaces.
}

func (c *Client) Fetch(runtime sdk.NodeRuntime, input *HttpFetchRequest) sdk.Promise[*HttpFetchResponse] {
	wrapped, err := anypb.New(input)
	if err != nil {
		return sdk.PromiseFromResult[*HttpFetchResponse](nil, err)
	}
	return sdk.Then(runtime.CallCapability(&pb.CapabilityRequest{
		Id:      "http@1.0.0",
		Payload: wrapped,
		Method:  "Fetch",
	}), func(i *pb.CapabilityResponse) (*HttpFetchResponse, error) {
		switch payload := i.Response.(type) {
		case *pb.CapabilityResponse_Error:
			return nil, errors.New(payload.Error)
		case *pb.CapabilityResponse_Payload:
			output := &HttpFetchResponse{}
			err = payload.Payload.UnmarshalTo(output)
			return output, err
		default:
			return nil, errors.New("unexpected response type")
		}
	})
}
