package core

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

type ValidationService interface {
	services.Service
	ValidateConfig(ctx context.Context, config map[string]any) error
}

type ValidationServiceClient interface {
	ValidateConfig(ctx context.Context, config map[string]any) error
}
type ValidationServiceServer interface {
	ValidateConfig(ctx context.Context, config map[string]any) error
}
