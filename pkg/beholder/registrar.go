package beholder

import (
	"context"
	"fmt"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

type Registrar interface {
	Register(ctx context.Context, schemas ...*pb.Schema) error
}

type schemaRegistry struct {
	client chipingress.Client
}

func NewRegistrar(client chipingress.Client) (Registrar, error) {
	if client == nil {
		return nil, fmt.Errorf("chip ingress client is nil")
	}

	return &schemaRegistry{
		client: client,
	}, nil
}

// Register registers one or more schemas with the Chip Ingress service
func (sr *schemaRegistry) Register(ctx context.Context, schemas ...*pb.Schema) error {
	request := &pb.RegisterSchemaRequest{
		Schemas: schemas,
	}

	_, err := sr.client.RegisterSchema(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to register schema: %w", err)
	}

	return nil
}
