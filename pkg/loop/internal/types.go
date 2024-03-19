package internal

import (
	"context"

	"github.com/smartcontractkit/chainlink-common/pkg/types"
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

type AutomationProvider interface {
	NewAutomationProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.AutomationProvider, error)
}

type CCIPExecProvider interface {
	NewExecutionProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.CCIPExecProvider, error)
}

// Relayer extends [types.Relayer] and includes [context.Context]s.
type Relayer interface {
	types.ChainService

	NewConfigProvider(context.Context, types.RelayArgs) (types.ConfigProvider, error)
	NewPluginProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.PluginProvider, error)
	NewLLOProvider(context.Context, types.RelayArgs, types.PluginArgs) (types.LLOProvider, error)
}
