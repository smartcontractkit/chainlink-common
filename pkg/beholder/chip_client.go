package beholder

import (
	"context"
	"fmt"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

type ChipIngressClient interface {
	RegisterSchema(ctx context.Context, schemas ...*pb.Schema) error
}

type chipIngressClient struct {
	client chipingress.Client
}

func NewChipIngressClient(client chipingress.Client) (ChipIngressClient, error) {
	if client == nil {
		return nil, fmt.Errorf("chip ingress client is nil")
	}

	return &chipIngressClient{
		client: client,
	}, nil
}

// RegisterSchema registers one or more schemas with the Chip Ingress service
func (sr *chipIngressClient) RegisterSchema(ctx context.Context, schemas ...*pb.Schema) error {
	request := &pb.RegisterSchemaRequest{
		Schemas: schemas,
	}

	_, err := sr.client.RegisterSchema(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to register schema: %w", err)
	}

	return nil
}
