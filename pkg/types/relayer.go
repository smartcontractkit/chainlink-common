package types

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/solana"
	"github.com/smartcontractkit/chainlink-common/pkg/types/chains/ton"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query"
	"github.com/smartcontractkit/chainlink-common/pkg/types/query/primitives"
	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
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
	TransmitterID       string
	PluginConfig        []byte
	CapRegConfigTracker ocrtypes.ContractConfigTracker
}

// RelayArgs are the args required to create relayer.
// The are common to all relayer implementations.
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

type ChainInfo struct {
	FamilyName  string
	ChainID     string
	NetworkName string
	// NetworkNameFull has network testnet, mainnet or devnet identifier attached.
	NetworkNameFull string
}

type NodeStatus struct {
	ChainID string
	Name    string
	Config  string // TOML
	State   string
}

// ChainService is a sub-interface that encapsulates the explicit interactions with a chain, rather than through a provider.
type ChainService interface {
	Service

	// LatestHead returns the latest head for the underlying chain.
	LatestHead(ctx context.Context) (Head, error)
	// GetChainInfo returns the ChainInfo for this Relayer.
	GetChainInfo(ctx context.Context) (ChainInfo, error)
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

// GethClient is the subset of go-ethereum client methods implemented by EVMService.
type GethClient interface {
	// BalanceAt returns the wei balance of the given account.
	//
	// Parameters:
	// request.BlockNumber - specifies at which block height to fetch the balance:
	//   - nil or -2: latest block
	//   - -3: finalized block
	//   - -4: safe block
	//   - positive value: specific block at that height
	//
	// request.ConfidenceLevel - determines if additional verification is required (only applicable for positive blockNumber values):
	//   - "Unconfirmed" or empty string: no additional verification
	//   - "Finalized": returns error if specified blockNumber is not finalized
	//   - "Safe": returns error if specified blockNumber is not safe
	BalanceAt(ctx context.Context, request evm.BalanceAtRequest) (*evm.BalanceAtReply, error)

	// CallContract executes a message call transaction, which is directly executed in the VM of the node,
	// but never mined into the blockchain.
	//
	// request.BlockNumber - defines block at which call will be executed:
	//   - nil or -2: latest block
	//   - -3: finalized block
	//   - -4: safe block
	//   - positive value: specific block at that height
	//
	// request.ConfidenceLevel - determines if additional verification is required (only applicable for positive blockNumber values):
	//   - "Unconfirmed" or empty string: no additional verification
	//   - "Finalized": returns error if call is executed at block that is not safe
	//   - "Safe": returns error if call is executed at block that is not safe
	CallContract(ctx context.Context, request evm.CallContractRequest) (*evm.CallContractReply, error)

	// FilterLogs executes a filter query.
	//
	// request.ConfidenceLevel - determines if additional verification is required (only applicable if both q.FromBlock and q.ToBlock are positive values):
	//   - "Unconfirmed" or empty string: no additional verification
	//   - "Finalized": returns error if specified q.ToBlockNumber is not finalized
	//   - "Safe": returns error if specified q.ToBlockNumber is not safe
	FilterLogs(ctx context.Context, request evm.FilterLogsRequest) (*evm.FilterLogsReply, error)

	// HeaderByNumber returns a block header from the current canonical chain with the specified block number.
	//
	// Parameters:
	// request.BlockNumber - specifies which block to fetch:
	//   - nil or -2: latest block
	//   - -3: finalized block
	//   - -4: safe block
	//   - positive value: specific block at that height
	//
	// request.ConfidenceLevel - determines if additional verification is required (only applicable for positive blockNumber values):
	//   - "Unconfirmed" or empty string: no additional verification
	//   - "Finalized": returns error if requested is not finalized
	//   - "Safe": returns error if requested block is not safe
	HeaderByNumber(ctx context.Context, request evm.HeaderByNumberRequest) (*evm.HeaderByNumberReply, error)
	EstimateGas(ctx context.Context, call *evm.CallMsg) (uint64, error)
	GetTransactionByHash(ctx context.Context, request evm.GetTransactionByHashRequest) (*evm.Transaction, error)
	GetTransactionReceipt(ctx context.Context, request evm.GeTransactionReceiptRequest) (*evm.Receipt, error)
}

type EVMService interface {
	GethClient
	// RegisterLogTracking registers a persistent log filter for tracking and caching logs
	// based on the provided filter parameters. Once registered, matching logs will be collected
	// over time and stored in a cache for future querying.
	// noop guaranteed when filter.Name exists
	RegisterLogTracking(ctx context.Context, filter evm.LPFilterQuery) error

	// UnregisterLogTracking removes a previously registered log filter by its name.
	// After removal, logs matching this filter will no longer be collected or cached.
	// noop guaranteed when filterName doesn't exist
	UnregisterLogTracking(ctx context.Context, filterName string) error

	// QueryTrackedLogs retrieves logs from the  log storage based on the provided
	// query expression, sorting, and confidence level. It only returns logs that were
	// collected through previously registered log filters.
	QueryTrackedLogs(ctx context.Context, filterQuery []query.Expression,
		limitAndSort query.LimitAndSort, confidenceLevel primitives.ConfidenceLevel) ([]*evm.Log, error)

	// GetLatestLPBlock retrieves current LatestBlock from cache perspective
	GetLatestLPBlock(ctx context.Context) (*evm.LPBlock, error)

	// GetFiltersNames returns all registered filters' names for later pruning
	// TODO PLEX-1465: once code is moved away, remove this GetFiltersNames method
	GetFiltersNames(ctx context.Context) ([]string, error)

	// GetTransactionFee retrieves the fee of a transaction in wei from the underlying chain
	GetTransactionFee(ctx context.Context, transactionID IdempotencyKey) (*evm.TransactionFee, error)

	// Submits a transaction to the EVM chain. It will return once the transaction is included in a block or an error occurs.
	SubmitTransaction(ctx context.Context, txRequest evm.SubmitTransactionRequest) (*evm.TransactionResult, error)

	// Utility function to calculate the total fee based on a tx receipt
	CalculateTransactionFee(ctx context.Context, receiptGasInfo evm.ReceiptGasInfo) (*evm.TransactionFee, error)

	// GetTransactionStatus returns the current status of a transaction in the underlying chain's TXM.
	GetTransactionStatus(ctx context.Context, transactionID IdempotencyKey) (TransactionStatus, error)

	// GetForwarderForEOA returns a proper forwarder for a given EOA. If ocr2AggregatorID is non-empty the forwarder is searched within the ocr2AggregatorID contract scope.
	GetForwarderForEOA(ctx context.Context, eoa, ocr2AggregatorID evm.Address, pluginType string) (forwarder evm.Address, err error)
}

type TONService interface {
	ton.LiteClient

	// TXM
	SendTx(ctx context.Context, msg ton.Message) error
	GetTxStatus(ctx context.Context, lt uint64) (TransactionStatus, ton.ExitCode, error)
	GetTxExecutionFees(ctx context.Context, lt uint64) (*ton.TransactionFee, error)

	// LogPoller
	HasFilter(ctx context.Context, name string) bool
	RegisterFilter(ctx context.Context, filter ton.LPFilterQuery) error
	UnregisterFilter(ctx context.Context, name string) error
}

type SolanaService interface {
	solana.Client

	// Submits a transaction to the chain. It will return once the transaction is finalized or an error occurs.
	SubmitTransaction(ctx context.Context, req solana.SubmitTransactionRequest) (*solana.SubmitTransactionReply, error)

	// RegisterLogTracking registers a persistent log filter for tracking and caching logs
	// based on the provided filter parameters. Once registered, matching logs will be collected
	// over time and stored in a cache for future querying.
	// noop guaranteed when filter.Name exists
	RegisterLogTracking(ctx context.Context, req solana.LPFilterQuery) error

	// UnregisterLogTracking removes a previously registered log filter by its name.
	// After removal, logs matching this filter will no longer be collected or cached.
	// noop guaranteed when filterName doesn't exist
	UnregisterLogTracking(ctx context.Context, filterName string) error

	// QueryTrackedLogs retrieves logs from the  log storage based on the provided
	// query expression, sorting, and confidence level. It only returns logs that were
	// collected through previously registered log filters.
	QueryTrackedLogs(ctx context.Context, filterQuery []query.Expression,
		limitAndSort query.LimitAndSort) ([]*solana.Log, error)

	// GetLatestLPBlock retrieves current LatestBlock from cache perspective
	GetLatestLPBlock(ctx context.Context) (*solana.LPBlock, error)
}

// Relayer extends ChainService with providers for each product.
type Relayer interface {
	ChainService

	// EVM returns EVMService that provides access to evm-family specific functionalities
	EVM() (EVMService, error)
	// TON returns TONService that provides access to TON specific functionalities
	TON() (TONService, error)

	Solana() (SolanaService, error)

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

	NewCCIPProvider(ctx context.Context, cargs CCIPProviderArgs) (CCIPProvider, error)
}

var _ Relayer = &UnimplementedRelayer{}

// UnimplementedRelayer implements the Relayer interface with stubbed methods that return codes.Unimplemented errors or panic.
// It is meant to be embedded in real Relayer implementations in order to get default behavior for new methods without having
// to react to each change.
// In the future, embedding this type may be required to implement Relayer (through use of an unexported method).
type UnimplementedRelayer struct{}

func (u *UnimplementedRelayer) Name() string {
	panic("method Name not implemented")
}

func (u *UnimplementedRelayer) Start(ctx context.Context) error {
	return status.Errorf(codes.Unimplemented, "method Start not implemented")
}

func (u *UnimplementedRelayer) Close() error {
	return status.Errorf(codes.Unimplemented, "method Close not implemented")
}

func (u *UnimplementedRelayer) Ready() error {
	return status.Errorf(codes.Unimplemented, "method Ready not implemented")
}

func (u *UnimplementedRelayer) HealthReport() map[string]error {
	panic("method HealthReport not implemented")
}

func (u *UnimplementedRelayer) LatestHead(ctx context.Context) (Head, error) {
	return Head{}, status.Errorf(codes.Unimplemented, "method LatestHead not implemented")
}

func (u *UnimplementedRelayer) GetChainInfo(ctx context.Context) (ChainInfo, error) {
	return ChainInfo{}, status.Errorf(codes.Unimplemented, "method GetChainInfo not implemented")
}

func (u *UnimplementedRelayer) GetChainStatus(ctx context.Context) (ChainStatus, error) {
	return ChainStatus{}, status.Errorf(codes.Unimplemented, "method GetChainStatus not implemented")
}

func (u *UnimplementedRelayer) ListNodeStatuses(ctx context.Context, pageSize int32, pageToken string) (stats []NodeStatus, nextPageToken string, total int, err error) {
	return []NodeStatus{}, "", -1, status.Errorf(codes.Unimplemented, "method ListNodeStatuses not implemented")
}

func (u *UnimplementedRelayer) Transact(ctx context.Context, from, to string, amount *big.Int, balanceCheck bool) error {
	return status.Errorf(codes.Unimplemented, "method Transact not implemented")
}

func (u *UnimplementedRelayer) Replay(ctx context.Context, fromBlock string, args map[string]any) error {
	return status.Errorf(codes.Unimplemented, "method Replay not implemented")
}

func (u *UnimplementedRelayer) EVM() (EVMService, error) {
	return nil, status.Errorf(codes.Unimplemented, "method EVM not implemented")
}

func (u *UnimplementedRelayer) TON() (TONService, error) {
	return nil, status.Errorf(codes.Unimplemented, "method TON not implemented")
}

func (u *UnimplementedRelayer) Solana() (SolanaService, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Solana not implemented")
}

func (u *UnimplementedRelayer) NewContractWriter(ctx context.Context, config []byte) (ContractWriter, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewContractWriter not implemented")
}

func (u *UnimplementedRelayer) NewContractReader(ctx context.Context, contractReaderConfig []byte) (ContractReader, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewContractReader not implemented")
}

func (u *UnimplementedRelayer) NewConfigProvider(ctx context.Context, rargs RelayArgs) (ConfigProvider, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewConfigProvider not implemented")
}

func (u *UnimplementedRelayer) NewMedianProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (MedianProvider, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewMedianProvider not implemented")
}

func (u *UnimplementedRelayer) NewMercuryProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (MercuryProvider, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewMercuryProvider not implemented")
}

func (u *UnimplementedRelayer) NewFunctionsProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (FunctionsProvider, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewFunctionsProvider not implemented")
}

func (u *UnimplementedRelayer) NewAutomationProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (AutomationProvider, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewAutomationProvider not implemented")
}

func (u *UnimplementedRelayer) NewLLOProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (LLOProvider, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewLLOProvider not implemented")
}

func (u *UnimplementedRelayer) NewCCIPCommitProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (CCIPCommitProvider, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewCCIPCommitProvider not implemented")
}

func (u *UnimplementedRelayer) NewCCIPExecProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (CCIPExecProvider, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewCCIPExecProvider not implemented")
}

func (u *UnimplementedRelayer) NewPluginProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (PluginProvider, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewPluginProvider not implemented")
}

func (u *UnimplementedRelayer) NewOCR3CapabilityProvider(ctx context.Context, rargs RelayArgs, pargs PluginArgs) (OCR3CapabilityProvider, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewOCR3CapabilityProvider not implemented")
}

func (u *UnimplementedRelayer) NewCCIPProvider(ctx context.Context, cargs CCIPProviderArgs) (CCIPProvider, error) {
	return nil, status.Errorf(codes.Unimplemented, "method NewCCIPProvider not implemented")
}

var _ EVMService = &UnimplementedEVMService{}

// UnimplementedEVMService implements the EVMService interface with stubbed methods that return codes.Unimplemented errors or panic.
// It is meant to be embedded in real EVMService implementations in order to get default behavior for new methods without having
// to react to each change.
// In the future, embedding this type may be required to implement EVMService (through use of an unexported method).
type UnimplementedEVMService struct{}

func (ues *UnimplementedEVMService) BalanceAt(ctx context.Context, request evm.BalanceAtRequest) (*evm.BalanceAtReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method BalanceAt not implemented")
}

func (ues *UnimplementedEVMService) CallContract(ctx context.Context, request evm.CallContractRequest) (*evm.CallContractReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CallContract not implemented")
}

func (ues *UnimplementedEVMService) FilterLogs(ctx context.Context, request evm.FilterLogsRequest) (*evm.FilterLogsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method  not implemented")
}

func (ues *UnimplementedEVMService) HeaderByNumber(ctx context.Context, request evm.HeaderByNumberRequest) (*evm.HeaderByNumberReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method HeaderByNumber not implemented")
}

func (ues *UnimplementedEVMService) EstimateGas(ctx context.Context, call *evm.CallMsg) (uint64, error) {
	return 0, status.Errorf(codes.Unimplemented, "method EstimateGas not implemented")
}

func (ues *UnimplementedEVMService) GetTransactionByHash(ctx context.Context, request evm.GetTransactionByHashRequest) (*evm.Transaction, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTransactionByHash not implemented")
}

func (ues *UnimplementedEVMService) GetTransactionReceipt(ctx context.Context, request evm.GeTransactionReceiptRequest) (*evm.Receipt, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTransactionReceipt not implemented")
}

func (ues *UnimplementedEVMService) RegisterLogTracking(ctx context.Context, filter evm.LPFilterQuery) error {
	return status.Errorf(codes.Unimplemented, "method RegisterLogTracking not implemented")
}

func (ues *UnimplementedEVMService) UnregisterLogTracking(ctx context.Context, filterName string) error {
	return status.Errorf(codes.Unimplemented, "method UnregisterLogTracking not implemented")
}

func (ues *UnimplementedEVMService) QueryTrackedLogs(ctx context.Context, filterQuery []query.Expression,
	limitAndSort query.LimitAndSort, confidenceLevel primitives.ConfidenceLevel) ([]*evm.Log, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryTrackedLogs not implemented")
}

func (ues *UnimplementedEVMService) GetLatestLPBlock(ctx context.Context) (*evm.LPBlock, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLatestLPBlock not implemented")
}

func (ues *UnimplementedEVMService) GetFiltersNames(ctx context.Context) ([]string, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFiltersNames not implemented")
}

func (ues *UnimplementedEVMService) GetTransactionFee(ctx context.Context, transactionID IdempotencyKey) (*evm.TransactionFee, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTransactionFee not implemented")
}

func (ues *UnimplementedEVMService) SubmitTransaction(ctx context.Context, txRequest evm.SubmitTransactionRequest) (*evm.TransactionResult, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitTransaction not implemented")
}

func (ues *UnimplementedEVMService) CalculateTransactionFee(ctx context.Context, receiptGasInfo evm.ReceiptGasInfo) (*evm.TransactionFee, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CalculateTransactionFee not implemented")
}

func (ues *UnimplementedEVMService) GetTransactionStatus(ctx context.Context, transactionID IdempotencyKey) (TransactionStatus, error) {
	return 0, status.Errorf(codes.Unimplemented, "method GetTransactionStatus not implemented")
}

func (ues *UnimplementedEVMService) GetForwarderForEOA(ctx context.Context, eoa, ocr2AggregatorID evm.Address, pluginType string) (forwarder evm.Address, err error) {
	return evm.Address{}, status.Errorf(codes.Unimplemented, "method GetForwarderForEOA not implemented")
}

var _ SolanaService = &UnimplementedSolanaService{}

// UnimplementedSolanaService implements the SolanaService interface with stubbed methods that return codes.Unimplemented errors or panic.
// It is meant to be embedded in real SolanaService implementations in order to get default behavior for new methods without having
// to react to each change.
// In the future, embedding this type may be required to implement SolanaService (through use of an unexported method).
type UnimplementedSolanaService struct{}

func (uss *UnimplementedSolanaService) SubmitTransaction(ctx context.Context, req solana.SubmitTransactionRequest) (*solana.SubmitTransactionReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SubmitTransaction not implemented")
}

func (uss *UnimplementedSolanaService) RegisterLogTracking(ctx context.Context, req solana.LPFilterQuery) error {
	return status.Errorf(codes.Unimplemented, "method RegisterLogTracking not implemented")
}

func (uss *UnimplementedSolanaService) UnregisterLogTracking(ctx context.Context, filterName string) error {
	return status.Errorf(codes.Unimplemented, "method UnregisterLogTracking not implemented")
}
func (uss *UnimplementedSolanaService) QueryTrackedLogs(ctx context.Context, filterQuery []query.Expression, limitAndSort query.LimitAndSort) ([]*solana.Log, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryTrackedLogs not implemented")
}
func (uss *UnimplementedSolanaService) GetBalance(ctx context.Context, req solana.GetBalanceRequest) (*solana.GetBalanceReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBalance not implemented")
}
func (uss *UnimplementedSolanaService) GetAccountInfoWithOpts(ctx context.Context, req solana.GetAccountInfoRequest) (*solana.GetAccountInfoReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAccountInfoWithOpts not implemented")
}
func (uss *UnimplementedSolanaService) GetMultipleAccountsWithOpts(ctx context.Context, req solana.GetMultipleAccountsRequest) (*solana.GetMultipleAccountsReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMultipleAccountsWithOpts not implemented")
}
func (uss *UnimplementedSolanaService) GetBlock(ctx context.Context, req solana.GetBlockRequest) (*solana.GetBlockReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBlock not implemented")
}
func (uss *UnimplementedSolanaService) GetSlotHeight(ctx context.Context, req solana.GetSlotHeightRequest) (*solana.GetSlotHeightReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSlotHeight not implemented")
}
func (uss *UnimplementedSolanaService) GetTransaction(ctx context.Context, req solana.GetTransactionRequest) (*solana.GetTransactionReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetTransaction not implemented")
}
func (uss *UnimplementedSolanaService) GetFeeForMessage(ctx context.Context, req solana.GetFeeForMessageRequest) (*solana.GetFeeForMessageReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetFeeForMessage not implemented")
}
func (uss *UnimplementedSolanaService) GetSignatureStatuses(ctx context.Context, req solana.GetSignatureStatusesRequest) (*solana.GetSignatureStatusesReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetSignatureStatuses not implemented")
}
func (uss *UnimplementedSolanaService) SimulateTX(ctx context.Context, req solana.SimulateTXRequest) (*solana.SimulateTXReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SimulateTX not implemented")
}
func (uss *UnimplementedSolanaService) GetLatestLPBlock(ctx context.Context) (*solana.LPBlock, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetLatestLPBlock not implemented")
}
