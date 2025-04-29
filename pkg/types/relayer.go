package types

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
)

type RelayID struct {
	Network string
	ChainID string
}

// ID uniquely identifies a relayer by network and chain id
func (i *RelayID) Name() string {
	return fmt.Sprintf("%s.%s", i.Network, i.ChainID)
}

func (i *RelayID) String() string {
	return i.Name()
}
func NewRelayID(n string, c string) RelayID {
	return RelayID{Network: n, ChainID: c}
}

func (i *RelayID) UnmarshalString(s string) error {
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return fmt.Errorf("error unmarshaling Identifier. %s does not match expected pattern", s)
	}

	i.Network = parts[0]
	i.ChainID = parts[1]
	return nil
}

// PluginArgs are the args required to create any OCR2 plugin components.
// It's possible that the plugin config might actually be different
// per relay type, so we pass the config directly through.
type PluginArgs struct {
	TransmitterID string
	PluginConfig  []byte
}

// RelayArgs are the args required to create relayer.
// The are common to all xelayer implementations.
type RelayArgs struct {
	ExternalJobID      uuid.UUID
	JobID              int32
	OracleSpecID       int32
	ContractID         string
	New                bool   // Whether this is a first time job add.
	RelayConfig        []byte // The specific configuration of a given relayer instance. Will vary by relayer type.
	ProviderType       string
	MercuryCredentials *MercuryCredentials
}

type MercuryCredentials struct {
	LegacyURL string
	URL       string
	Username  string
	Password  string
}

type ChainStatus struct {
	ID      string
	Enabled bool
	Config  string // TOML
}

type NodeStatus struct {
	ChainID string
	Name    string
	Config  string // TOML
	State   string
}

type TransactionFee struct {
	TransactionFee *big.Int
}

// ChainService is a sub-interface that encapsulates the explicit interactions with a chain, rather than through a provider.
type ChainService interface {
	Service

	// LatestHead returns the latest head for the underlying chain.
	LatestHead(ctx context.Context) (Head, error)
	// GetChainStatus returns the ChainStatus for this Relayer.
	GetChainStatus(ctx context.Context) (ChainStatus, error)
	// ListNodeStatuses returns the status of RPC nodes.
	ListNodeStatuses(ctx context.Context, pageSize int32, pageToken string) (stats []NodeStatus, nextPageToken string, total int, err error)
	// Transact submits a transaction to transfer tokens.
	// If balanceCheck is true, the balance will be checked before submitting.
	Transact(ctx context.Context, from, to string, amount *big.Int, balanceCheck bool) error
	// Replay is an emergency recovery tool to re-process blocks starting at the provided fromBlock
	Replay(ctx context.Context, fromBlock string, args map[string]any) error
}

type EVMService interface {
	// -- ChainService
	// Direct Calls
	CallContract(ctx context.Context, msg *evm.CallMsg, confidence primitives.ConfidenceLevel) ([]byte, error)
	GetLogs(ctx context.Context, filterQuery evm.EVMFilterQuery) ([]*evm.Log, error)
	BalanceAt(ctx context.Context, account string, blockNumber *big.Int) (*big.Int, error)
	EstimateGas(ctx context.Context, call *evm.CallMsg) (uint64, error)
	TransactionByHash(ctx context.Context, hash string) (*evm.Transaction, error)
	TransactionReceipt(ctx context.Context, txHash string) (*evm.Receipt, error)

	// ChainService

	// GetTransactionFee retrieves the fee of a transaction in wei from the underlying chain's TXM
	GetTransactionFee(ctx context.Context, transactionID string) (*TransactionFee, error)
	LatestAndFinalizedHead(ctx context.Context) (latest Head, finalized Head, err error)
	QueryLogsFromCache(ctx context.Context, filterQuery []query.Expression,
		limitAndSort query.LimitAndSort, confidenceLevel primitives.ConfidenceLevel) ([]*evm.Log, error)
	SubscribeLogTrigger(ctx context.Context, filterQuery []query.Expression) (chan<- *evm.Log, error)
	RegisterLogTracking(ctx context.Context, filter evm.FilterQuery) error
	UnregisterLogTracking(ctx context.Context, filterName string) error
	GetTransactionStatus(ctx context.Context, transactionID string) (TransactionStatus, error)
}

// Relayer extends ChainService with providers for each product.
type Relayer interface {
	ChainService

	EVM() (EVMService, error)
	// NewContractWriter returns a new ContractWriter.
	// The format of config depends on the implementation.
	NewContractWriter(ctx context.Context, config []byte) (ContractWriter, error)

	// NewContractReader returns a new ContractReader.
	// The format of contractReaderConfig depends on the implementation.
	// See evm.ContractReaderConfig
	NewContractReader(ctx context.Context, contractReaderConfig []byte) (ContractReader, error)

	NewConfigProvider(ctx context.Context, rargs RelayArgs) (ConfigProvider, error)

	NewMedianProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (MedianProvider, error)
	NewMercuryProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (MercuryProvider, error)
	NewFunctionsProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (FunctionsProvider, error)
	NewAutomationProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (AutomationProvider, error)
	NewLLOProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (LLOProvider, error)
	NewCCIPCommitProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (CCIPCommitProvider, error)
	NewCCIPExecProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (CCIPExecProvider, error)

	NewPluginProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (PluginProvider, error)

	NewOCR3CapabilityProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (OCR3CapabilityProvider, error)
}
