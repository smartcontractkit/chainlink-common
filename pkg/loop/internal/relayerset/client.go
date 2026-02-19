package relayerset

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/chains/aptos"
	"github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	"github.com/smartcontractkit/chainlink-common/pkg/chains/solana"
	"github.com/smartcontractkit/chainlink-common/pkg/chains/ton"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
	rel "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

type Client struct {
	*net.BrokerExt
	*goplugin.ServiceClient

	log logger.Logger

	relayerSetClient       relayerset.RelayerSetClient
	contractReaderClient   pb.ContractReaderClient
	evmRelayerSetClient    evm.EVMClient
	tonRelayerSetClient    ton.TONClient
	solanaRelayerSetClient solana.SolanaClient
	aptosRelayerSetClient  aptos.AptosClient
}

func NewRelayerSetClient(log logger.Logger, b *net.BrokerExt, conn grpc.ClientConnInterface) *Client {
	b = b.WithName("ChainRelayerClient")
	return &Client{
		log:                    log,
		BrokerExt:              b,
		ServiceClient:          goplugin.NewServiceClient(b, conn),
		relayerSetClient:       relayerset.NewRelayerSetClient(conn),
		evmRelayerSetClient:    evm.NewEVMClient(conn),
		tonRelayerSetClient:    ton.NewTONClient(conn),
		solanaRelayerSetClient: solana.NewSolanaClient(conn),
		aptosRelayerSetClient:  aptos.NewAptosClient(conn),
		contractReaderClient:   pb.NewContractReaderClient(conn)}
}

func (k *Client) Get(ctx context.Context, relayID types.RelayID) (core.Relayer, error) {
	_, err := k.relayerSetClient.Get(ctx, &relayerset.GetRelayerRequest{Id: &relayerset.RelayerId{ChainId: relayID.ChainID, Network: relayID.Network}})
	if err != nil {
		return nil, fmt.Errorf("error getting relayer: %w", err)
	}

	return newRelayer(k.log, k, relayID), nil
}

func (k *Client) List(ctx context.Context, relayIDs ...types.RelayID) (map[types.RelayID]core.Relayer, error) {
	ids := make([]*relayerset.RelayerId, 0, len(relayIDs))
	for _, id := range relayIDs {
		ids = append(ids, &relayerset.RelayerId{ChainId: id.ChainID, Network: id.Network})
	}

	resp, err := k.relayerSetClient.List(ctx, &relayerset.ListAllRelayersRequest{Ids: ids})
	if err != nil {
		return nil, fmt.Errorf("error getting all relayers: %w", err)
	}

	result := map[types.RelayID]core.Relayer{}
	for _, id := range resp.Ids {
		relayID := types.RelayID{ChainID: id.ChainId, Network: id.Network}
		result[relayID] = newRelayer(k.log, k, relayID)
	}

	return result, nil
}

func (k *Client) StartRelayer(ctx context.Context, relayID types.RelayID) error {
	_, err := k.relayerSetClient.StartRelayer(ctx, &relayerset.RelayerId{ChainId: relayID.ChainID, Network: relayID.Network})
	return err
}

func (k *Client) CloseRelayer(ctx context.Context, relayID types.RelayID) error {
	_, err := k.relayerSetClient.CloseRelayer(ctx, &relayerset.RelayerId{ChainId: relayID.ChainID, Network: relayID.Network})
	return err
}

func (k *Client) RelayerReady(ctx context.Context, relayID types.RelayID) error {
	_, err := k.relayerSetClient.RelayerReady(ctx, &relayerset.RelayerId{ChainId: relayID.ChainID, Network: relayID.Network})
	return err
}

func (k *Client) RelayerHealthReport(ctx context.Context, relayID types.RelayID) (map[string]error, error) {
	report, err := k.relayerSetClient.RelayerHealthReport(ctx, &relayerset.RelayerId{ChainId: relayID.ChainID, Network: relayID.Network})
	if err != nil {
		return nil, fmt.Errorf("error getting health report: %w", err)
	}

	result := map[string]error{}
	for k, v := range report.Report {
		result[k] = errors.New(v)
	}

	return result, nil
}

func (k *Client) RelayerName(ctx context.Context, relayID types.RelayID) (string, error) {
	resp, err := k.relayerSetClient.RelayerName(ctx, &relayerset.RelayerId{ChainId: relayID.ChainID, Network: relayID.Network})
	if err != nil {
		return "", fmt.Errorf("error getting name: %w", err)
	}

	return resp.Name, nil
}

func (k *Client) RelayerGetChainInfo(ctx context.Context, relayID types.RelayID) (types.ChainInfo, error) {
	req := &relayerset.GetChainInfoRequest{
		RelayerId: &relayerset.RelayerId{ChainId: relayID.ChainID, Network: relayID.Network},
	}

	chainInfoReply, err := k.relayerSetClient.RelayerGetChainInfo(ctx, req)
	if err != nil {
		return types.ChainInfo{}, fmt.Errorf("error getting chain info from relayerset client for relayer: %w", err)
	}

	chainInfo := chainInfoReply.GetChainInfo()
	return types.ChainInfo{
		FamilyName:      chainInfo.GetFamilyName(),
		ChainID:         chainInfo.GetChainId(),
		NetworkName:     chainInfo.GetNetworkName(),
		NetworkNameFull: chainInfo.GetNetworkNameFull(),
	}, nil
}

func (k *Client) RelayerLatestHead(ctx context.Context, relayID types.RelayID) (types.Head, error) {
	req := &relayerset.LatestHeadRequest{
		RelayerId: &relayerset.RelayerId{ChainId: relayID.ChainID, Network: relayID.Network},
	}
	resp, err := k.relayerSetClient.RelayerLatestHead(ctx, req)
	if err != nil {
		return types.Head{}, fmt.Errorf("error getting latest head from relayerset client for relayer: %w", err)
	}
	return types.Head{
		Height:    resp.Height,
		Hash:      resp.Hash,
		Timestamp: resp.Timestamp,
	}, nil
}

// EVM creates an EVM Relayer Set client which is a wrapper over the regular EVM client that attaches the Relayer ID to every request.
// This wrapper is then returned as a regular EVMClient .
func (k *Client) EVM(relayID types.RelayID) (types.EVMService, error) {
	if k.evmRelayerSetClient == nil {
		return nil, errors.New("evmRelayerSetClient can't be nil")
	}
	return rel.NewEVMCClient(&evmClient{
		relayID: relayID,
		client:  k.evmRelayerSetClient,
	}), nil
}

func (k *Client) TON(relayID types.RelayID) (types.TONService, error) {
	if k.tonRelayerSetClient == nil {
		return nil, errors.New("tonRelayerSetClient can't be nil")
	}
	return rel.NewTONClient(&tonClient{
		relayID: relayID,
		client:  k.tonRelayerSetClient,
	}), nil
}

func (k *Client) Solana(relayID types.RelayID) (types.SolanaService, error) {
	if k.solanaRelayerSetClient == nil {
		return nil, errors.New("solanaRelayerSetClient can't be nil")
	}

	return rel.NewSolanaClient(
		&solClient{
			relayID: relayID,
			client:  k.solanaRelayerSetClient,
		},
	), nil
}

func (k *Client) Aptos(relayID types.RelayID) (types.AptosService, error) {
	if k.aptosRelayerSetClient == nil {
		return nil, errors.New("aptosRelayerSetClient can't be nil")
	}
	return rel.NewAptosClient(&aptosClient{
		relayID: relayID,
		client:  k.aptosRelayerSetClient,
	}), nil
}

func (k *Client) NewPluginProvider(ctx context.Context, relayID types.RelayID, relayArgs core.RelayArgs, pluginArgs core.PluginArgs) (uint32, error) {
	// TODO at a later phase these credentials should be set as part of the relay config and not as a separate field
	var mercuryCredentials *relayerset.MercuryCredentials
	if relayArgs.MercuryCredentials != nil {
		mercuryCredentials = &relayerset.MercuryCredentials{
			LegacyUrl: relayArgs.MercuryCredentials.LegacyURL,
			Url:       relayArgs.MercuryCredentials.URL,
			Username:  relayArgs.MercuryCredentials.Username,
			Password:  relayArgs.MercuryCredentials.Password,
		}
	}

	req := &relayerset.NewPluginProviderRequest{
		RelayerId:  &relayerset.RelayerId{ChainId: relayID.ChainID, Network: relayID.Network},
		RelayArgs:  &relayerset.RelayArgs{ContractID: relayArgs.ContractID, RelayConfig: relayArgs.RelayConfig, ProviderType: relayArgs.ProviderType, MercuryCredentials: mercuryCredentials},
		PluginArgs: &relayerset.PluginArgs{TransmitterID: pluginArgs.TransmitterID, PluginConfig: pluginArgs.PluginConfig},
	}

	resp, err := k.relayerSetClient.NewPluginProvider(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("error getting new plugin provider: %w", err)
	}
	return resp.PluginProviderId, nil
}

func (k *Client) NewContractReader(ctx context.Context, relayID types.RelayID, contractReaderConfig []byte) (types.ContractReader, error) {
	req := &relayerset.NewContractReaderRequest{
		RelayerId:            &relayerset.RelayerId{ChainId: relayID.ChainID, Network: relayID.Network},
		ContractReaderConfig: contractReaderConfig,
	}
	resp, err := k.relayerSetClient.NewContractReader(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create new contract reader: %w", err)
	}

	serviceClient := &contractReaderServiceClient{
		contractReaderID: resp.ContractReaderId,
		client:           k,
	}

	reader := &contractReader{
		contractReaderID: resp.ContractReaderId,
		client:           k,
	}

	return contractreader.NewClient(serviceClient, reader), nil
}

func (k *Client) NewContractWriter(ctx context.Context, relayID types.RelayID, contractWriterConfig []byte) (uint32, error) {
	req := &relayerset.NewContractWriterRequest{
		RelayerId:            &relayerset.RelayerId{ChainId: relayID.ChainID, Network: relayID.Network},
		ContractWriterConfig: contractWriterConfig,
	}
	resp, err := k.relayerSetClient.NewContractWriter(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("error getting new contract writer: %w", err)
	}
	return resp.ContractWriterId, nil
}
