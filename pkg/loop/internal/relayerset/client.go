package relayerset

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"google.golang.org/protobuf/types/known/emptypb"
)

type connProvider interface {
	ClientConn() grpc.ClientConnInterface
}

type Client struct {
	*net.BrokerExt
	*goplugin.ServiceClient

	log logger.Logger

	relayerSetClient relayerset.RelayerSetClient
}

func NewRelayerSetClient(log logger.Logger, b *net.BrokerExt, conn grpc.ClientConnInterface) *Client {
	b = b.WithName("ChainRelayerClient")
	return &Client{log: log, BrokerExt: b, ServiceClient: goplugin.NewServiceClient(b, conn), relayerSetClient: relayerset.NewRelayerSetClient(conn)}
}

func (k *Client) Get(ctx context.Context, relayID types.RelayID) (core.Relayer, error) {
	_, err := k.relayerSetClient.Get(ctx, &relayerset.GetRelayerRequest{Id: &relayerset.RelayerId{ChainId: relayID.ChainID, Network: relayID.Network}})
	if err != nil {
		return nil, fmt.Errorf("error getting relayer: %w", err)
	}

	return newRelayerClient(k.log, k, relayID), nil
}

func (k *Client) List(ctx context.Context, relayIDs ...types.RelayID) (map[types.RelayID]core.Relayer, error) {
	var ids []*relayerset.RelayerId
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
		result[relayID] = newRelayerClient(k.log, k, relayID)
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

type contractReaderServiceClient struct {
	contractReaderID string
	client           *Client
}

func (s *contractReaderServiceClient) ClientConn() grpc.ClientConnInterface {
	return s.client.ClientConn()
}

func (s *contractReaderServiceClient) Start(ctx context.Context) error {
	_, err := s.client.relayerSetClient.ContractReaderStart(ctx, &relayerset.ContractReaderStartRequest{
		ContractReaderId: s.contractReaderID,
	})
	if err != nil {
		return fmt.Errorf("error starting contract reader: %w", err)
	}
	return nil
}

func (s *contractReaderServiceClient) Close() error {
	_, err := s.client.relayerSetClient.ContractReaderClose(context.Background(), &relayerset.ContractReaderCloseRequest{
		ContractReaderId: s.contractReaderID,
	})
	if err != nil {
		return fmt.Errorf("error closing contract reader: %w", err)
	}
	return nil
}

func (s *contractReaderServiceClient) HealthReport() map[string]error {
	return map[string]error{}
}
func (s *contractReaderServiceClient) Name() string {
	return "RelayerSetContractReader"
}

func (s *contractReaderServiceClient) Ready() error {
	return nil
}

type contractReader struct {
	contractReaderID string
	client           *Client
}

func (c *contractReader) GetLatestValue(ctx context.Context, in *pb.GetLatestValueRequest, opts ...grpc.CallOption) (*pb.GetLatestValueReply, error) {
	return c.client.relayerSetClient.ContractReaderGetLatestValue(ctx, &relayerset.ContractReaderGetLatestValueRequest{
		ContractReaderId: c.contractReaderID,
		Request:          in,
	}, opts...)
}

func (c *contractReader) GetLatestValueWithHeadData(ctx context.Context, in *pb.GetLatestValueRequest, opts ...grpc.CallOption) (*pb.GetLatestValueWithHeadDataReply, error) {
	return c.client.relayerSetClient.ContractReaderGetLatestValueWithHeadData(ctx, &relayerset.ContractReaderGetLatestValueRequest{
		ContractReaderId: c.contractReaderID,
		Request:          in,
	}, opts...)
}

func (c *contractReader) BatchGetLatestValues(ctx context.Context, in *pb.BatchGetLatestValuesRequest, opts ...grpc.CallOption) (*pb.BatchGetLatestValuesReply, error) {
	return c.client.relayerSetClient.ContractReaderBatchGetLatestValues(ctx, &relayerset.ContractReaderBatchGetLatestValuesRequest{
		ContractReaderId: c.contractReaderID,
		Request:          in,
	}, opts...)
}

func (c *contractReader) QueryKey(ctx context.Context, in *pb.QueryKeyRequest, opts ...grpc.CallOption) (*pb.QueryKeyReply, error) {
	return c.client.relayerSetClient.ContractReaderQueryKey(ctx, &relayerset.ContractReaderQueryKeyRequest{
		ContractReaderId: c.contractReaderID,
		Request:          in,
	}, opts...)
}

func (c *contractReader) QueryKeys(ctx context.Context, in *pb.QueryKeysRequest, opts ...grpc.CallOption) (*pb.QueryKeysReply, error) {
	return c.client.relayerSetClient.ContractReaderQueryKeys(ctx, &relayerset.ContractReaderQueryKeysRequest{
		ContractReaderId: c.contractReaderID,
		Request:          in,
	}, opts...)
}

func (c *contractReader) Bind(ctx context.Context, in *pb.BindRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return c.client.relayerSetClient.ContractReaderBind(ctx, &relayerset.ContractReaderBindRequest{
		ContractReaderId: c.contractReaderID,
		Request:          in,
	}, opts...)
}

func (c *contractReader) Unbind(ctx context.Context, in *pb.UnbindRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	return c.client.relayerSetClient.ContractReaderUnbind(ctx, &relayerset.ContractReaderUnbindRequest{
		ContractReaderId: c.contractReaderID,
		Request:          in,
	}, opts...)
}
