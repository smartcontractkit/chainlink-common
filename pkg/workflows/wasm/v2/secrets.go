package wasm

import (
	"errors"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

type secretsProvider struct {
	runtime sdk.RuntimeBase
}

func (s *secretsProvider) GetSecret(sdk.GetSecretRequest) string {
	promise := s.getSecret(nil)
	resp, err := promise.Await()
	if err != nil {
		panic(err)
	}
	return resp.Value
}

func (s *secretsProvider) getSecret(req *sdkpb.GetSecretRequest) sdk.Promise[*sdkpb.GetSecretResponse] {
	wrapped, err := anypb.New(req)
	if err != nil {
		return sdk.PromiseFromResult[*sdkpb.GetSecretResponse](nil, err)
	}
	return sdk.Then(s.runtime.CallCapability(&sdkpb.CapabilityRequest{
		Id:      "vault@1.0.0",
		Payload: wrapped,
		Method:  "GetSecret",
	}), func(i *sdkpb.CapabilityResponse) (*sdkpb.GetSecretResponse, error) {
		switch payload := i.Response.(type) {
		case *sdkpb.CapabilityResponse_Error:
			return nil, errors.New(payload.Error)
		case *sdkpb.CapabilityResponse_Payload:
			output := &sdkpb.GetSecretResponse{}
			err = payload.Payload.UnmarshalTo(output)
			return output, err
		default:
			return nil, errors.New("unexpected response type")
		}
	})
}
