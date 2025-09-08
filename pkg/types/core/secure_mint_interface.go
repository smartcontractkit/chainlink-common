package core

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core/securemint"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
)

const PluginSecureMintName = "securemint"

// PluginSecureMint is the interface for the secure mint plugin.
type PluginSecureMint interface {
	services.Service
	// NewSecureMintFactory returns a ReportingPluginFactory for the secure mint plugin.
	NewSecureMintFactory(ctx context.Context, lggr logger.Logger, externalAdapter securemint.ExternalAdapter) (ReportingPluginFactory[securemint.ChainSelector], error)
}

// ReportingPluginFactory wraps ocr3types.ReportingPluginFactory[RI] to add a Service to it.
type ReportingPluginFactory[RI any] interface {
	services.Service
	ocr3types.ReportingPluginFactory[RI]
}
