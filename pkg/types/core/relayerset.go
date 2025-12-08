package core

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
)

type RelayerSet interface {
	Get(ctx context.Context, relayID types.RelayID) (Relayer, error)

	// List lists the relayers corresponding to `...types.RelayID`
	// returning all relayers if len(...types.RelayID) == 0.
	List(ctx context.Context, relayIDs ...types.RelayID) (map[types.RelayID]Relayer, error)
}

type PluginArgs struct {
	TransmitterID       string
	PluginConfig        []byte
	CapRegConfigTracker ocrtypes.ContractConfigTracker
}

type RelayArgs struct {
	ContractID         string
	RelayConfig        []byte
	ProviderType       string
	MercuryCredentials *types.MercuryCredentials
}

type Relayer interface {
	services.Service
	// EVM returns EVMService that provides access to evm-family specific functionalities
	EVM() (types.EVMService, error)
	// TON returns TONService that provides access to TON specific functionalities
	TON() (types.TONService, error)
	// Solana returns SolanaService that provides access to Solana specific functionalities
	Solana() (types.SolanaService, error)
	NewPluginProvider(context.Context, RelayArgs, PluginArgs) (PluginProvider, error)
	NewContractReader(_ context.Context, contractReaderConfig []byte) (types.ContractReader, error)
	NewContractWriter(_ context.Context, contractWriterConfig []byte) (types.ContractWriter, error)
	LatestHead(context.Context) (types.Head, error)
	GetChainInfo(ctx context.Context) (types.ChainInfo, error)
}

// PluginProvider provides config required by the oracle factory.
type PluginProvider interface {
	types.ConfigProvider
}
