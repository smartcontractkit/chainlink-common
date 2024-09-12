package relayerset

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/chainwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

type relayerClient struct {
	log              logger.Logger
	relayerSetClient *Client
	relayerID        types.RelayID
}

func newRelayerClient(log logger.Logger, client *Client, relayID types.RelayID) *relayerClient {
	return &relayerClient{log: log, relayerSetClient: client, relayerID: relayID}
}

func (r *relayerClient) NewPluginProvider(ctx context.Context, rargs core.RelayArgs, pargs core.PluginArgs) (types.PluginProvider, error) {
	cc := r.relayerSetClient.NewClientConn("PluginProvider", func(ctx context.Context) (uint32, net.Resources, error) {
		providerID, err := r.relayerSetClient.NewPluginProvider(ctx, r.relayerID, rargs, pargs)
		if err != nil {
			return 0, nil, fmt.Errorf("error getting plugin provider: %w", err)
		}

		return providerID, nil, nil
	})

	return relayer.WrapProviderClientConnection(rargs.ProviderType, cc, r.relayerSetClient.BrokerExt)
}

func (r *relayerClient) NewContractReader(_ context.Context, contractReaderConfig []byte) (types.ContractReader, error) {
	cc := r.relayerSetClient.NewClientConn("ContractReader", func(ctx context.Context) (uint32, net.Resources, error) {
		contractReaderID, err := r.relayerSetClient.NewContractReader(ctx, r.relayerID, contractReaderConfig)
		if err != nil {
			return 0, nil, fmt.Errorf("error getting NewContractReader from relayerSetServer: %w", err)
		}

		return contractReaderID, nil, nil
	})

	return contractreader.NewClient(r.relayerSetClient.BrokerExt.WithName("ContractReaderClientInRelayerSet"), cc), nil
}

func (r *relayerClient) NewChainWriter(_ context.Context, chainWriterConfig []byte) (types.ChainWriter, error) {
	cwc := r.relayerSetClient.NewClientConn("ChainWriter", func(ctx context.Context) (uint32, net.Resources, error) {
		chainWriterID, err := r.relayerSetClient.NewChainWriter(ctx, r.relayerID, chainWriterConfig)
		if err != nil {
			return 0, nil, err
		}
		return chainWriterID, nil, nil
	})
	return chainwriter.NewClient(r.relayerSetClient.BrokerExt.WithName("ChainWriterClient"), cwc), nil
}

func (r *relayerClient) Start(context.Context) error {
	return r.relayerSetClient.StartRelayer(context.Background(), r.relayerID)
}

func (r *relayerClient) Close() error {
	return r.relayerSetClient.CloseRelayer(context.Background(), r.relayerID)
}

func (r *relayerClient) Ready() error {
	return r.relayerSetClient.RelayerReady(context.Background(), r.relayerID)
}

func (r *relayerClient) HealthReport() map[string]error {
	report, err := r.relayerSetClient.RelayerHealthReport(context.Background(), r.relayerID)

	if err != nil {
		r.log.Error("error getting health report", "error", err)
		return map[string]error{}
	}

	return report
}

func (r *relayerClient) Name() string {
	name, err := r.relayerSetClient.RelayerName(context.Background(), r.relayerID)
	if err != nil {
		r.log.Error("error getting name", "error", err)
		return ""
	}

	return name
}
