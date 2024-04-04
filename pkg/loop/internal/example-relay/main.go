// This file contains an example implementation of a relayer plugin.
package main

import (
	"context"
	"errors"
	"math/big"

	"github.com/hashicorp/go-plugin"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop"
	"github.com/smartcontractkit/chainlink-common/pkg/sqlutil"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

const (
	loggerName = "PluginExample"
)

func main() {
	s := loop.MustNewStartedServer(loggerName)
	defer s.Stop()

	p := &pluginRelayer{lggr: s.Logger, ds: s.DataSource}
	defer s.Logger.ErrorIfFn(p.Close, "Failed to close")

	s.MustRegister(p)

	stopCh := make(chan struct{})
	defer close(stopCh)

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: loop.PluginRelayerHandshakeConfig(),
		Plugins: map[string]plugin.Plugin{
			loop.PluginRelayerName: &loop.GRPCPluginRelayer{
				PluginServer: p,
				BrokerConfig: loop.BrokerConfig{
					StopCh:   stopCh,
					Logger:   s.Logger,
					GRPCOpts: s.GRPCOpts,
				},
			},
		},
		GRPCServer: s.GRPCOpts.NewServer,
	})
}

type pluginRelayer struct {
	lggr logger.Logger
	ds   sqlutil.DataSource
}

func (p *pluginRelayer) Ready() error { return nil }

func (p *pluginRelayer) HealthReport() map[string]error { return map[string]error{p.Name(): nil} }

func (p *pluginRelayer) Name() string { return p.lggr.Name() }

func (p *pluginRelayer) NewRelayer(ctx context.Context, config string, keystore core.Keystore, cr core.CapabilitiesRegistry) (loop.Relayer, error) {
	return &relayer{lggr: logger.Named(p.lggr, "Relayer"), ds: p.ds}, nil
}

func (p *pluginRelayer) Close() error { return nil }

type relayer struct {
	lggr logger.Logger
	ds   sqlutil.DataSource
}

func (r *relayer) Name() string { return r.lggr.Name() }

func (r *relayer) Start(ctx context.Context) error {
	var names []string
	err := r.ds.SelectContext(ctx, names, "SELECT table_name FROM information_schema.tables WHERE table_schema='public'")
	if err != nil {
		return err
	}
	r.lggr.Info("Queried table names", "names", names)
	return nil
}

func (r *relayer) Close() error { return nil }

func (r *relayer) Ready() error { return nil }

func (r *relayer) HealthReport() map[string]error { return map[string]error{r.Name(): nil} }

func (r *relayer) LatestHead(ctx context.Context) (types.Head, error) {
	return types.Head{}, errors.New("unimplemented")
}

func (r *relayer) GetChainStatus(ctx context.Context) (types.ChainStatus, error) {
	return types.ChainStatus{}, errors.New("unimplemented")
}

func (r *relayer) ListNodeStatuses(ctx context.Context, pageSize int32, pageToken string) (stats []types.NodeStatus, nextPageToken string, total int, err error) {
	return nil, "", -1, errors.New("unimplemented")
}

func (r *relayer) Transact(ctx context.Context, from, to string, amount *big.Int, balanceCheck bool) error {
	return errors.New("unimplemented")
}

func (r *relayer) NewContractWriter(ctx context.Context, chainWriterConfig []byte) (types.ContractWriter, error) {
	return nil, errors.New("unimplemented")
}

func (r *relayer) NewContractReader(ctx context.Context, contractReaderConfig []byte) (types.ContractReader, error) {
	return nil, errors.New("unimplemented")
}

func (r *relayer) NewConfigProvider(ctx context.Context, args types.RelayArgs) (types.ConfigProvider, error) {
	return nil, errors.New("unimplemented")
}

func (r *relayer) NewPluginProvider(ctx context.Context, args types.RelayArgs, args2 types.PluginArgs) (types.PluginProvider, error) {
	return nil, errors.New("unimplemented")
}

func (r *relayer) NewLLOProvider(ctx context.Context, args types.RelayArgs, args2 types.PluginArgs) (types.LLOProvider, error) {
	return nil, errors.New("unimplemented")
}
