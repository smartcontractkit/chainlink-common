package ccip

import (
	"context"

	libocr "github.com/smartcontractkit/libocr/offchainreporting2plus/types"

	ccippb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccip"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/ccip"
)

var _ types.CCIPExecProvider = (*ExecClient)(nil)

type ExecClient struct {
	impl ccippb.ExecutionFactoryGeneratorClient
}

// ChainReader implements types.CCIPExecProvider.
func (e *ExecClient) ChainReader() types.ChainReader {
	panic("unimplemented")
}

// Close implements types.CCIPExecProvider.
func (e *ExecClient) Close() error {
	panic("unimplemented")
}

// Codec implements types.CCIPExecProvider.
func (e *ExecClient) Codec() types.Codec {
	panic("unimplemented")
}

// ContractConfigTracker implements types.CCIPExecProvider.
func (e *ExecClient) ContractConfigTracker() libocr.ContractConfigTracker {
	panic("unimplemented")
}

// ContractTransmitter implements types.CCIPExecProvider.
func (e *ExecClient) ContractTransmitter() libocr.ContractTransmitter {
	panic("unimplemented")
}

// HealthReport implements types.CCIPExecProvider.
func (e *ExecClient) HealthReport() map[string]error {
	panic("unimplemented")
}

// Name implements types.CCIPExecProvider.
func (e *ExecClient) Name() string {
	panic("unimplemented")
}

// NewCommitStoreReader implements types.CCIPExecProvider.
func (e *ExecClient) NewCommitStoreReader(ctx context.Context, addr ccip.Address) (ccip.CommitStoreReader, error) {
	panic("unimplemented")
}

// NewOffRampReader implements types.CCIPExecProvider.
func (e *ExecClient) NewOffRampReader(ctx context.Context, addr ccip.Address) (ccip.OffRampReader, error) {
	panic("unimplemented")
}

// NewOnRampReader implements types.CCIPExecProvider.
func (e *ExecClient) NewOnRampReader(ctx context.Context, addr ccip.Address) (ccip.OnRampReader, error) {
	panic("unimplemented")
}

// NewPriceRegistryReader implements types.CCIPExecProvider.
func (e *ExecClient) NewPriceRegistryReader(ctx context.Context, addr ccip.Address) (ccip.PriceRegistryReader, error) {
	panic("unimplemented")
}

// NewTokenDataReader implements types.CCIPExecProvider.
func (e *ExecClient) NewTokenDataReader(ctx context.Context, tokenAddress ccip.Address) (ccip.TokenDataReader, error) {
	panic("unimplemented")
}

// NewTokenPoolBatchedReader implements types.CCIPExecProvider.
func (e *ExecClient) NewTokenPoolBatchedReader(ctx context.Context) (ccip.TokenPoolBatchedReader, error) {
	panic("unimplemented")
}

// OffchainConfigDigester implements types.CCIPExecProvider.
func (e *ExecClient) OffchainConfigDigester() libocr.OffchainConfigDigester {
	panic("unimplemented")
}

// Ready implements types.CCIPExecProvider.
func (e *ExecClient) Ready() error {
	panic("unimplemented")
}

// SourceNativeToken implements types.CCIPExecProvider.
func (e *ExecClient) SourceNativeToken(ctx context.Context) (ccip.Address, error) {
	panic("unimplemented")
}

// Start implements types.CCIPExecProvider.
func (e *ExecClient) Start(context.Context) error {
	panic("unimplemented")
}
