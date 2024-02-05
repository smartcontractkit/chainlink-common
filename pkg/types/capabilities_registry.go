package types

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

type CapabilitiesRegistry interface {
	Get(ctx context.Context, ID string) (capabilities.BaseCapability, error)
}
