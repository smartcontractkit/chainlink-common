package internal

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/keystore"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

type PluginRelayer interface {
	services.Service
	NewRelayer(ctx context.Context, config string, keystore, csaKeystore core.Keystore, capabilityRegistry core.CapabilitiesRegistry) (Relayer, error)
}

type MedianProvider interface {
	NewMedianProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.MedianProvider, error)
}

type MercuryProvider interface {
	NewMercuryProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.MercuryProvider, error)
}

type FunctionsProvider interface {
	NewFunctionsProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.FunctionsProvider, error)
}

type AutomationProvider interface {
	NewAutomationProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.AutomationProvider, error)
}

type CCIPExecProvider interface {
	NewExecutionProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.CCIPExecProvider, error)
}

type CCIPCommitProvider interface {
	NewCommitProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.CCIPCommitProvider, error)
}

type OCR3CapabilityProvider interface {
	NewOCR3CapabilityProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.OCR3CapabilityProvider, error)
}

// Relayer is like types.Relayer, but with a dynamic NewPluginProvider method.
type Relayer interface {
	types.ChainService

	EVM() (types.EVMService, error)
	TON() (types.TONService, error)
	Solana() (types.SolanaService, error)
	Aptos() (types.AptosService, error)
	// NewContractWriter returns a new ContractWriter.
	// The format of config depends on the implementation.
	NewContractWriter(ctx context.Context, contractWriterConfig []byte) (types.ContractWriter, error)

	// NewContractReader returns a new ContractReader.
	// The format of contractReaderConfig depends on the implementation.
	NewContractReader(ctx context.Context, contractReaderConfig []byte) (types.ContractReader, error)
	NewConfigProvider(context.Context, types.RelayArgs) (types.ConfigProvider, error)
	NewPluginProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.PluginProvider, error)
	NewLLOProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.LLOProvider, error)
	NewCCIPProvider(context.Context, types.CCIPProviderArgs) (types.CCIPProvider, error)
}

// Keystore This interface contains all the keystore GRPC functionality, keystore.Keystore is meant to be exposed to consumers and the keystore.Management interface in exposed only to the core node
type Keystore interface {
	services.Service
	keystore.GRPCService
}
