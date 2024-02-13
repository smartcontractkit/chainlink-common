package types

import (
	"context"

	"github.com/smartcontractkit/chainlink/v2/core/services/ocr2/plugins/ccip/cciptypes"
)

type CCIPCommitProvider interface {
	PluginProvider

	NewOnRampReader(ctx context.Context, addr cciptypes.Address) (cciptypes.OnRampReader, error)
	NewOffRampReader(ctx context.Context, addr cciptypes.Address) (cciptypes.OffRampReader, error)
	NewCommitStoreReader(ctx context.Context, addr cciptypes.Address) (cciptypes.CommitStoreReader, error)
	NewPriceRegistryReader(ctx context.Context, addr cciptypes.Address) (cciptypes.PriceRegistryReader, error)
	NewPriceGetter(ctx context.Context) (cciptypes.PriceGetter, error)
	SourceNativeToken(ctx context.Context) (cciptypes.Address, error)
}

type CCIPExecProvider interface {
	PluginProvider

	NewOnRampReader(ctx context.Context, addr cciptypes.Address) (cciptypes.OnRampReader, error)
	NewOffRampReader(ctx context.Context, addr cciptypes.Address) (cciptypes.OffRampReader, error)
	NewCommitStoreReader(ctx context.Context, addr cciptypes.Address) (cciptypes.CommitStoreReader, error)
	NewPriceRegistryReader(ctx context.Context, addr cciptypes.Address) (cciptypes.PriceRegistryReader, error)
	NewTokenDataReader(ctx context.Context, tokenAddress cciptypes.Address) (cciptypes.TokenDataReader, error)
	SourceNativeToken(ctx context.Context) (cciptypes.Address, error)
}
