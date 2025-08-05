package main

import (
	"context"
	"fmt"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	libocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

// RelayerExternalAdapter implements the external plugin's ExternalAdapter interface using Relayer
type RelayerExternalAdapter struct {
	provider types.SecureMintProvider
	logger   logger.Logger
}

func NewRelayerExternalAdapter(provider types.SecureMintProvider, logger logger.Logger) *RelayerExternalAdapter {
	return &RelayerExternalAdapter{
		provider: provider,
		logger:   logger,
	}
}

// GetPayload returns mintable amounts and latest blocks for queried blocks
// This is a placeholder implementation that would need to be adapted to the external plugin's interface
func (r *RelayerExternalAdapter) GetPayload(ctx context.Context, blocks map[uint64]uint64) (types.ExternalAdapterPayload, error) {
	// TODO: Implement using provider.ExternalAdapter() and Relayer contract reading
	// This would typically:
	// 1. Use the provider's ExternalAdapter to get mintable amounts
	// 2. Use the Relayer to read contract state for each chain
	// 3. Return the payload in the format expected by the external plugin

	r.logger.Debugw("RelayerExternalAdapter.GetPayload called", "blocks", blocks)

	// Placeholder implementation
	return types.ExternalAdapterPayload{
		Mintables:    make(map[uint64]types.BlockMintablePair),
		ReserveInfo:  types.ReserveInfo{},
		LatestBlocks: blocks,
	}, nil
}

// RelayerContractReader implements the external plugin's ContractReader interface using Relayer
type RelayerContractReader struct {
	provider types.SecureMintProvider
	logger   logger.Logger
}

func NewRelayerContractReader(provider types.SecureMintProvider, logger logger.Logger) *RelayerContractReader {
	return &RelayerContractReader{
		provider: provider,
		logger:   logger,
	}
}

// GetLatestTransmittedReportDetails retrieves latest transmission details
// This is a placeholder implementation that would need to be adapted to the external plugin's interface
func (r *RelayerContractReader) GetLatestTransmittedReportDetails(ctx context.Context, chain uint64) (types.TransmittedReportDetails, error) {
	// TODO: Implement using Relayer to read contract state
	// This would typically:
	// 1. Use the provider's ContractReader to get latest transmission details
	// 2. Use the Relayer to read the specific contract state
	// 3. Return the details in the format expected by the external plugin

	r.logger.Debugw("RelayerContractReader.GetLatestTransmittedReportDetails called", "chain", chain)

	// Placeholder implementation
	return types.TransmittedReportDetails{
		ConfigDigest:    libocrtypes.ConfigDigest{},
		SeqNr:           0,
		LatestTimestamp: time.Now(),
	}, nil
}

// ChainlinkReportMarshaler implements the external plugin's ReportMarshaler interface
type ChainlinkReportMarshaler struct {
	logger logger.Logger
}

func NewChainlinkReportMarshaler(logger logger.Logger) *ChainlinkReportMarshaler {
	return &ChainlinkReportMarshaler{
		logger: logger,
	}
}

// Serialize serializes a report for a specific chain
func (c *ChainlinkReportMarshaler) Serialize(ctx context.Context, chain uint64, report types.PorReport) ([]byte, error) {
	// TODO: Implement report serialization using chainlink-common utilities
	// This would typically:
	// 1. Use chainlink-common serialization utilities
	// 2. Format the report according to the expected protocol
	// 3. Return the serialized bytes

	c.logger.Debugw("ChainlinkReportMarshaler.Serialize called", "chain", chain, "report", report)

	// Placeholder implementation
	return []byte{}, nil
}

// MaxReportSize returns maximum serialized report size
func (c *ChainlinkReportMarshaler) MaxReportSize(ctx context.Context) int {
	// TODO: Return appropriate maximum report size based on protocol requirements
	return 1024
}

// SecureMintFactory wraps the external plugin factory in our LOOPP interface
type SecureMintFactory struct {
	// TODO: Add external plugin factory field when external plugin is imported
	// porFactory *por.PorReportingPluginFactory
	config types.SecureMintConfig
	logger logger.Logger
}

func NewSecureMintFactory(config types.SecureMintConfig, logger logger.Logger) *SecureMintFactory {
	return &SecureMintFactory{
		config: config,
		logger: logger,
	}
}

// NewSecureMintFactory creates a new reporting plugin factory
func (f *SecureMintFactory) NewSecureMintFactory(ctx context.Context, provider types.SecureMintProvider, config types.SecureMintConfig) (types.ReportingPluginFactory, error) {
	// TODO: Implement when external plugin is integrated
	// This would typically:
	// 1. Create the external plugin factory using the imported por package
	// 2. Configure it with the provided provider and config
	// 3. Return a wrapper that implements types.ReportingPluginFactory

	f.logger.Debugw("SecureMintFactory.NewSecureMintFactory called", "config", config)

	return nil, fmt.Errorf("not implemented yet")
}

// Service interface methods
func (f *SecureMintFactory) Start(ctx context.Context) error {
	return nil
}

func (f *SecureMintFactory) Close() error {
	return nil
}

func (f *SecureMintFactory) Ready() error {
	return nil
}

func (f *SecureMintFactory) HealthReport() map[string]error {
	return nil
}

func (f *SecureMintFactory) Name() string {
	return "SecureMintFactory"
}
