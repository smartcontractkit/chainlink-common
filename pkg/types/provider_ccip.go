package types

import (
	"context"

	"github.com/google/uuid"
	"github.com/smartcontractkit/libocr/offchainreporting2plus/ocr3types"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccip"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
)

type CCIPCommitProvider interface {
	PluginProvider

	NewCommitStoreReader(ctx context.Context, addr ccip.Address) (ccip.CommitStoreReader, error)
	NewOffRampReader(ctx context.Context, addr ccip.Address) (ccip.OffRampReader, error)
	NewOnRampReader(ctx context.Context, addr ccip.Address, sourceSelector uint64, destSelector uint64) (ccip.OnRampReader, error)
	NewPriceGetter(ctx context.Context) (ccip.PriceGetter, error)
	NewPriceRegistryReader(ctx context.Context, addr ccip.Address) (ccip.PriceRegistryReader, error)
	SourceNativeToken(ctx context.Context, addr ccip.Address) (ccip.Address, error)
}

type CCIPExecProvider interface {
	PluginProvider

	GetTransactionStatus(ctx context.Context, transactionID string) (TransactionStatus, error)
	NewCommitStoreReader(ctx context.Context, addr ccip.Address) (ccip.CommitStoreReader, error)
	NewOffRampReader(ctx context.Context, addr ccip.Address) (ccip.OffRampReader, error)
	NewOnRampReader(ctx context.Context, addr ccip.Address, sourceSelector uint64, destSelector uint64) (ccip.OnRampReader, error)
	NewPriceRegistryReader(ctx context.Context, addr ccip.Address) (ccip.PriceRegistryReader, error)
	NewTokenDataReader(ctx context.Context, tokenAddress ccip.Address) (ccip.TokenDataReader, error)
	NewTokenPoolBatchedReader(ctx context.Context, offRampAddress ccip.Address, sourceSelector uint64) (ccip.TokenPoolBatchedReader, error)
	SourceNativeToken(ctx context.Context, addr ccip.Address) (ccip.Address, error)
}

type CCIPCommitFactoryGenerator interface {
	services.Service
	NewCommitFactory(ctx context.Context, provider CCIPCommitProvider) (ReportingPluginFactory, error)
}

type CCIPExecutionFactoryGenerator interface {
	services.Service
	NewExecutionFactory(ctx context.Context, srcProvider CCIPExecProvider, dstProvider CCIPExecProvider, srcChainID int64, dstChainID int64, sourceTokenAddress string) (ReportingPluginFactory, error)
}
type CCIPFactoryGenerator interface {
	CCIPCommitFactoryGenerator
	CCIPExecutionFactoryGenerator
}

// CCIPProvider is a product-specific interface that exposes the necessary components
// for running CCIP on a specific chain/chain family. It includes access to the ChainAccessor,
type CCIPProvider interface {
	services.Service
	ChainAccessor() ccipocr3.ChainAccessor
	ContractTransmitter() ocr3types.ContractTransmitter[[]byte]
	Codec() ccipocr3.Codec
}

// CCIPProviderArgs are the args required to create a CCIP Provider through a Relayer.
// The are common to all relayer implementations.
type CCIPProviderArgs struct {
	ExternalJobID        uuid.UUID
	ContractReaderConfig []byte
	ChainWriterConfig    []byte
	OffRampAddress       string
	PluginType           uint32
}

// ExtraDataCodecRegistryService maintains a registry of SourceChainExtraDataCodec instances by chain family.
// It implements the Service interface and manages the lifecycle of codec registrations.
type ExtraDataCodecRegistryService interface {
	services.Service
	// RegisterChainFamily pre-registers a chain family that will be initialized later.
	RegisterChainFamily(chainFamily string)
	// SetSourceChainCodec registers a SourceChainExtraDataCodec for the given chain family
	// and marks it as initialized.
	SetSourceChainCodec(chainFamily string, codec ccipocr3.SourceChainExtraDataCodec)
	// GetExtraDataCodec returns the complete ExtraDataCodec map.
	// This includes all registered chain families, with no-op codecs for uninitialized ones.
	GetExtraDataCodec() (ccipocr3.ExtraDataCodec, error)
}
