package test

import (
	ocr3test "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ocr3/test"
	testtypes "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/test/types"
)

var (
	// ContractTransmitter is a static implementation of the ContractTransmitterTester interface for testing
	// We reuse the OCR3 contract transmitter since it's the same interface
	ContractTransmitter testtypes.OCR3ContractTransmitterEvaluator = ocr3test.ContractTransmitter
)
