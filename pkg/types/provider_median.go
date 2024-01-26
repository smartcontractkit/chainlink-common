package types

import (
	"context"

	"github.com/smartcontractkit/libocr/offchainreporting2/reportingplugin/median"
	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

// MedianProvider provides all components needed for a median OCR2 plugin.
type MedianProvider interface {
	PluginProvider
	ReportCodec() median.ReportCodec
	MedianContract() median.MedianContract
	OnchainConfigCodec() median.OnchainConfigCodec
}

type PluginMedian interface {
	services.Service
	// NewMedianFactory returns a new ReportingPluginFactory. If provider implements GRPCClientConn, it can be forwarded efficiently via proxy.
	NewMedianFactory(ctx context.Context, provider MedianProvider, dataSource, juelsPerFeeCoin median.DataSource, errorLog ErrorLog) (ReportingPluginFactory, error)
}

type ReportingPluginFactory interface {
	services.Service
	libocr.ReportingPluginFactory
}
