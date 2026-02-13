package relayerset

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	aptospb "github.com/smartcontractkit/chainlink-common/pkg/chains/aptos"
	evmpb "github.com/smartcontractkit/chainlink-common/pkg/chains/evm"
	solpb "github.com/smartcontractkit/chainlink-common/pkg/chains/solana"
	tonpb "github.com/smartcontractkit/chainlink-common/pkg/chains/ton"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractreader"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/contractwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

type readerAndServer struct {
	reader types.ContractReader
	server pb.ContractReaderServer
}

type Server struct {
	log logger.Logger

	relayerset.UnimplementedRelayerSetServer

	impl   core.RelayerSet
	broker *net.BrokerExt

	sol            *solServer
	ton            *tonServer
	evm            *evmServer
	aptos          *aptosServer
	contractReader *readerServer

	serverResources net.Resources

	readers map[string]*readerAndServer

	Name string

	readersMux sync.Mutex
}

var _ relayerset.RelayerSetServer = (*Server)(nil)

func NewRelayerSetServer(log logger.Logger, underlying core.RelayerSet, broker *net.BrokerExt) (*Server, net.Resource) {
	pluginProviderServers := make(net.Resources, 0)
	server := &Server{log: log, impl: underlying, broker: broker, serverResources: pluginProviderServers,
		readers: map[string]*readerAndServer{}}
	server.sol = &solServer{parent: server}
	server.ton = &tonServer{parent: server}
	server.evm = &evmServer{parent: server}
	server.aptos = &aptosServer{parent: server}
	server.contractReader = &readerServer{parent: server}

	return server, net.Resource{
		Name:   "PluginProviderServers",
		Closer: server,
	}
}

func (s *Server) SolanaServer() solpb.SolanaServer              { return s.sol }
func (s *Server) TONServer() tonpb.TONServer                    { return s.ton }
func (s *Server) EVMServer() evmpb.EVMServer                    { return s.evm }
func (s *Server) AptosServer() aptospb.AptosServer              { return s.aptos }
func (s *Server) ContractReaderServer() pb.ContractReaderServer { return s.contractReader }

func (s *Server) Close() error {
	for _, pluginProviderServer := range s.serverResources {
		if err := pluginProviderServer.Close(); err != nil {
			return fmt.Errorf("error closing plugin provider server: %w", err)
		}
	}

	return nil
}

func (s *Server) getReader(ctx context.Context) (*readerAndServer, error) {
	s.readersMux.Lock()
	defer s.readersMux.Unlock()

	id, err := readContextValue(ctx, metadataContractReader)
	if err != nil {
		return nil, err
	}

	reader, ok := s.readers[id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "contract reader not found for id %s", id)
	}

	return reader, nil
}

func (s *Server) addReader(id string, reader *readerAndServer) {
	s.readersMux.Lock()
	defer s.readersMux.Unlock()

	s.readers[id] = reader
}

func (s *Server) removeReader(id string) {
	s.readersMux.Lock()
	defer s.readersMux.Unlock()

	delete(s.readers, id)
}

func (s *Server) Get(ctx context.Context, req *relayerset.GetRelayerRequest) (*relayerset.GetRelayerResponse, error) {
	id := types.RelayID{ChainID: req.Id.ChainId, Network: req.Id.Network}

	relayers, err := s.impl.List(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error getting all relayers: %v", err)
	}

	if _, ok := relayers[id]; ok {
		return &relayerset.GetRelayerResponse{Id: req.Id}, nil
	}

	return nil, status.Errorf(codes.NotFound, "relayer not found for id %s", id)
}

func (s *Server) List(ctx context.Context, req *relayerset.ListAllRelayersRequest) (*relayerset.ListAllRelayersResponse, error) {
	relayIDs := make([]types.RelayID, 0, len(req.Ids))
	for _, id := range req.Ids {
		relayIDs = append(relayIDs, types.RelayID{ChainID: id.ChainId, Network: id.Network})
	}

	relayers, err := s.impl.List(ctx, relayIDs...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error getting all relayers: %v", err)
	}

	ids := make([]*relayerset.RelayerId, len(relayers))

	for id := range relayers {
		ids = append(ids, &relayerset.RelayerId{ChainId: id.ChainID, Network: id.Network})
	}

	return &relayerset.ListAllRelayersResponse{Ids: ids}, nil
}

// RelayerSet is supposed to serve relayers, which then hold a ContractReader and ContractWriter. Serving NewContractReader
// and NewContractWriter from RelayerSet is a way to save us from instantiating an extra server for the Relayer. Without
// this approach, the calls we would make normally are
//   - RelayerSet.Get -> Relayer
//   - Relayer.NewContractReader -> ContractReader
//
// We could translate this to the GRPC world by having each call to RelayerSet.Get wrap the returned relayer in a server
// and register that to the GRPC server. However this is actually pretty inefficient since a relayer object on its own
// is not useful. Users will always want to use the relayer to instantiate a contractreader or contractwriter. So we can avoid
// the intermediate server for the relayer by just storing a reference to the relayerSet client and the relayer we want
// to fetch. I.e. the calls described above instead would become:
//   - RelayerSet.Get -> (RelayerSetClient, RelayerID). Effectively this call just acts as check that Relayer exists
//
// RelayerClient.NewContractReader -> This is a call to RelayerSet.NewContractReader with (relayerID, []contractReaderConfig);
// The implementation will then fetch the relayer and call NewContractReader on it
func (s *Server) NewContractReader(ctx context.Context, req *relayerset.NewContractReaderRequest) (*relayerset.NewContractReaderResponse, error) {
	relayer, err := s.getRelayer(ctx, req.RelayerId)
	if err != nil {
		return nil, err
	}

	reader, err := relayer.NewContractReader(ctx, req.ContractReaderConfig)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error creating contract reader: %v", err)
	}

	readerID := uuid.New().String()
	server := contractreader.NewServer(reader)

	s.addReader(readerID, &readerAndServer{
		reader: reader,
		server: server,
	})

	return &relayerset.NewContractReaderResponse{ContractReaderId: readerID}, nil
}

// RelayerSet is supposed to serve relayers, which then hold a ContractReader and ContractWriter. Serving NewContractWriter
// and NewContractWriter from RelayerSet is a way to save us from instantiating an extra server for the Relayer. Without
// this approach, the calls we would make normally are
//   - RelayerSet.Get -> Relayer
//   - Relayer.NewContractWriter -> ContractWriter
//
// We could translate this to the GRPC world by having each call to RelayerSet.Get wrap the returned relayer in a server
// and register that to the GRPC server. However this is actually pretty inefficient since a relayer object on its own
// is not useful. Users will always want to use the relayer to instantiate a contractreader or contractwriter. So we can avoid
// the intermediate server for the relayer by just storing a reference to the relayerSet client and the relayer we want
// to fetch. I.e. the calls described above instead would become:
//   - RelayerSet.Get -> (RelayerSetClient, RelayerID). Effectively this call just acts as check that Relayer exists
//
// RelayerClient.NewContractWriter -> This is a call to RelayerSet.NewContractWriter with (relayerID, []contractWriterConfig);
// The implementation will then fetch the relayer and call NewContractWriter on it
func (s *Server) NewContractWriter(ctx context.Context, req *relayerset.NewContractWriterRequest) (*relayerset.NewContractWriterResponse, error) {
	relayer, err := s.getRelayer(ctx, req.RelayerId)
	if err != nil {
		return nil, err
	}

	contractWriter, err := relayer.NewContractWriter(ctx, req.ContractWriterConfig)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error creating contract reader: %v", err)
	}

	// Start ContractWriter service
	if err = contractWriter.Start(ctx); err != nil {
		return nil, err
	}

	// Start gRPC service for the ContractWriter service above
	const name = "ContractWriterInRelayerSet"
	id, _, err := s.broker.ServeNew(name, func(s *grpc.Server) {
		contractwriter.RegisterContractWriterService(s, contractWriter)
	}, net.Resource{Closer: contractWriter, Name: name})
	if err != nil {
		return nil, err
	}

	return &relayerset.NewContractWriterResponse{ContractWriterId: id}, nil
}

func (s *Server) StartRelayer(ctx context.Context, relayID *relayerset.RelayerId) (*emptypb.Empty, error) {
	relayer, err := s.getRelayer(ctx, relayID)
	if err != nil {
		return nil, err
	}

	if err := relayer.Start(ctx); err != nil {
		return nil, status.Errorf(codes.Internal, "error starting relayer: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) CloseRelayer(ctx context.Context, relayID *relayerset.RelayerId) (*emptypb.Empty, error) {
	relayer, err := s.getRelayer(ctx, relayID)
	if err != nil {
		return nil, err
	}

	if err = relayer.Close(); err != nil {
		return nil, status.Errorf(codes.Internal, "error starting relayer: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) RelayerReady(ctx context.Context, relayID *relayerset.RelayerId) (*emptypb.Empty, error) {
	relayer, err := s.getRelayer(ctx, relayID)
	if err != nil {
		return nil, err
	}

	if err := relayer.Ready(); err != nil {
		return nil, status.Errorf(codes.Internal, "error getting relayer ready: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *Server) RelayerHealthReport(ctx context.Context, relayID *relayerset.RelayerId) (*relayerset.RelayerHealthReportResponse, error) {
	relayer, err := s.getRelayer(ctx, relayID)
	if err != nil {
		return nil, err
	}

	result := map[string]string{}
	healthReport := relayer.HealthReport()
	for k, v := range healthReport {
		result[k] = v.Error()
	}

	return &relayerset.RelayerHealthReportResponse{Report: result}, nil
}

func (s *Server) RelayerName(ctx context.Context, relayID *relayerset.RelayerId) (*relayerset.RelayerNameResponse, error) {
	relayer, err := s.getRelayer(ctx, relayID)
	if err != nil {
		return nil, err
	}

	return &relayerset.RelayerNameResponse{Name: relayer.Name()}, nil
}

func (s *Server) RelayerGetChainInfo(ctx context.Context, req *relayerset.GetChainInfoRequest) (*pb.GetChainInfoReply, error) {
	relayer, err := s.getRelayer(ctx, req.RelayerId)
	if err != nil {
		return nil, err
	}

	chainInfo, err := relayer.GetChainInfo(ctx)
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

func (s *Server) RelayerLatestHead(ctx context.Context, req *relayerset.LatestHeadRequest) (*relayerset.LatestHeadResponse, error) {
	relayer, err := s.getRelayer(ctx, req.RelayerId)
	if err != nil {
		return nil, err
	}

	latestHead, err := relayer.LatestHead(ctx)
	if err != nil {
		return nil, err
	}
	return &relayerset.LatestHeadResponse{
		Height:    latestHead.Height,
		Hash:      latestHead.Hash,
		Timestamp: latestHead.Timestamp,
	}, nil
}

func (s *Server) getRelayer(ctx context.Context, relayerID *relayerset.RelayerId) (core.Relayer, error) {
	relayer, err := s.impl.Get(ctx, types.RelayID{ChainID: relayerID.ChainId, Network: relayerID.Network})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "error getting relayer: %v", err)
	}

	return relayer, nil
}

func readContextValue(ctx context.Context, key string) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		contractReaderIds := md.Get(key)
		if len(contractReaderIds) == 1 {
			return contractReaderIds[0], nil
		}
		return "", fmt.Errorf("num values is not 1 but %d", len(contractReaderIds))
	}
	return "", errors.New("could not read ctx metadata")
}
