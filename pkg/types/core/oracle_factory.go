package core

import (
	"context"

	"github.com/smartcontractkit/libocr/offchainreporting2plus"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

type OracleFactory interface {
	NewOracle(ctx context.Context, args OracleArgs) (offchainreporting2plus.Oracle, error)
}

type OracleArgs struct {
	LocalConfig                   types.LocalConfig
	ReportingPluginFactoryService ocr3types.ReportingPluginFactory[[]byte]
	ContractConfigTracker         types.ContractConfigTracker
	ContractTransmitter           ocr3types.ContractTransmitter[[]byte]
	OffchainConfigDigester        types.OffchainConfigDigester
}
