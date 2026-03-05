package relayer

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
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	aptospb "github.com/smartcontractkit/chainlink-common/pkg/chains/aptos"
	evmpb "github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	solpb "github.com/smartcontractkit/chainlink-common/pkg/chains/solana"
	tonpb "github.com/smartcontractkit/chainlink-common/pkg/chains/ton"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/capability"
	ks "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/keystore"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	ccipocr3pb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/ccipocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/ccip"
	ccipocr3loop "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/ccipocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/median"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/mercury"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ext/ocr3capability"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/ocr2"
	looptypes "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	ccipocr3types "github.com/smartcontractkit/chainlink-common/pkg/types/ccipocr3"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

var _ looptypes.PluginRelayer = (*PluginRelayerClient)(nil)

type PluginRelayerClient struct {
	*goplugin.PluginClient
	*goplugin.ServiceClient

	pluginRelayer pb.PluginRelayerClient
}

func NewPluginRelayerClient(brokerCfg net.BrokerConfig) *PluginRelayerClient {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "PluginRelayerClient")
	pc := goplugin.NewPluginClient(brokerCfg)
	return &PluginRelayerClient{PluginClient: pc, pluginRelayer: pb.NewPluginRelayerClient(pc), ServiceClient: goplugin.NewServiceClient(pc.BrokerExt, pc)}
}

func (p *PluginRelayerClient) NewRelayer(ctx context.Context, config string, keystore, csaKeystore core.Keystore, capabilityRegistry core.CapabilitiesRegistry) (looptypes.Relayer, error) {
	cc := p.NewClientConn("Relayer", func(ctx context.Context) (relayerID uint32, deps net.Resources, err error) {
		var ksRes net.Resource
		ksID, ksRes, err := p.ServeNew("Keystore", func(s *grpc.Server) {
			pb.RegisterKeystoreServer(s, ks.NewServer(keystore))
		})
		if err != nil {
			return 0, nil, fmt.Errorf("Failed to create relayer client: failed to serve keystore: %w", err)
		}
		deps.Add(ksRes)

		var ksCSARes net.Resource
		ksCSAID, ksCSARes, err := p.ServeNew("CSAKeystore", func(s *grpc.Server) {
			pb.RegisterKeystoreServer(s, ks.NewServer(csaKeystore))
		})
		if err != nil {
			return 0, nil, fmt.Errorf("Failed to create relayer client: failed to serve CSA keystore: %w", err)
		}
		deps.Add(ksCSARes)

		capabilityRegistryID, capabilityRegistryResource, err := p.ServeNew("CapabilitiesRegistry", func(s *grpc.Server) {
			pb.RegisterCapabilitiesRegistryServer(s, capability.NewCapabilitiesRegistryServer(p.BrokerExt, capabilityRegistry))
		})
		if err != nil {
			return 0, nil, fmt.Errorf("failed to serve new capability registry: %w", err)
		}
		deps.Add(capabilityRegistryResource)

		reply, err := p.pluginRelayer.NewRelayer(ctx, &pb.NewRelayerRequest{
			Config:               config,
			KeystoreID:           ksID,
			KeystoreCSAID:        ksCSAID,
			CapabilityRegistryID: capabilityRegistryID,
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

	impl looptypes.PluginRelayer
}

func RegisterPluginRelayerServer(server *grpc.Server, broker net.Broker, brokerCfg net.BrokerConfig, impl looptypes.PluginRelayer) error {
	pb.RegisterServiceServer(server, &goplugin.ServiceServer{Srv: impl})
	pb.RegisterPluginRelayerServer(server, newPluginRelayerServer(broker, brokerCfg, impl))
	return nil
}

func newPluginRelayerServer(broker net.Broker, brokerCfg net.BrokerConfig, impl looptypes.PluginRelayer) *pluginRelayerServer {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "RelayerPluginServer")
	return &pluginRelayerServer{BrokerExt: &net.BrokerExt{Broker: broker, BrokerConfig: brokerCfg}, impl: impl}
}

func (p *pluginRelayerServer) NewRelayer(ctx context.Context, request *pb.NewRelayerRequest) (*pb.NewRelayerReply, error) {
	ksConn, err := p.Dial(request.KeystoreID)
	if err != nil {
		return nil, net.ErrConnDial{Name: "Keystore", ID: request.KeystoreID, Err: err}
	}
	ksRes := net.Resource{Closer: ksConn, Name: "Keystore"}

	ksCSAConn, err := p.Dial(request.KeystoreCSAID)
	if err != nil {
		p.CloseAll(ksRes)
		return nil, net.ErrConnDial{Name: "CSAKeystore", ID: request.KeystoreCSAID, Err: err}
	}
	ksCSARes := net.Resource{Closer: ksConn, Name: "CSAKeystore"}

	capRegistryConn, err := p.Dial(request.CapabilityRegistryID)
	if err != nil {
		p.CloseAll(ksRes, ksCSARes)
		return nil, net.ErrConnDial{Name: "CapabilityRegistry", ID: request.CapabilityRegistryID, Err: err}
	}
	crRes := net.Resource{Closer: capRegistryConn, Name: "CapabilityRegistry"}
	capRegistry := capability.NewCapabilitiesRegistryClient(capRegistryConn, p.BrokerExt)

	csaKeystore := ks.NewClient(ksCSAConn)

	// Sets the auth header signing mechanism
	beholder.GetClient().SetSigner(csaKeystore)

	r, err := p.impl.NewRelayer(ctx, request.Config, ks.NewClient(ksConn), csaKeystore, capRegistry)
	if err != nil {
		p.CloseAll(ksRes, ksCSARes, crRes)
		return nil, err
	}
	err = r.Start(ctx)
	if err != nil {
		p.CloseAll(ksRes, ksCSARes, crRes)
		return nil, err
	}

	const name = "Relayer"
	rRes := net.Resource{Closer: r, Name: name}
	id, _, err := p.ServeNew(name, func(s *grpc.Server) {
		pb.RegisterServiceServer(s, &goplugin.ServiceServer{Srv: r})
		pb.RegisterRelayerServer(s, newChainRelayerServer(r, p.BrokerExt))
		if evmService, ok := r.(types.EVMService); ok {
			evmpb.RegisterEVMServer(s, newEVMServer(evmService, p.BrokerExt))
		}
		if tonService, ok := r.(types.TONService); ok {
			tonpb.RegisterTONServer(s, newTONServer(tonService, p.BrokerExt))
		}
		if solService, ok := r.(types.SolanaService); ok {
			solpb.RegisterSolanaServer(s, newSolServer(solService, p.BrokerExt))
		}
		if aptosService, ok := r.(types.AptosService); ok {
			aptospb.RegisterAptosServer(s, newAptosServer(aptosService, p.BrokerExt))
		}
	}, rRes, ksRes, ksCSARes, crRes)
	if err != nil {
		return nil, err
	}

	return &pb.NewRelayerReply{RelayerID: id}, nil
}

// relayerClient adapts a GRPC [pb.RelayerClient] to implement [Relayer].
type relayerClient struct {
	*net.BrokerExt
	*goplugin.ServiceClient

	relayer     pb.RelayerClient
	evmClient   evmpb.EVMClient
	tonClient   tonpb.TONClient
	solClient   solpb.SolanaClient
	aptosClient aptospb.AptosClient
}

func newRelayerClient(b *net.BrokerExt, conn grpc.ClientConnInterface) *relayerClient {
	b = b.WithName("RelayerClient")
	return &relayerClient{
		b, goplugin.NewServiceClient(b, conn),
		pb.NewRelayerClient(conn),
		evmpb.NewEVMClient(conn),
		tonpb.NewTONClient(conn),
		solpb.NewSolanaClient(conn),
		aptospb.NewAptosClient(conn),
	}
}

func (r *relayerClient) NewContractWriter(_ context.Context, contractWriterConfig []byte) (types.ContractWriter, error) {
	cwc := r.NewClientConn("ContractWriter", func(ctx context.Context) (uint32, net.Resources, error) {
		reply, err := r.relayer.NewContractWriter(ctx, &pb.NewContractWriterRequest{ContractWriterConfig: contractWriterConfig})
		if err != nil {
			return 0, nil, err
		}
		return reply.ContractWriterID, nil, nil
	})
	return contractwriter.NewClient(r.WithName("ContractWriterClient"), cwc), nil
}

func (r *relayerClient) NewContractReader(_ context.Context, contractReaderConfig []byte) (types.ContractReader, error) {
	cc := r.NewClientConn("ContractReader", func(ctx context.Context) (uint32, net.Resources, error) {
		reply, err := r.relayer.NewContractReader(ctx, &pb.NewContractReaderRequest{ContractReaderConfig: contractReaderConfig})
		if err != nil {
			return 0, nil, err
		}
		return reply.ContractReaderID, nil, nil
	})

	return contractreader.NewClient(goplugin.NewServiceClient(r.WithName("ContractReaderClient"), cc),
		pb.NewContractReaderClient(cc)), nil
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
				ProviderType:  rargs.ProviderType,
			},
		})
		if err != nil {
			return 0, nil, err
		}
		return reply.ConfigProviderID, nil, nil
	})
	return ocr2.NewConfigProviderClient(r.WithName(rargs.ExternalJobID.String()).WithName("ConfigProviderClient"), cc), nil
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

	return WrapProviderClientConnection(ctx, rargs.ProviderType, cc, r.WithName(rargs.ExternalJobID.String()).WithName("PluginProviderClient"))
}

type PluginProviderClient interface {
	types.PluginProvider
	goplugin.GRPCClientConn
}

func WrapProviderClientConnection(ctx context.Context, providerType string, cc grpc.ClientConnInterface, broker *net.BrokerExt) (PluginProviderClient, error) {
	// TODO: Remove this when we have fully transitioned all relayers to running in LOOPPs.
	// This allows callers to type assert a PluginProvider into a product provider type (eg. MedianProvider)
	// for interoperability with legacy code.
	switch providerType {
	case string(types.Median):
		pc := median.NewProviderClient(broker, cc)
		pc.RmUnimplemented(ctx)
		return pc, nil
	case string(types.GenericPlugin):
		return ocr2.NewPluginProviderClient(broker, cc), nil
	case string(types.OCR3Capability):
		return ocr3capability.NewProviderClient(broker, cc), nil
	case string(types.Mercury):
		return mercury.NewProviderClient(broker, cc), nil
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
		return ccip.NewExecProviderClient(broker, cc), fmt.Errorf("need to fix BCF-3061")
	case string(types.CCIPCommit):
		return ccip.NewCommitProviderClient(broker, cc), fmt.Errorf("need to fix BCF-3061")
	default:
		return nil, fmt.Errorf("provider type not supported: %s", providerType)
	}
}

func (r *relayerClient) NewLLOProvider(ctx context.Context, rargs types.RelayArgs, pargs types.PluginArgs) (types.LLOProvider, error) {
	return nil, fmt.Errorf("llo provider not supported: %w", errors.ErrUnsupported)
}

func (r *relayerClient) NewCCIPProvider(ctx context.Context, cargs types.CCIPProviderArgs) (types.CCIPProvider, error) {
	var ccipProvider *ccipocr3loop.CCIPProviderClient
	cc := r.NewClientConn("CCIPProvider", func(ctx context.Context) (uint32, net.Resources, error) {
		var deps net.Resources

		var extraDataCodecBundleID uint32
		if cargs.ExtraDataCodecBundle != nil {
			edcID, edcRes, err := r.ServeNew("ExtraDataCodecBundle", func(s *grpc.Server) {
				ccipocr3pb.RegisterExtraDataCodecBundleServer(s, ccipocr3loop.NewExtraDataCodecBundleServer(cargs.ExtraDataCodecBundle))
			})
			if err != nil {
				return 0, nil, fmt.Errorf("failed to serve ExtraDataCodecBundle: %w", err)
			}
			deps.Add(edcRes)
			extraDataCodecBundleID = edcID
		}

		persistedSyncs := ccipProvider.GetSyncRequests()
		reply, err := r.relayer.NewCCIPProvider(ctx, &pb.NewCCIPProviderRequest{
			CcipProviderArgs: &pb.CCIPProviderArgs{
				ExternalJobID:          cargs.ExternalJobID[:],
				ContractReaderConfig:   cargs.ContractReaderConfig,
				ChainWriterConfig:      cargs.ChainWriterConfig,
				OffRampAddress:         cargs.OffRampAddress,
				TransmitterAddress:     string(cargs.TransmitterAddress),
				PluginType:             uint32(cargs.PluginType),
				SyncedAddresses:        persistedSyncs,
				ExtraDataCodecBundleID: extraDataCodecBundleID,
			},
		})
		if err != nil {
			return 0, nil, err
		}
		return reply.CcipProviderID, deps, nil
	})

	ccipProvider = ccipocr3loop.NewCCIPProviderClient(r.WithName(cargs.ExternalJobID.String()).WithName("CCIPProviderClient"), cc)
	return ccipProvider, nil
}

func (r *relayerClient) LatestHead(ctx context.Context) (types.Head, error) {
	reply, err := r.relayer.LatestHead(ctx, &pb.LatestHeadRequest{})
	if err != nil {
		return types.Head{}, err
	}

	return types.Head{
		Height:    reply.Head.Height,
		Hash:      reply.Head.Hash,
		Timestamp: reply.Head.Timestamp,
	}, nil
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

func (r *relayerClient) GetChainInfo(ctx context.Context) (types.ChainInfo, error) {
	chainInfoReply, err := r.relayer.GetChainInfo(ctx, &pb.GetChainInfoRequest{})
	if err != nil {
		return types.ChainInfo{}, err
	}

	chainInfo := chainInfoReply.GetChainInfo()
	return types.ChainInfo{
		FamilyName:      chainInfo.GetFamilyName(),
		ChainID:         chainInfo.GetChainId(),
		NetworkName:     chainInfo.GetNetworkName(),
		NetworkNameFull: chainInfo.GetNetworkNameFull(),
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

func (r *relayerClient) Replay(ctx context.Context, fromBlock string, args map[string]any) error {
	argsStruct, err := structpb.NewStruct(args)
	if err != nil {
		return err
	}

	_, err = r.relayer.Replay(ctx, &pb.ReplayRequest{
		FromBlock: fromBlock,
		Args:      argsStruct,
	})
	return err
}

func (r *relayerClient) EVM() (types.EVMService, error) {
	return &EVMClient{
		r.evmClient,
	}, nil
}

func (r *relayerClient) TON() (types.TONService, error) {
	return &TONClient{
		r.tonClient,
	}, nil
}

func (r *relayerClient) Solana() (types.SolanaService, error) {
	return &SolClient{
		r.solClient,
	}, nil
}

func (r *relayerClient) Aptos() (types.AptosService, error) {
	return &AptosClient{
		r.aptosClient,
	}, nil
}

var _ pb.RelayerServer = (*relayerServer)(nil)

// relayerServer exposes [Relayer] as a GRPC [pb.RelayerServer].
type relayerServer struct {
	pb.UnimplementedRelayerServer

	*net.BrokerExt

	impl looptypes.Relayer
}

func newChainRelayerServer(impl looptypes.Relayer, b *net.BrokerExt) *relayerServer {
	return &relayerServer{impl: impl, BrokerExt: b.WithName("ChainRelayerServer")}
}

func (r *relayerServer) NewContractWriter(ctx context.Context, request *pb.NewContractWriterRequest) (*pb.NewContractWriterReply, error) {
	cw, err := r.impl.NewContractWriter(ctx, request.GetContractWriterConfig())
	if err != nil {
		return nil, err
	}

	if err = cw.Start(ctx); err != nil {
		return nil, err
	}

	const name = "ContractWriter"
	id, _, err := r.ServeNew(name, func(s *grpc.Server) {
		contractwriter.RegisterContractWriterService(s, cw)
	}, net.Resource{Closer: cw, Name: name})
	if err != nil {
		return nil, err
	}

	return &pb.NewContractWriterReply{ContractWriterID: id}, nil
}

func (r *relayerServer) NewContractReader(ctx context.Context, request *pb.NewContractReaderRequest) (*pb.NewContractReaderReply, error) {
	cr, err := r.impl.NewContractReader(ctx, request.GetContractReaderConfig())
	if err != nil {
		return nil, err
	}

	if err = cr.Start(ctx); err != nil {
		return nil, err
	}

	const name = "ContractReader"
	id, _, err := r.ServeNew(name, func(s *grpc.Server) {
		contractreader.RegisterContractReaderService(s, cr)
	}, net.Resource{Closer: cr, Name: name})
	if err != nil {
		return nil, err
	}

	return &pb.NewContractReaderReply{ContractReaderID: id}, nil
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
		ProviderType:  request.RelayArgs.ProviderType,
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
		ocr2.RegisterConfigProviderServices(s, cp)
	}, net.Resource{Closer: cp, Name: name})
	if err != nil {
		return nil, err
	}

	return &pb.NewConfigProviderReply{ConfigProviderID: id}, nil
}

func (r *relayerServer) NewPluginProvider(ctx context.Context, request *pb.NewPluginProviderRequest) (*pb.NewPluginProviderReply, error) {
	rargs := request.RelayArgs
	pargs := request.PluginArgs

	exJobID, err := uuid.FromBytes(rargs.ExternalJobID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid bytes for ExternalJobID: %w", err)
	}
	relayArgs := types.RelayArgs{
		ExternalJobID: exJobID,
		JobID:         rargs.JobID,
		ContractID:    rargs.ContractID,
		New:           rargs.New,
		RelayConfig:   rargs.RelayConfig,
		ProviderType:  rargs.ProviderType,
	}
	pluginArgs := types.PluginArgs{
		TransmitterID: pargs.TransmitterID,
		PluginConfig:  pargs.PluginConfig,
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
	case string(types.CCIPCommit):
		id, err := r.newCommitProvider(ctx, relayArgs, pluginArgs)
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
	case string(types.OCR3Capability):
		id, err := r.newOCR3CapabilityProvider(ctx, relayArgs, pluginArgs)
		if err != nil {
			return nil, err
		}
		return &pb.NewPluginProviderReply{PluginProviderID: id}, nil
	}
	return nil, fmt.Errorf("provider type not supported: %s", relayArgs.ProviderType)
}

func (r *relayerServer) newOCR3CapabilityProvider(ctx context.Context, relayArgs types.RelayArgs, pluginArgs types.PluginArgs) (uint32, error) {
	i, ok := r.impl.(looptypes.OCR3CapabilityProvider)
	if !ok {
		return 0, status.Error(codes.Unimplemented, "median not supported")
	}

	provider, err := i.NewOCR3CapabilityProvider(ctx, relayArgs, pluginArgs)
	if err != nil {
		return 0, err
	}
	err = provider.Start(ctx)
	if err != nil {
		return 0, err
	}
	const name = "OCR3CapabilityProvider"
	providerRes := net.Resource{Name: name, Closer: provider}

	id, _, err := r.ServeNew(name, func(s *grpc.Server) {
		ocr3capability.RegisterProviderServices(s, provider)
	}, providerRes)
	if err != nil {
		return 0, err
	}

	return id, err
}

func (r *relayerServer) newMedianProvider(ctx context.Context, relayArgs types.RelayArgs, pluginArgs types.PluginArgs) (uint32, error) {
	i, ok := r.impl.(looptypes.MedianProvider)
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
		median.RegisterProviderServices(s, provider)
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
		ocr2.RegisterPluginProviderServices(s, provider)
	}, providerRes)
	if err != nil {
		return 0, err
	}

	return id, err
}

func (r *relayerServer) newMercuryProvider(ctx context.Context, relayArgs types.RelayArgs, pluginArgs types.PluginArgs) (uint32, error) {
	i, ok := r.impl.(looptypes.MercuryProvider)
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
		ocr2.RegisterPluginProviderServices(s, provider)
		mercury.RegisterProviderServices(s, provider)
	}, providerRes)
	if err != nil {
		return 0, err
	}

	return id, err
}

func (r *relayerServer) newExecProvider(ctx context.Context, relayArgs types.RelayArgs, pluginArgs types.PluginArgs) (uint32, error) {
	i, ok := r.impl.(looptypes.CCIPExecProvider)
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
		ocr2.RegisterPluginProviderServices(s, provider)
		ccip.RegisterExecutionProviderServices(s, provider, r.BrokerExt)
	}, providerRes)
	if err != nil {
		return 0, err
	}

	return id, err
}

func (r *relayerServer) newCommitProvider(ctx context.Context, relayArgs types.RelayArgs, pluginArgs types.PluginArgs) (uint32, error) {
	i, ok := r.impl.(looptypes.CCIPCommitProvider)
	if !ok {
		return 0, status.Error(codes.Unimplemented, fmt.Sprintf("ccip execution not supported by %T", r.impl))
	}

	provider, err := i.NewCommitProvider(ctx, relayArgs, pluginArgs)
	if err != nil {
		return 0, err
	}
	err = provider.Start(ctx)
	if err != nil {
		return 0, err
	}
	const name = "CCIPCommitProvider"
	providerRes := net.Resource{Name: name, Closer: provider}

	id, _, err := r.ServeNew(name, func(s *grpc.Server) {
		ocr2.RegisterPluginProviderServices(s, provider)
		ccip.RegisterCommitProviderServices(s, provider, r.BrokerExt)
	}, providerRes)
	if err != nil {
		return 0, err
	}

	return id, err
}

func (r *relayerServer) NewCCIPProvider(ctx context.Context, request *pb.NewCCIPProviderRequest) (*pb.NewCCIPProviderReply, error) {
	rargs := request.CcipProviderArgs

	exJobID, err := uuid.FromBytes(rargs.ExternalJobID)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid bytes for ExternalJobID: %w", err)
	}

	var extraDataCodecBundle ccipocr3types.ExtraDataCodecBundle
	var extraDataCodecRes net.Resource

	// If ExtraDataCodecBundleID is provided, dial the service
	if rargs.ExtraDataCodecBundleID != 0 {
		extraDataCodecConn, err := r.Dial(rargs.ExtraDataCodecBundleID)
		if err != nil {
			return nil, net.ErrConnDial{Name: "ExtraDataCodecBundle", ID: rargs.ExtraDataCodecBundleID, Err: err}
		}
		extraDataCodecRes = net.Resource{Closer: extraDataCodecConn, Name: "ExtraDataCodecBundle"}
		extraDataCodecBundle = ccipocr3loop.NewExtraDataCodecBundleClient(r.BrokerExt, extraDataCodecConn)
	}

	ccipProviderArgs := types.CCIPProviderArgs{
		ExternalJobID:        exJobID,
		ContractReaderConfig: rargs.ContractReaderConfig,
		ChainWriterConfig:    rargs.ChainWriterConfig,
		OffRampAddress:       rargs.OffRampAddress,
		PluginType:           ccipocr3types.PluginType(rargs.PluginType),
		TransmitterAddress:   ccipocr3types.UnknownEncodedAddress(rargs.TransmitterAddress),
		ExtraDataCodecBundle: extraDataCodecBundle,
	}

	provider, err := r.impl.NewCCIPProvider(ctx, ccipProviderArgs)
	if err != nil {
		if extraDataCodecRes.Closer != nil {
			extraDataCodecRes.Close()
		}
		return nil, err
	}

	// Sync persisted sync requests after provider has initted accessor
	for contractName, addressBytes := range rargs.SyncedAddresses {
		err = provider.ChainAccessor().Sync(ctx, contractName, addressBytes)
		if err != nil {
			if extraDataCodecRes.Closer != nil {
				extraDataCodecRes.Close()
			}
			return nil, err
		}
	}

	const name = "CCIPProvider"
	resources := []net.Resource{{Closer: provider, Name: name}}
	if extraDataCodecRes.Closer != nil {
		resources = append(resources, extraDataCodecRes)
	}

	id, _, err := r.ServeNew(name, func(s *grpc.Server) {
		ccipocr3loop.RegisterProviderServices(s, provider)
	}, resources...)
	if err != nil {
		return nil, err
	}

	return &pb.NewCCIPProviderReply{CcipProviderID: id}, nil
}

func (r *relayerServer) LatestHead(ctx context.Context, _ *pb.LatestHeadRequest) (*pb.LatestHeadReply, error) {
	head, err := r.impl.LatestHead(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.LatestHeadReply{
		Head: &pb.Head{
			Height:    head.Height,
			Hash:      head.Hash,
			Timestamp: head.Timestamp,
		},
	}, nil
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

func (r *relayerServer) GetChainInfo(ctx context.Context, _ *pb.GetChainInfoRequest) (*pb.GetChainInfoReply, error) {
	chainInfo, err := r.impl.GetChainInfo(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GetChainInfoReply{
		ChainInfo: &pb.ChainInfo{
			FamilyName:      chainInfo.FamilyName,
			ChainId:         chainInfo.ChainID,
			NetworkName:     chainInfo.NetworkName,
			NetworkNameFull: chainInfo.NetworkNameFull,
		},
	}, nil
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

func (r *relayerServer) Replay(ctx context.Context, request *pb.ReplayRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, r.impl.Replay(ctx, request.FromBlock, request.Args.AsMap())
}

// RegisterStandAloneMedianProvider register the servers needed for a median plugin provider,
// this is a workaround to test the Node API on EVM until the EVM relayer is loopifyed.
func RegisterStandAloneMedianProvider(s *grpc.Server, p types.MedianProvider) {
	median.RegisterProviderServices(s, p)
}

// RegisterStandAlonePluginProvider register the servers needed for a generic plugin provider,
// this is a workaround to test the Node API on EVM until the EVM relayer is loopifyed.
func RegisterStandAlonePluginProvider(s *grpc.Server, p types.PluginProvider) {
	ocr2.RegisterPluginProviderServices(s, p)
}

// RegisterStandAloneOCR3CapabilityProvider register the servers needed for a generic plugin provider,
// this is a workaround to test the Node API on EVM until the EVM relayer is loopifyed.
func RegisterStandAloneOCR3CapabilityProvider(s *grpc.Server, p types.OCR3CapabilityProvider) {
	ocr3capability.RegisterProviderServices(s, p)
}
