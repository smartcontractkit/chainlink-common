package main

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/por_mock_ocr3plugin/por"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
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
// This bridges between our types and the external plugin's types
func (r *RelayerExternalAdapter) GetPayload(ctx context.Context, blocks por.Blocks) (por.ExternalAdapterPayload, error) {
	// TODO(gg): Implement using provider.ExternalAdapter() and Relayer contract reading
	// This would typically:
	// 1. Use the provider's ExternalAdapter to get mintable amounts
	// 2. Use the Relayer to read contract state for each chain
	// 3. Convert between our types and external plugin types
	
	r.logger.Debugw("RelayerExternalAdapter.GetPayload called", "blocks", blocks)
	
	// Convert blocks to our format
	ourBlocks := make(map[uint64]uint64)
	for chain, block := range blocks {
		ourBlocks[uint64(chain)] = uint64(block)
	}
	
	// Get payload from our provider
	ourPayload, err := r.provider.ExternalAdapter().GetPayload(ctx, ourBlocks)
	if err != nil {
		return por.ExternalAdapterPayload{}, fmt.Errorf("failed to get payload from provider: %w", err)
	}
	
	// Convert to external plugin format
	externalMintables := make(por.Mintables)
	for chain, pair := range ourPayload.Mintables {
		externalMintables[por.ChainSelector(chain)] = por.BlockMintablePair{
			Block:    por.BlockNumber(pair.Block),
			Mintable: pair.Mintable,
		}
	}
	
	externalLatestBlocks := make(por.Blocks)
	for chain, block := range ourPayload.LatestBlocks {
		externalLatestBlocks[por.ChainSelector(chain)] = por.BlockNumber(block)
	}
	
	return por.ExternalAdapterPayload{
		Mintables:   externalMintables,
		ReserveInfo: por.ReserveInfo{
			ReserveAmount: ourPayload.ReserveInfo.ReserveAmount,
			Timestamp:     ourPayload.ReserveInfo.Timestamp,
		},
		LatestBlocks: externalLatestBlocks,
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
// This bridges between our types and the external plugin's types
func (r *RelayerContractReader) GetLatestTransmittedReportDetails(ctx context.Context, chain por.ChainSelector) (por.TransmittedReportDetails, error) {
	// TODO(gg): Implement using Relayer to read contract state
	// This would typically:
	// 1. Use the provider's ContractReader to get latest transmission details
	// 2. Use the Relayer to read the specific contract state
	// 3. Convert between our types and external plugin types
	
	r.logger.Debugw("RelayerContractReader.GetLatestTransmittedReportDetails called", "chain", chain)
	
	// Get details from our provider
	ourDetails, err := r.provider.SecureMintContractReader().GetLatestTransmittedReportDetails(ctx, uint64(chain))
	if err != nil {
		return por.TransmittedReportDetails{}, fmt.Errorf("failed to get details from provider: %w", err)
	}
	
	// Convert to external plugin format
	return por.TransmittedReportDetails{
		ConfigDigest:    ourDetails.ConfigDigest,
		SeqNr:           ourDetails.SeqNr,
		LatestTimestamp: ourDetails.LatestTimestamp,
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
func (c *ChainlinkReportMarshaler) Serialize(ctx context.Context, chain por.ChainSelector, report por.PorReport) ([]byte, error) {
	// TODO(gg): Implement report serialization using chainlink-common utilities
	// This would typically:
	// 1. Use chainlink-common serialization utilities
	// 2. Format the report according to the expected protocol
	// 3. Return the serialized bytes
	
	c.logger.Debugw("ChainlinkReportMarshaler.Serialize called", "chain", chain, "report", report)
	
	// Convert to our format and use our marshaler
	_ = types.PorReport{
		ConfigDigest: report.ConfigDigest,
		SeqNr:        report.SeqNr,
		Block:        uint64(report.Block),
		Mintable:     report.Mintable,
	}
	
	// Use our provider's marshaler
	// TODO(gg): Get marshaler from provider when available
	return []byte{}, nil
}

// MaxReportSize returns maximum serialized report size
func (c *ChainlinkReportMarshaler) MaxReportSize(ctx context.Context) int {
	// Based on typical OCR report sizes and SecureMint requirements
	// OCR reports are typically 1-2KB, and SecureMint reports include:
	// - ConfigDigest (32 bytes)
	// - SeqNr (8 bytes)
	// - Block number (8 bytes)
	// - Mintable amount (32 bytes)
	// - Additional metadata and padding
	return 2048 // 2KB should be sufficient for SecureMint reports
}

// SecureMintFactory wraps the external plugin factory in our LOOPP interface
type SecureMintFactory struct {
	porFactory *por.PorReportingPluginFactory
	config     types.SecureMintConfig
	logger     logger.Logger
}

func NewSecureMintFactory(config types.SecureMintConfig, logger logger.Logger, porFactory *por.PorReportingPluginFactory) *SecureMintFactory {
	return &SecureMintFactory{
		porFactory: porFactory,
		config:     config,
		logger:     logger,
	}
}

// NewSecureMintFactory creates a new reporting plugin factory
func (f *SecureMintFactory) NewSecureMintFactory(ctx context.Context, provider types.SecureMintProvider, config types.SecureMintConfig) (types.ReportingPluginFactory, error) {
	f.logger.Debugw("SecureMintFactory.NewSecureMintFactory called", "config", config)
	
	// Convert our config to external plugin config
	externalConfig := por.PorOffchainConfig{
		MaxChains: config.MaxChains,
	}
	
	// Serialize the external config
	offchainConfigBytes, err := externalConfig.Serialize()
	if err != nil {
		return nil, fmt.Errorf("failed to serialize external config: %w", err)
	}
	
	// Create OCR3 config for the external plugin
	ocr3Config := ocr3types.ReportingPluginConfig{
		ConfigDigest:    libocrtypes.ConfigDigest{}, // TODO(gg): Get from context
		OracleID:        0,                          // TODO(gg): Get from context
		N:               4,                          // TODO(gg): Get from context
		F:               1,                          // TODO(gg): Get from context
		OnchainConfig:   []byte{},                   // TODO(gg): Get from context
		OffchainConfig:  offchainConfigBytes,
		EstimatedRoundInterval: 0,                   // TODO(gg): Get from context
		MaxDurationQuery:       0,                   // TODO(gg): Get from context
		MaxDurationObservation: 0,                   // TODO(gg): Get from context
	}
	
	// Create the external plugin
	externalPlugin, _, err := f.porFactory.NewReportingPlugin(ctx, ocr3Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create external plugin: %w", err)
	}
	
	// Wrap the external plugin in our LOOPP interface
	wrapper := &SecureMintReportingPluginWrapper{
		plugin: externalPlugin,
		logger: f.logger,
	}
	
	return wrapper, nil
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

// SecureMintReportingPluginWrapper wraps the external OCR3 plugin to implement our OCR2 interface
type SecureMintReportingPluginWrapper struct {
	plugin ocr3types.ReportingPlugin[por.ChainSelector]
	logger logger.Logger
}

// NewReportingPlugin creates a new reporting plugin instance
func (w *SecureMintReportingPluginWrapper) NewReportingPlugin(ctx context.Context, config libocrtypes.ReportingPluginConfig) (libocrtypes.ReportingPlugin, libocrtypes.ReportingPluginInfo, error) {
	// TODO(gg): Convert OCR2 config to OCR3 config and create plugin
	// For now, return a placeholder
	w.logger.Debugw("SecureMintReportingPluginWrapper.NewReportingPlugin called", "config", config)
	
	return nil, libocrtypes.ReportingPluginInfo{}, fmt.Errorf("not implemented yet")
}

// Service interface methods
func (w *SecureMintReportingPluginWrapper) Start(ctx context.Context) error {
	return nil
}

func (w *SecureMintReportingPluginWrapper) Close() error {
	return w.plugin.Close()
}

func (w *SecureMintReportingPluginWrapper) Ready() error {
	return nil
}

func (w *SecureMintReportingPluginWrapper) HealthReport() map[string]error {
	return nil
}

func (w *SecureMintReportingPluginWrapper) Name() string {
	return "SecureMintReportingPluginWrapper"
}
