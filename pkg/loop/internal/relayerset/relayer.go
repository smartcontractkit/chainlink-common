package relayerset

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	rel "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

type relayer struct {
	log              logger.Logger
	relayerSetClient *Client
	relayerID        types.RelayID
}

func newRelayer(log logger.Logger, client *Client, relayID types.RelayID) *relayer {
	return &relayer{log: log, relayerSetClient: client, relayerID: relayID}
}

func (r *relayer) NewPluginProvider(ctx context.Context, rargs core.RelayArgs, pargs core.PluginArgs) (core.PluginProvider, error) {
	cc := r.relayerSetClient.NewClientConn("PluginProvider", func(ctx context.Context) (uint32, net.Resources, error) {
		providerID, err := r.relayerSetClient.NewPluginProvider(ctx, r.relayerID, rargs, pargs)
		if err != nil {
			return 0, nil, fmt.Errorf("error getting plugin provider: %w", err)
		}

		return providerID, nil, nil
	})

	return rel.WrapProviderClientConnection(ctx, rargs.ProviderType, cc, r.relayerSetClient.BrokerExt)
}

func (r *relayer) EVM() (types.EVMService, error) {
	return r.relayerSetClient.EVM(r.relayerID)
}

func (r *relayer) TON() (types.TONService, error) {
	return r.relayerSetClient.TON(r.relayerID)
}

func (r *relayer) Solana() (types.SolanaService, error) {
	return r.relayerSetClient.Solana(r.relayerID)
}

func (r *relayer) Aptos() (types.AptosService, error) {
	return r.relayerSetClient.Aptos(r.relayerID)
}

func (r *relayer) NewContractReader(ctx context.Context, contractReaderConfig []byte) (types.ContractReader, error) {
	return r.relayerSetClient.NewContractReader(ctx, r.relayerID, contractReaderConfig)
}

func (r *relayer) NewContractWriter(_ context.Context, contractWriterConfig []byte) (types.ContractWriter, error) {
	cwc := r.relayerSetClient.NewClientConn("ContractWriter", func(ctx context.Context) (uint32, net.Resources, error) {
		contractWriterID, err := r.relayerSetClient.NewContractWriter(ctx, r.relayerID, contractWriterConfig)
		if err != nil {
			return 0, nil, err
		}
		return contractWriterID, nil, nil
	})
	return contractwriter.NewClient(r.relayerSetClient.BrokerExt.WithName("ContractWriterClient"), cwc), nil
}

func (r *relayer) Start(ctx context.Context) error {
	return r.relayerSetClient.StartRelayer(ctx, r.relayerID)
}

func (r *relayer) Close() error {
	return r.relayerSetClient.CloseRelayer(context.Background(), r.relayerID)
}

func (r *relayer) Ready() error {
	return r.relayerSetClient.RelayerReady(context.Background(), r.relayerID)
}

func (r *relayer) HealthReport() map[string]error {
	report, err := r.relayerSetClient.RelayerHealthReport(context.Background(), r.relayerID)

	if err != nil {
		r.log.Error("error getting health report", "error", err)
		return map[string]error{}
	}

	return report
}

func (r *relayer) Name() string {
	name, err := r.relayerSetClient.RelayerName(context.Background(), r.relayerID)
	if err != nil {
		r.log.Error("error getting name", "error", err)
		return ""
	}

	return name
}

func (r *relayer) GetChainInfo(ctx context.Context) (types.ChainInfo, error) {
	chainInfo, err := r.relayerSetClient.RelayerGetChainInfo(ctx, r.relayerID)
	if err != nil {
		r.log.Error("error getting chain info", "error", err)
		return types.ChainInfo{}, err
	}
	return chainInfo, nil
}

func (r *relayer) LatestHead(ctx context.Context) (types.Head, error) {
	latestHead, err := r.relayerSetClient.RelayerLatestHead(ctx, r.relayerID)
	if err != nil {
		r.log.Error("error getting latestHead", "error", err)
		return types.Head{}, err
	}
	return latestHead, err
}
