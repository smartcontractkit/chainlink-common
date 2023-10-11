package internal

import (
	"context"

	"github.com/smartcontractkit/chainlink-relay/pkg/logger"
	"github.com/smartcontractkit/chainlink-relay/pkg/types"
)

type PluginRelayer interface {
	NewRelayer(ctx context.Context, config string, keystore types.Keystore) (Relayer, error)
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

// Relayer extends [types.Relayer] and includes [context.Context]s.
type Relayer interface {
	types.ChainService

	NewConfigProvider(context.Context, types.RelayArgs) (types.ConfigProvider, error)
	NewPluginProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.PluginProvider, error)
}

type PipelineRunner interface {
	ExecuteRun(ctx context.Context, spec types.Spec, vars types.Vars, l logger.Logger) (run *types.Run, trrs types.TaskRunResults, err error)
}
