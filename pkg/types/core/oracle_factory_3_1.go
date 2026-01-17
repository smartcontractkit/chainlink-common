package core

import (
	"context"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3_1types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

type OracleFactory3_1 interface {
	NewOracle(ctx context.Context, args OracleArgs) (Oracle, error)
}

type OracleArgs3_1 struct {
	LocalConfig                   types.LocalConfig
	ReportingPluginFactoryService ocr3_1types.ReportingPluginFactory[[]byte]
	ContractTransmitter           ocr3types.ContractTransmitter[[]byte]
}
