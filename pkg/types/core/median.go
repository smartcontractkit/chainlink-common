package core

import (
	"context"

	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

type PluginMedian interface {
	services.Service
	// NewMedianFactory returns a new ReportingPluginFactory. If provider implements GRPCClientConn, it can be forwarded efficiently via proxy.
	NewMedianFactory(ctx context.Context, provider types.MedianProvider, contractID string, dataSource, juelsPerFeeCoin, gasPriceSubunits median.DataSource, errorLog ErrorLog) (types.ReportingPluginFactory, error)
}
