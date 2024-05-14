package model

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type ExecutionStrategy interface {
	Apply(ctx context.Context, l logger.Logger, cap capabilities.CallbackCapability, req capabilities.CapabilityRequest) (values.Value, error)
}
