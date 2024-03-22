package internal

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	mercury_common_internal "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/mercury/common"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	mercury_pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/mercury"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
)

var _ PluginRelayer = (*PluginRelayerClient)(nil)

type PluginRelayerClient struct {
	*PluginClient

	grpc pb.PluginRelayerClient
}

func NewPluginRelayerClient(broker net.Broker, brokerCfg net.BrokerConfig, conn *grpc.ClientConn) *PluginRelayerClient {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "PluginRelayerClient")
	pc := NewPluginClient(broker, brokerCfg, conn)
	return &PluginRelayerClient{PluginClient: pc, grpc: pb.NewPluginRelayerClient(pc)}
}

func (p *PluginRelayerClient) NewRelayer(ctx context.Context, config string, keystore types.Keystore) (Relayer, error) {
	cc := p.NewClientConn("Relayer", func(ctx context.Context) (id uint32, deps net.Resources, err error) {
		var ksRes net.Resource
		id, ksRes, err = p.ServeNew("Keystore", func(s *grpc.Server) {
			pb.RegisterKeystoreServer(s, &keystoreServer{impl: keystore})
		})
		if err != nil {
			return 0, nil, fmt.Errorf("Failed to create relayer client: failed to serve keystore: %w", err)
		}
		deps.Add(ksRes)

		reply, err := p.grpc.NewRelayer(ctx, &pb.NewRelayerRequest{
			Config:     config,
			KeystoreID: id,
		})
		if err != nil {
			return 0, nil, fmt.Errorf("Failed to create relayer client: failed request: %w", err)
		}
		return reply.RelayerID, nil, nil
	})
	return newRelayerClient(p.BrokerExt, cc), nil
}

type pluginRelayerServer struct {
	pb.UnimplementedPluginRelayerServer

	*net.BrokerExt

	impl PluginRelayer
}

func RegisterPluginRelayerServer(server *grpc.Server, broker net.Broker, brokerCfg net.BrokerConfig, impl PluginRelayer) error {
	pb.RegisterPluginRelayerServer(server, newPluginRelayerServer(broker, brokerCfg, impl))
	return nil
}

func newPluginRelayerServer(broker net.Broker, brokerCfg net.BrokerConfig, impl PluginRelayer) *pluginRelayerServer {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "RelayerPluginServer")
	return &pluginRelayerServer{BrokerExt: &net.BrokerExt{Broker: broker, BrokerConfig: brokerCfg}, impl: impl}
}

func (p *pluginRelayerServer) NewRelayer(ctx context.Context, request *pb.NewRelayerRequest) (*pb.NewRelayerReply, error) {
	ksConn, err := p.Dial(request.KeystoreID)
	if err != nil {
		return nil, net.ErrConnDial{Name: "Keystore", ID: request.KeystoreID, Err: err}
	}
	ksRes := net.Resource{Closer: ksConn, Name: "Keystore"}
	r, err := p.impl.NewRelayer(ctx, request.Config, newKeystoreClient(ksConn))
	if err != nil {
		p.CloseAll(ksRes)
		return nil, err
	}
	err = r.Start(ctx)
	if err != nil {
		p.CloseAll(ksRes)
		return nil, err
	}

	const name = "Relayer"
	rRes := net.Resource{Closer: r, Name: name}
	id, _, err := p.ServeNew(name, func(s *grpc.Server) {
		pb.RegisterServiceServer(s, &ServiceServer{Srv: r})
		pb.RegisterRelayerServer(s, newChainRelayerServer(r, p.BrokerExt))
	}, rRes, ksRes)
	if err != nil {
		return nil, err
	}

	return &pb.NewRelayerReply{RelayerID: id}, nil
}

var _ types.Keystore = (*keystoreClient)(nil)

type keystoreClient struct {
	grpc pb.KeystoreClient
}

func newKeystoreClient(cc *grpc.ClientConn) *keystoreClient {
	return &keystoreClient{pb.NewKeystoreClient(cc)}
}

func (k *keystoreClient) Accounts(ctx context.Context) (accounts []string, err error) {
	reply, err := k.grpc.Accounts(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	return reply.Accounts, nil
}

func (k *keystoreClient) Sign(ctx context.Context, account string, data []byte) ([]byte, error) {
	reply, err := k.grpc.Sign(ctx, &pb.SignRequest{Account: account, Data: data})
	if err != nil {
		return nil, err
	}
	return reply.SignedData, nil
}

var _ pb.KeystoreServer = (*keystoreServer)(nil)

type keystoreServer struct {
	pb.UnimplementedKeystoreServer

	impl types.Keystore
}

func (k *keystoreServer) Accounts(ctx context.Context, _ *emptypb.Empty) (*pb.AccountsReply, error) {
	as, err := k.impl.Accounts(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.AccountsReply{Accounts: as}, nil
}

func (k *keystoreServer) Sign(ctx context.Context, request *pb.SignRequest) (*pb.SignReply, error) {
	signed, err := k.impl.Sign(ctx, request.Account, request.Data)
	if err != nil {
		return nil, err
	}
	return &pb.SignReply{SignedData: signed}, nil
}

var _ Relayer = (*relayerClient)(nil)

// relayerClient adapts a GRPC [pb.RelayerClient] to implement [Relayer].
type relayerClient struct {
	*net.BrokerExt
	*ServiceClient

	relayer pb.RelayerClient
}

func newRelayerClient(b *net.BrokerExt, conn grpc.ClientConnInterface) *relayerClient {
	b = b.WithName("ChainRelayerClient")
	return &relayerClient{b, NewServiceClient(b, conn), pb.NewRelayerClient(conn)}
}

func (r *relayerClient) NewConfigProvider(ctx context.Context, rargs types.RelayArgs) (types.ConfigProvider, error) {
	cc := r.NewClientConn("ConfigProvider", func(ctx context.Context) (uint32, net.Resources, error) {
		reply, err := r.relayer.NewConfigProvider(ctx, &pb.NewConfigProviderRequest{
			RelayArgs: &pb.RelayArgs{
				ExternalJobID: rargs.ExternalJobID[:],
				JobID:         rargs.JobID,
				ContractID:    rargs.ContractID,
				New:           rargs.New,
				RelayConfig:   rargs.RelayConfig,
			},
		})
		if err != nil {
			return 0, nil, err
		}
		return reply.ConfigProviderID, nil, nil
	})
	return newConfigProviderClient(r.WithName("ConfigProviderClient"), cc), nil
}

func (r *relayerClient) NewPluginProvider(ctx context.Context, rargs types.RelayArgs, pargs types.PluginArgs) (types.PluginProvider, error) {
	cc := r.NewClientConn("PluginProvider", func(ctx context.Context) (uint32, net.Resources, error) {
		reply, err := r.relayer.NewPluginProvider(ctx, &pb.NewPluginProviderRequest{
			RelayArgs: &pb.RelayArgs{
				ExternalJobID: rargs.ExternalJobID[:],
				JobID:         rargs.JobID,
				ContractID:    rargs.ContractID,
				New:           rargs.New,
				RelayConfig:   rargs.RelayConfig,
				ProviderType:  rargs.ProviderType,
			},
			PluginArgs: &pb.PluginArgs{
				TransmitterID: pargs.TransmitterID,
				PluginConfig:  pargs.PluginConfig,
			},
		})
		if err != nil {
			return 0, nil, err
		}
		return reply.PluginProviderID, nil, nil
	})

	// TODO: Remove this when we have fully transitioned all relayers to running in LOOPPs.
	// This allows callers to type assert a PluginProvider into a product provider type (eg. MedianProvider)
	// for interoperability with legacy code.
	switch rargs.ProviderType {
	case string(types.Median):
		return newMedianProviderClient(r.BrokerExt, cc), nil
	case string(types.GenericPlugin):
		return newPluginProviderClient(r.BrokerExt, cc), nil
	case string(types.Mercury):
		return newMercuryProviderClient(r.BrokerExt, cc), nil
	case string(types.CCIPExecution):
		// TODO BCF-3061
		// what do i do here? for the local embedded relayer, we are using the broker
		// to share state so the the reporting plugin, as a client to the (embedded) relayer,
		// calls the provider to get network.resources and then the provider calls the broker to serve them
		// maybe the same mechanism can be used here, but we need very careful reference passing to
		// ensure that this relayer client has the same broker as the server. that doesn't really
		// even make sense to me because the relayer client will in the reporting plugin loop
		// for now we return an error and test for the this error case
		// return nil, fmt.Errorf("need to fix BCF-3061")
		return newExecProviderClient(r.BrokerExt, cc), fmt.Errorf("need to fix BCF-3061")
	default:
		return nil, fmt.Errorf("provider type not supported: %s", rargs.ProviderType)
	}
}

func (r *relayerClient) NewLLOProvider(ctx context.Context, rargs types.RelayArgs, pargs types.PluginArgs) (types.LLOProvider, error) {
	return nil, fmt.Errorf("llo provider not supported: %w", errors.ErrUnsupported)
}

func (r *relayerClient) GetChainStatus(ctx context.Context) (types.ChainStatus, error) {
	reply, err := r.relayer.GetChainStatus(ctx, &pb.GetChainStatusRequest{})
	if err != nil {
		return types.ChainStatus{}, err
	}

	return types.ChainStatus{
		ID:      reply.Chain.Id,
		Enabled: reply.Chain.Enabled,
		Config:  reply.Chain.Config,
	}, nil
}

func (r *relayerClient) ListNodeStatuses(ctx context.Context, pageSize int32, pageToken string) (nodes []types.NodeStatus, nextPageToken string, total int, err error) {
	reply, err := r.relayer.ListNodeStatuses(ctx, &pb.ListNodeStatusesRequest{
		PageSize:  pageSize,
		PageToken: pageToken,
	})
	if err != nil {
		return nil, "", -1, err
	}
	for _, n := range reply.Nodes {
		nodes = append(nodes, types.NodeStatus{
			ChainID: n.ChainID,
			Name:    n.Name,
			Config:  n.Config,
			State:   n.State,
		})
	}
	total = int(reply.Total)
	return
}

func (r *relayerClient) Transact(ctx context.Context, from, to string, amount *big.Int, balanceCheck bool) error {
	_, err := r.relayer.Transact(ctx, &pb.TransactionRequest{
		From:         from,
		To:           to,
		Amount:       pb.NewBigIntFromInt(amount),
		BalanceCheck: balanceCheck,
	})
	return err
}

var _ pb.RelayerServer = (*relayerServer)(nil)

// relayerServer exposes [Relayer] as a GRPC [pb.RelayerServer].
type relayerServer struct {
	pb.UnimplementedRelayerServer

	*net.BrokerExt

	impl Relayer
}

func newChainRelayerServer(impl Relayer, b *net.BrokerExt) *relayerServer {
	return &relayerServer{impl: impl, BrokerExt: b.WithName("ChainRelayerServer")}
}

func (r *relayerServer) NewConfigProvider(ctx context.Context, request *pb.NewConfigProviderRequest) (*pb.NewConfigProviderReply, error) {
	exJobID, err := uuid.FromBytes(request.RelayArgs.ExternalJobID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid bytes for ExternalJobID: %w", err)
	}
	cp, err := r.impl.NewConfigProvider(ctx, types.RelayArgs{
		ExternalJobID: exJobID,
		JobID:         request.RelayArgs.JobID,
		ContractID:    request.RelayArgs.ContractID,
		New:           request.RelayArgs.New,
		RelayConfig:   request.RelayArgs.RelayConfig,
	})
	if err != nil {
		return nil, err
	}
	err = cp.Start(ctx)
	if err != nil {
		return nil, err
	}

	const name = "ConfigProvider"
	id, _, err := r.ServeNew(name, func(s *grpc.Server) {
		pb.RegisterServiceServer(s, &ServiceServer{Srv: cp})
		pb.RegisterOffchainConfigDigesterServer(s, &offchainConfigDigesterServer{impl: cp.OffchainConfigDigester()})
		pb.RegisterContractConfigTrackerServer(s, &contractConfigTrackerServer{impl: cp.ContractConfigTracker()})
	}, net.Resource{Closer: cp, Name: name})
	if err != nil {
		return nil, err
	}

	return &pb.NewConfigProviderReply{ConfigProviderID: id}, nil
}

func (r *relayerServer) NewPluginProvider(ctx context.Context, request *pb.NewPluginProviderRequest) (*pb.NewPluginProviderReply, error) {
	exJobID, err := uuid.FromBytes(request.RelayArgs.ExternalJobID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid bytes for ExternalJobID: %w", err)
	}
	relayArgs := types.RelayArgs{
		ExternalJobID: exJobID,
		JobID:         request.RelayArgs.JobID,
		ContractID:    request.RelayArgs.ContractID,
		New:           request.RelayArgs.New,
		RelayConfig:   request.RelayArgs.RelayConfig,
		ProviderType:  request.RelayArgs.ProviderType,
	}
	pluginArgs := types.PluginArgs{
		TransmitterID: request.PluginArgs.TransmitterID,
		PluginConfig:  request.PluginArgs.PluginConfig,
	}

	switch request.RelayArgs.ProviderType {
	case string(types.Median):
		id, err := r.newMedianProvider(ctx, relayArgs, pluginArgs)
		if err != nil {
			return nil, err
		}
		return &pb.NewPluginProviderReply{PluginProviderID: id}, nil
	case string(types.GenericPlugin):
		id, err := r.newPluginProvider(ctx, relayArgs, pluginArgs)
		if err != nil {
			return nil, err
		}
		return &pb.NewPluginProviderReply{PluginProviderID: id}, nil
	case string(types.Mercury):
		id, err := r.newMercuryProvider(ctx, relayArgs, pluginArgs)
		if err != nil {
			return nil, err
		}
		return &pb.NewPluginProviderReply{PluginProviderID: id}, nil
	case string(types.CCIPExecution):
		id, err := r.newExecProvider(ctx, relayArgs, pluginArgs)
		if err != nil {
			return nil, err
		}
		return &pb.NewPluginProviderReply{PluginProviderID: id}, nil
	}
	return nil, fmt.Errorf("provider type not supported: %s", relayArgs.ProviderType)
}

func (r *relayerServer) newMedianProvider(ctx context.Context, relayArgs types.RelayArgs, pluginArgs types.PluginArgs) (uint32, error) {
	i, ok := r.impl.(MedianProvider)
	if !ok {
		return 0, status.Error(codes.Unimplemented, "median not supported")
	}

	provider, err := i.NewMedianProvider(ctx, relayArgs, pluginArgs)
	if err != nil {
		return 0, err
	}
	err = provider.Start(ctx)
	if err != nil {
		return 0, err
	}
	const name = "MedianProvider"
	providerRes := net.Resource{Name: name, Closer: provider}

	id, _, err := r.ServeNew(name, func(s *grpc.Server) {
		registerPluginProviderServices(s, provider)
		pb.RegisterReportCodecServer(s, &reportCodecServer{impl: provider.ReportCodec()})
		pb.RegisterMedianContractServer(s, &medianContractServer{impl: provider.MedianContract()})
		pb.RegisterOnchainConfigCodecServer(s, &onchainConfigCodecServer{impl: provider.OnchainConfigCodec()})
	}, providerRes)
	if err != nil {
		return 0, err
	}

	return id, err
}

func (r *relayerServer) newPluginProvider(ctx context.Context, relayArgs types.RelayArgs, pluginArgs types.PluginArgs) (uint32, error) {
	provider, err := r.impl.NewPluginProvider(ctx, relayArgs, pluginArgs)
	if err != nil {
		return 0, err
	}
	err = provider.Start(ctx)
	if err != nil {
		return 0, err
	}
	const name = "PluginProvider"
	providerRes := net.Resource{Name: name, Closer: provider}

	id, _, err := r.ServeNew(name, func(s *grpc.Server) {
		registerPluginProviderServices(s, provider)
	}, providerRes)
	if err != nil {
		return 0, err
	}

	return id, err
}

func (r *relayerServer) newMercuryProvider(ctx context.Context, relayArgs types.RelayArgs, pluginArgs types.PluginArgs) (uint32, error) {
	i, ok := r.impl.(MercuryProvider)
	if !ok {
		return 0, status.Error(codes.Unimplemented, fmt.Sprintf("mercury not supported by %T", r.impl))
	}

	provider, err := i.NewMercuryProvider(ctx, relayArgs, pluginArgs)
	if err != nil {
		return 0, err
	}
	err = provider.Start(ctx)
	if err != nil {
		return 0, err
	}
	const name = "MercuryProvider"
	providerRes := net.Resource{Name: name, Closer: provider}

	id, _, err := r.ServeNew(name, func(s *grpc.Server) {
		registerPluginProviderServices(s, provider)

		mercury_pb.RegisterOnchainConfigCodecServer(s, mercury_common_internal.NewOnchainConfigCodecServer(provider.OnchainConfigCodec()))
		mercury_pb.RegisterReportCodecV1Server(s, mercury_common_internal.NewReportCodecV1Server(s, provider.ReportCodecV1()))
		mercury_pb.RegisterReportCodecV2Server(s, mercury_common_internal.NewReportCodecV2Server(s, provider.ReportCodecV2()))
		mercury_pb.RegisterReportCodecV3Server(s, mercury_common_internal.NewReportCodecV3Server(s, provider.ReportCodecV3()))
		mercury_pb.RegisterServerFetcherServer(s, mercury_common_internal.NewServerFetcherServer(provider.MercuryServerFetcher()))
		mercury_pb.RegisterMercuryChainReaderServer(s, mercury_common_internal.NewChainReaderServer(provider.MercuryChainReader()))
	}, providerRes)
	if err != nil {
		return 0, err
	}

	return id, err
}

func (r *relayerServer) newExecProvider(ctx context.Context, relayArgs types.RelayArgs, pluginArgs types.PluginArgs) (uint32, error) {
	i, ok := r.impl.(CCIPExecProvider)
	if !ok {
		return 0, status.Error(codes.Unimplemented, fmt.Sprintf("ccip execution not supported by %T", r.impl))
	}

	provider, err := i.NewExecutionProvider(ctx, relayArgs, pluginArgs)
	if err != nil {
		return 0, err
	}
	err = provider.Start(ctx)
	if err != nil {
		return 0, err
	}
	const name = "CCIPExecutionProvider"
	providerRes := net.Resource{Name: name, Closer: provider}

	id, _, err := r.ServeNew(name, func(s *grpc.Server) {
		registerPluginProviderServices(s, provider)
		registerCustomExecutionProviderServices(s, provider, r.BrokerExt)
	}, providerRes)
	if err != nil {
		return 0, err
	}

	return id, err
}

func (r *relayerServer) GetChainStatus(ctx context.Context, request *pb.GetChainStatusRequest) (*pb.GetChainStatusReply, error) {
	chain, err := r.impl.GetChainStatus(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.GetChainStatusReply{Chain: &pb.ChainStatus{
		Id:      chain.ID,
		Enabled: chain.Enabled,
		Config:  chain.Config,
	}}, nil
}

func (r *relayerServer) ListNodeStatuses(ctx context.Context, request *pb.ListNodeStatusesRequest) (*pb.ListNodeStatusesReply, error) {
	nodeConfigs, nextPageToken, total, err := r.impl.ListNodeStatuses(ctx, request.PageSize, request.PageToken)
	if err != nil {
		return nil, err
	}
	var nodes []*pb.NodeStatus
	for _, n := range nodeConfigs {
		nodes = append(nodes, &pb.NodeStatus{
			ChainID: n.ChainID,
			Name:    n.Name,
			Config:  n.Config,
			State:   n.State,
		})
	}
	return &pb.ListNodeStatusesReply{Nodes: nodes, NextPageToken: nextPageToken, Total: int32(total)}, nil
}

func (r *relayerServer) Transact(ctx context.Context, request *pb.TransactionRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, r.impl.Transact(ctx, request.From, request.To, request.Amount.Int(), request.BalanceCheck)
}

// RegisterStandAloneMedianProvider register the servers needed for a median plugin provider,
// this is a workaround to test the Node API on EVM until the EVM relayer is loopifyed.
func RegisterStandAloneMedianProvider(s *grpc.Server, p types.MedianProvider) {
	registerPluginProviderServices(s, p)
	pb.RegisterReportCodecServer(s, &reportCodecServer{impl: p.ReportCodec()})
	pb.RegisterMedianContractServer(s, &medianContractServer{impl: p.MedianContract()})
	pb.RegisterOnchainConfigCodecServer(s, &onchainConfigCodecServer{impl: p.OnchainConfigCodec()})
}

// RegisterStandAlonePluginProvider register the servers needed for a generic plugin provider,
// this is a workaround to test the Node API on EVM until the EVM relayer is loopifyed.
func RegisterStandAlonePluginProvider(s *grpc.Server, p types.PluginProvider) {
	registerPluginProviderServices(s, p)
}
