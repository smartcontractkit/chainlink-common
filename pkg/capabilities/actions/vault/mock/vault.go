package mock

import (
	"context"
	"errors"

	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/actions/vault"
)

type Vault struct {
	Fn func(ctx context.Context, req *vault.GetSecretsRequest) (*vault.GetSecretsResponse, error)
}

func (m Vault) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	return capabilities.CapabilityInfo{
		ID:             vault.CapabilityID,
		CapabilityType: capabilities.CapabilityTypeAction,
		IsLocal:        true,
	}, nil
}

func (m Vault) Execute(ctx context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	vr := &vault.GetSecretsRequest{}
	err := req.Payload.UnmarshalTo(vr)
	if err != nil {
		return capabilities.CapabilityResponse{}, errors.New("received unexpected payload: want *vault.GetSecretsRequest")
	}
	if req.Method != "vault.secrets.get" {
		return capabilities.CapabilityResponse{}, errors.New("received unexpected method: want vault.MethodGetSecrets")
	}

	vresp, err := m.Fn(ctx, vr)
	if err != nil {
		return capabilities.CapabilityResponse{}, err
	}

	anyvresp, err := anypb.New(vresp)
	if err != nil {
		return capabilities.CapabilityResponse{}, err
	}

	return capabilities.CapabilityResponse{
		Payload: anyvresp,
	}, err
}

func (m Vault) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	return errors.New("not used")
}

func (m Vault) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	return errors.New("not used")
}
