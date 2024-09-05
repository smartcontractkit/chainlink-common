package relayerset

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/chainreader"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayer/pluginprovider/chainwriter"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayerset/inprocessprovider"
	"github.com/smartcontractkit/chainlink-common/pkg/types"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

type Server struct {
	log logger.Logger

	relayerset.UnimplementedRelayerSetServer
	impl   core.RelayerSet
	broker *net.BrokerExt

	serverResources net.Resources

	Name string
}

func NewRelayerSetServer(log logger.Logger, underlying core.RelayerSet, broker *net.BrokerExt) (*Server, net.Resource) {
	pluginProviderServers := make(net.Resources, 0)
	server := &Server{log: log, impl: underlying, broker: broker, serverResources: pluginProviderServers}

	return server, net.Resource{
		Name:   "PluginProviderServers",
		Closer: server,
	}
}

func (s *Server) Close() error {
	for _, pluginProviderServer := range s.serverResources {
		if err := pluginProviderServer.Close(); err != nil {
			return fmt.Errorf("error closing plugin provider server: %w", err)
		}
	}

	return nil
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
	var relayIDs []types.RelayID
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

func (s *Server) NewPluginProvider(ctx context.Context, req *relayerset.NewPluginProviderRequest) (*relayerset.NewPluginProviderResponse, error) {
	relayer, err := s.getRelayer(ctx, req.RelayerId)
	if err != nil {
		return nil, err
	}

	relayArgs := core.RelayArgs{
		ContractID:   req.RelayArgs.ContractID,
		RelayConfig:  req.RelayArgs.RelayConfig,
		ProviderType: req.RelayArgs.ProviderType,
	}

	// TODO - the mercury credentials should be set as part of the relay config and not as a separate field
	if req.RelayArgs.MercuryCredentials != nil {
		relayArgs.MercuryCredentials = &types.MercuryCredentials{
			LegacyURL: req.RelayArgs.MercuryCredentials.LegacyUrl,
			URL:       req.RelayArgs.MercuryCredentials.Url,
			Username:  req.RelayArgs.MercuryCredentials.Username,
			Password:  req.RelayArgs.MercuryCredentials.Password,
		}
	}

	pluginArgs := core.PluginArgs{
		TransmitterID: req.PluginArgs.TransmitterID,
		PluginConfig:  req.PluginArgs.PluginConfig,
	}

	pluginProvider, err := relayer.NewPluginProvider(ctx, relayArgs, pluginArgs)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error creating plugin provider: %v", err)
	}

	var providerClientConn grpc.ClientConnInterface
	providerConn, ok := pluginProvider.(connProvider)
	if !ok {
		providerClientConn, err = s.getProviderConnection(pluginProvider, relayArgs.ProviderType)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "error getting provider connection: %v", err)
		}
	} else {
		providerClientConn = providerConn.ClientConn()
	}

	providerID, providerRes, err := s.broker.Serve("PluginProvider", proxy.NewProxy(providerClientConn))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error serving plugin provider: %v", err)
	}
	s.serverResources.Add(providerRes)

	return &relayerset.NewPluginProviderResponse{PluginProviderId: providerID}, nil
}

// RelayerSet is supposed to serve relayers, which then hold a ContractReader and ChainWriter. Serving NewContractReader
// and NewChainWriter from RelayerSet is a way to save us from instantiating an extra server for the Relayer. Without
// this approach, the calls we would make normally are
//   - RelayerSet.Get -> Relayer
//   - Relayer.NewContractReader -> ContractReader
//
// We could translate this to the GRPC world by having each call to RelayerSet.Get wrap the returned relayer in a server
// and register that to the GRPC server. However this is actually pretty inefficient since a relayer object on its own
// is not useful. Users will always want to use the relayer to instantiate a chainreader or chainwriter. So we can avoid
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

	contractReader, err := relayer.NewContractReader(ctx, req.ContractReaderConfig)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error creating contract reader: %v", err)
	}

	// Start ContractReader service
	if err = contractReader.Start(ctx); err != nil {
		return nil, err
	}

	// Start gRPC service for the ContractReader service above
	const name = "ContractReaderInRelayerSet"
	id, _, err := s.broker.ServeNew(name, func(s *grpc.Server) {
		chainreader.RegisterContractReaderService(s, contractReader)
	}, net.Resource{Closer: contractReader, Name: name})
	if err != nil {
		return nil, err
	}

	return &relayerset.NewContractReaderResponse{ContractReaderId: id}, nil
}

// RelayerSet is supposed to serve relayers, which then hold a ContractReader and ChainWriter. Serving NewChainWriter
// and NewChainWriter from RelayerSet is a way to save us from instantiating an extra server for the Relayer. Without
// this approach, the calls we would make normally are
//   - RelayerSet.Get -> Relayer
//   - Relayer.NewChainWriter -> ChainWriter
//
// We could translate this to the GRPC world by having each call to RelayerSet.Get wrap the returned relayer in a server
// and register that to the GRPC server. However this is actually pretty inefficient since a relayer object on its own
// is not useful. Users will always want to use the relayer to instantiate a chainreader or chainwriter. So we can avoid
// the intermediate server for the relayer by just storing a reference to the relayerSet client and the relayer we want
// to fetch. I.e. the calls described above instead would become:
//   - RelayerSet.Get -> (RelayerSetClient, RelayerID). Effectively this call just acts as check that Relayer exists
//
// RelayerClient.NewChainWriter -> This is a call to RelayerSet.NewChainWriter with (relayerID, []contractReaderConfig);
// The implementation will then fetch the relayer and call NewChainWriter on it
func (s *Server) NewChainWriter(ctx context.Context, req *relayerset.NewChainWriterRequest) (*relayerset.NewChainWriterResponse, error) {
	relayer, err := s.getRelayer(ctx, req.RelayerId)
	if err != nil {
		return nil, err
	}

	chainWriter, err := relayer.NewChainWriter(ctx, req.ChainWriterConfig)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error creating contract reader: %v", err)
	}

	// Start ChainWriter service
	if err = chainWriter.Start(ctx); err != nil {
		return nil, err
	}

	// Start gRPC service for the ChainWriter service above
	const name = "ChainWriterInRelayerSet"
	id, _, err := s.broker.ServeNew(name, func(s *grpc.Server) {
		chainwriter.RegisterChainWriterService(s, chainWriter)
	}, net.Resource{Closer: chainWriter, Name: name})
	if err != nil {
		return nil, err
	}

	return &relayerset.NewChainWriterResponse{ChainWriterId: id}, nil
}

// getProviderConnection wraps a non-LOOPP provider in an in process provider server.  This can be removed once all providers are LOOPP providers.
// For completeness the original comment from the equivalent code in core is included here:
//
// We chose to deal with the difference between a LOOP provider and an embedded provider here rather than
// in NewServerAdapter because this has a smaller blast radius, as the scope of this workaround is to
// enable the medianpoc for EVM and not touch the other providers.
// TODO: remove this workaround when the EVM relayer is running inside of an LOOPP
func (s *Server) getProviderConnection(pluginProvider types.PluginProvider, providerType string) (grpc.ClientConnInterface, error) {
	s.log.Info("wrapping provider %s in an in process provider server as it is not a LOOPP provider, ", pluginProvider.Name())

	ps, err := inprocessprovider.NewProviderServer(pluginProvider, types.OCR2PluginType(providerType), s.log)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap provider %s in in process provider server: %w", pluginProvider.Name(), err)
	}

	s.serverResources.Add(net.Resource{
		Closer: ps,
		Name:   fmt.Sprintf("InProcessProviderServer-%s", pluginProvider.Name()),
	})

	providerClientConn, err := ps.GetConn()
	if err != nil {
		return nil, fmt.Errorf("failed to get connection to in process provider server: %w", err)
	}
	return providerClientConn, nil
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

func (s *Server) getRelayer(ctx context.Context, relayerID *relayerset.RelayerId) (core.Relayer, error) {
	relayer, err := s.impl.Get(ctx, types.RelayID{ChainID: relayerID.ChainId, Network: relayerID.Network})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "error getting relayer: %v", err)
	}

	return relayer, nil
}
