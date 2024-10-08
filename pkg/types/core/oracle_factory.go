package core

import (
	"context"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

type Oracle interface {
	Start(ctx context.Context) error
	Close(ctx context.Context) error
}

type OracleFactory interface {
	NewOracle(ctx context.Context, args OracleArgs) (Oracle, error)
}

type OracleArgs struct {
	LocalConfig                   types.LocalConfig
	ReportingPluginFactoryService ocr3types.ReportingPluginFactory[[]byte]
	ContractTransmitter           ocr3types.ContractTransmitter[[]byte]
}
