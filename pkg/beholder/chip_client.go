package beholder

import (
	"context"
	"fmt"
	"io"

	"github.com/smartcontractkit/chainlink-common/pkg/chipingress"
	"github.com/smartcontractkit/chainlink-common/pkg/chipingress/pb"
)

type ChipIngressClient interface {
	RegisterSchema(ctx context.Context, schemas ...*pb.Schema) (map[string]int, error)
	io.Closer
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
func (sr *chipIngressClient) Close() error { return nil }

// RegisterSchema registers one or more schemas with the Chip Ingress service. Returns a map of subject to version for each registered schema.
func (sr *chipIngressClient) RegisterSchema(ctx context.Context, schemas ...*pb.Schema) (map[string]int, error) {
	request := &pb.RegisterSchemaRequest{
		Schemas: schemas,
	}

	resp, err := sr.client.RegisterSchema(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to register schema: %w", err)
	}

	registeredMap := make(map[string]int)
	for _, schema := range resp.Registered {
		registeredMap[schema.Subject] = int(schema.Version)
	}

	return registeredMap, err
}
