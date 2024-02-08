package types

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/types/cciptypes"
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
	NewUSDCReader(ctx context.Context) (cciptypes.USDCReader, error)
	SourceNativeToken(ctx context.Context) (cciptypes.Address, error)
}
