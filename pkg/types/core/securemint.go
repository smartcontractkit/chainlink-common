package core

import (
	"context"

	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type PluginSecureMint interface {
	services.Service
	// NewSecureMintFactory returns a new ReportingPluginFactory. If provider implements GRPCClientConn, it can be forwarded efficiently via proxy.
	// Implemented in pkg/loop/internal/reportingplugin/securemint
	// TODO(gg): update signature, no need for all params probably
	NewSecureMintFactory(ctx context.Context, provider types.SecureMintProvider, contractID string, dataSource, juelsPerFeeCoin, gasPriceSubunits median.DataSource, errorLog ErrorLog, deviationFuncDefinition map[string]any) (types.ReportingPluginFactory, error)
}
