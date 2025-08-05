package types

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

// SecureMintProvider provides components needed for a SecureMint OCR3 plugin
type SecureMintProvider interface {
	PluginProvider

	// ExternalAdapter provides mintable amounts and latest blocks per chain
	ExternalAdapter() ExternalAdapter

	// SecureMintContractReader reads latest transmitted report details from contracts
	SecureMintContractReader() SecureMintContractReader

	// ReportMarshaler serializes reports for transmission
	ReportMarshaler() ReportMarshaler
}

// ExternalAdapter interface for PoR calculations
type ExternalAdapter interface {
	// GetPayload returns mintable amounts and latest blocks for queried blocks
	GetPayload(ctx context.Context, blocks map[uint64]uint64) (ExternalAdapterPayload, error)
}

// ExternalAdapterPayload contains mintable amounts, reserve info, and latest blocks
type ExternalAdapterPayload struct {
	Mintables   map[uint64]BlockMintablePair // ChainSelector -> BlockMintablePair
	ReserveInfo ReserveInfo
	LatestBlocks map[uint64]uint64 // ChainSelector -> BlockNumber
}

// BlockMintablePair contains block number and mintable amount
type BlockMintablePair struct {
	Block    uint64
	Mintable *big.Int
}

// ReserveInfo contains reserve amount and timestamp
type ReserveInfo struct {
	ReserveAmount *big.Int
	Timestamp     time.Time
}

// SecureMintContractReader interface for reading Secure Mint contract state
type SecureMintContractReader interface {
	// GetLatestTransmittedReportDetails retrieves latest transmission details
	GetLatestTransmittedReportDetails(ctx context.Context, chain uint64) (TransmittedReportDetails, error)
}

// TransmittedReportDetails contains transmission information
type TransmittedReportDetails struct {
	ConfigDigest    types.ConfigDigest
	SeqNr           uint64
	LatestTimestamp time.Time
}

// ReportMarshaler interface for report serialization
type ReportMarshaler interface {
	// Serialize serializes a report for a specific chain
	Serialize(ctx context.Context, chain uint64, report PorReport) ([]byte, error)

	// MaxReportSize returns maximum serialized report size
	MaxReportSize(ctx context.Context) int
}

// PorReport represents a Secure Mint report
type PorReport struct {
	ConfigDigest types.ConfigDigest
	SeqNr        uint64
	Block        uint64
	Mintable     *big.Int
}

// PluginSecureMint interface for the LOOPP plugin
type PluginSecureMint interface {
	Service
	NewSecureMintFactory(ctx context.Context, provider SecureMintProvider, config SecureMintConfig) (SecureMintPluginFactory, error)
}

// SecureMintPluginFactory interface
type SecureMintPluginFactory interface {
	Service
	ocr3types.ReportingPluginFactory[uint64]
}

// SecureMintConfig holds configuration for the SecureMint plugin
// Configuration comes from offchainConfig, not job specification
type SecureMintConfig struct {
	MaxChains uint32 `json:"maxChains"` // Maximum number of chains to track
}

// Validate validates the SecureMintConfig
func (c SecureMintConfig) Validate() error {
	if c.MaxChains == 0 {
		return fmt.Errorf("maxChains must be greater than 0")
	}
	return nil
} 