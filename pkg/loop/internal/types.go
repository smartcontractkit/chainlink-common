package internal

import (
	"context"

	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

type PluginRelayer interface {
	NewRelayer(ctx context.Context, config string, keystore types.Keystore) (Relayer, error)
}

// Relayer extends [types.Relayer] and includes [context.Context]s.
type Relayer interface {
	types.ChainService

	NewConfigProvider(context.Context, types.RelayArgs) (types.ConfigProvider, error)
	NewMedianProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.MedianProvider, error)
	NewMercuryProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.MercuryProvider, error)
	NewFunctionsProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.FunctionsProvider, error)
	/*
			GetChainStatus(ctx context.Context) (types.ChainStatus, error)
			ListNodeStatuses(ctx context.Context, page_size int32, page_token string) (stats []types.NodeStatus, next_page_token string, err error)
			// choose different name than SendTx to avoid collison during refactor.
			Transact(ctx context.Context, from, to string, amount *big.Int, balanceCheck bool) error

		// TODO remove in BCF-2441
		ChainStatus(ctx context.Context, id string) (types.ChainStatus, error)
		ChainStatuses(ctx context.Context, offset, limit int) (chains []types.ChainStatus, count int, err error)

		NodeStatuses(ctx context.Context, offset, limit int, chainIDs ...string) (nodes []types.NodeStatus, count int, err error)

		SendTx(ctx context.Context, chainID, from, to string, amount *big.Int, balanceCheck bool) error
	*/
}
