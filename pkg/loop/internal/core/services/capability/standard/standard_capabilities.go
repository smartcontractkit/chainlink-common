package capability

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/capability"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/errorlog"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/keyvalue"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/oraclefactory"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/pipeline"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/telemetry"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	oraclefactorypb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/oraclefactory"
	relayersetpb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayerset"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

type StandardCapabilities interface {
	services.Service
	Initialise(ctx context.Context, config string, telemetryService core.TelemetryService, store core.KeyValueStore,
		capabilityRegistry core.CapabilitiesRegistry, errorLog core.ErrorLog,
		pipelineRunner core.PipelineRunnerService, relayerSet core.RelayerSet, oracleFactory core.OracleFactory) error
	Infos(ctx context.Context) ([]capabilities.CapabilityInfo, error)
}

type StandardCapabilitiesClient struct {
	*goplugin.PluginClient
	capabilitiespb.StandardCapabilitiesClient
	*goplugin.ServiceClient
	*net.BrokerExt

	resources []net.Resource
}

var _ StandardCapabilities = (*StandardCapabilitiesClient)(nil)

func NewStandardCapabilitiesClient(brokerCfg net.BrokerConfig) *StandardCapabilitiesClient {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "StandardCapabilitiesClient")
	pc := goplugin.NewPluginClient(brokerCfg)
	return &StandardCapabilitiesClient{
		PluginClient:               pc,
		ServiceClient:              goplugin.NewServiceClient(pc.BrokerExt, pc),
		StandardCapabilitiesClient: capabilitiespb.NewStandardCapabilitiesClient(pc),
		BrokerExt:                  pc.BrokerExt,
	}
}

func (c *StandardCapabilitiesClient) Initialise(ctx context.Context, config string, telemetryService core.TelemetryService,
	keyValueStore core.KeyValueStore, capabilitiesRegistry core.CapabilitiesRegistry, errorLog core.ErrorLog,
	pipelineRunner core.PipelineRunnerService, relayerSet core.RelayerSet, oracleFactory core.OracleFactory) error {
	telemetryID, telemetryRes, err := c.ServeNew("Telemetry", func(s *grpc.Server) {
		pb.RegisterTelemetryServer(s, telemetry.NewTelemetryServer(telemetryService))
	})

	if err != nil {
		return fmt.Errorf("failed to serve new telemetry: %w", err)
	}
	var resources []net.Resource
	resources = append(resources, telemetryRes)

	keyValueStoreID, keyValueStoreRes, err := c.ServeNew("KeyValueStore", func(s *grpc.Server) {
		pb.RegisterKeyValueStoreServer(s, keyvalue.NewServer(keyValueStore))
	})
	if err != nil {
		c.CloseAll(resources...)
		return fmt.Errorf("failed to serve new key value store: %w", err)
	}
	resources = append(resources, keyValueStoreRes)

	capabilitiesRegistryID, capabilityRegistryResource, err := c.ServeNew("CapabilitiesRegistry", func(s *grpc.Server) {
		pb.RegisterCapabilitiesRegistryServer(s, capability.NewCapabilitiesRegistryServer(c.BrokerExt, capabilitiesRegistry))
	})
	if err != nil {
		c.CloseAll(resources...)
		return fmt.Errorf("failed to serve new key value store: %w", err)
	}
	resources = append(resources, capabilityRegistryResource)

	errorLogID, errorLogRes, err := c.ServeNew("ErrorLog", func(s *grpc.Server) {
		pb.RegisterErrorLogServer(s, errorlog.NewServer(errorLog))
	})
	if err != nil {
		c.CloseAll(resources...)
		return fmt.Errorf("failed to serve error log: %w", err)
	}
	resources = append(resources, errorLogRes)

	pipelineRunnerID, pipelineRunnerRes, err := c.ServeNew("PipelineRunner", func(s *grpc.Server) {
		pb.RegisterPipelineRunnerServiceServer(s, pipeline.NewRunnerServer(pipelineRunner))
	})
	if err != nil {
		c.CloseAll(resources...)
		return fmt.Errorf("failed to serve pipeline runner: %w", err)
	}
	resources = append(resources, pipelineRunnerRes)

	relayerSetServer, relayerSetServerRes := relayerset.NewRelayerSetServer(c.Logger, relayerSet, c.BrokerExt)
	resources = append(resources, relayerSetServerRes)

	relayerSetID, relayerSetRes, err := c.ServeNew("RelayerSet", func(s *grpc.Server) {
		relayersetpb.RegisterRelayerSetServer(s, relayerSetServer)
	})
	if err != nil {
		c.CloseAll(resources...)
		return fmt.Errorf("failed to serve relayer set: %w", err)
	}
	resources = append(resources, relayerSetRes)

	oracleFactoryServer, oracleFactoryServerRes := oraclefactory.NewServer(c.Logger, oracleFactory, c.BrokerExt)
	resources = append(resources, oracleFactoryServerRes)

	oracleFactoryID, oracleFactoryRes, err := c.ServeNew("OracleFactory", func(s *grpc.Server) {
		oraclefactorypb.RegisterOracleFactoryServer(s, oracleFactoryServer)
	})
	if err != nil {
		c.CloseAll(resources...)
		return fmt.Errorf("failed to serve oracle factory: %w", err)
	}
	resources = append(resources, oracleFactoryRes)

	_, err = c.StandardCapabilitiesClient.Initialise(ctx, &capabilitiespb.InitialiseRequest{
		Config:           config,
		ErrorLogId:       errorLogID,
		PipelineRunnerId: pipelineRunnerID,
		TelemetryId:      telemetryID,
		CapRegistryId:    capabilitiesRegistryID,
		KeyValueStoreId:  keyValueStoreID,
		RelayerSetId:     relayerSetID,
		OracleFactoryId:  oracleFactoryID,
	})

	if err != nil {
		c.CloseAll(resources...)
		return fmt.Errorf("failed to initialise standard capability: %w", err)
	}

	c.resources = resources

	return nil
}

func (c *StandardCapabilitiesClient) Infos(ctx context.Context) ([]capabilities.CapabilityInfo, error) {
	infosResponse, err := c.StandardCapabilitiesClient.Infos(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to get capability infos: %w", err)
	}

	var infos []capabilities.CapabilityInfo
	for _, infoResponse := range infosResponse.Infos {
		info, err := capability.InfoReplyToInfo(infoResponse)
		if err != nil {
			return nil, fmt.Errorf("failed to convert capability info: %w", err)
		}

		infos = append(infos, info)
	}

	return infos, nil
}

func (c *StandardCapabilitiesClient) Close() error {
	c.CloseAll(c.resources...)
	return c.ServiceClient.Close()
}

type standardCapabilitiesServer struct {
	capabilitiespb.UnimplementedStandardCapabilitiesServer
	*net.BrokerExt
	impl StandardCapabilities

	resources []net.Resource
}

func newStandardCapabilitiesServer(brokerExt *net.BrokerExt, impl StandardCapabilities) *standardCapabilitiesServer {
	return &standardCapabilitiesServer{
		impl:      impl,
		BrokerExt: brokerExt,
	}
}

var _ capabilitiespb.StandardCapabilitiesServer = (*standardCapabilitiesServer)(nil)

func RegisterStandardCapabilitiesServer(server *grpc.Server, broker net.Broker, brokerCfg net.BrokerConfig, impl StandardCapabilities) error {
	bext := &net.BrokerExt{
		BrokerConfig: brokerCfg,
		Broker:       broker,
	}

	capabilityServer := newStandardCapabilitiesServer(bext, impl)
	capabilitiespb.RegisterStandardCapabilitiesServer(server, capabilityServer)
	pb.RegisterServiceServer(server, &goplugin.ServiceServer{Srv: &resourceClosingServer{
		StandardCapabilities: impl,
		server:               capabilityServer,
	}})
	return nil
}

func (s *standardCapabilitiesServer) Initialise(ctx context.Context, request *capabilitiespb.InitialiseRequest) (*emptypb.Empty, error) {
	telemetryConn, err := s.Dial(request.TelemetryId)
	if err != nil {
		return nil, net.ErrConnDial{Name: "Telemetry", ID: request.TelemetryId, Err: err}
	}

	var resources []net.Resource
	resources = append(resources, net.Resource{Closer: telemetryConn, Name: "TelemetryConn"})

	telemetry := telemetry.NewTelemetryServiceClient(telemetryConn)

	keyValueStoreConn, err := s.Dial(request.KeyValueStoreId)
	if err != nil {
		s.CloseAll(resources...)
		return nil, net.ErrConnDial{Name: "KeyValueStore", ID: request.KeyValueStoreId, Err: err}
	}
	resources = append(resources, net.Resource{Closer: keyValueStoreConn, Name: "KeyValueStoreConn"})
	keyValueStore := keyvalue.NewClient(keyValueStoreConn)

	capabilitiesRegistryConn, err := s.Dial(request.CapRegistryId)
	if err != nil {
		s.CloseAll(resources...)
		return nil, net.ErrConnDial{Name: "CapabilitiesRegistry", ID: request.CapRegistryId, Err: err}
	}
	resources = append(resources, net.Resource{Closer: capabilitiesRegistryConn, Name: "CapabilitiesRegistryConn"})
	capabilitiesRegistry := capability.NewCapabilitiesRegistryClient(capabilitiesRegistryConn, s.BrokerExt)

	errorLogConn, err := s.Dial(request.ErrorLogId)
	if err != nil {
		s.CloseAll(resources...)
		return nil, net.ErrConnDial{Name: "ErrorLog", ID: request.ErrorLogId, Err: err}
	}
	resources = append(resources, net.Resource{Closer: errorLogConn, Name: "ErrorLog"})
	errorLog := errorlog.NewClient(errorLogConn)

	pipelineRunnerConn, err := s.Dial(request.PipelineRunnerId)
	if err != nil {
		s.CloseAll(resources...)
		return nil, net.ErrConnDial{Name: "PipelineRunner", ID: request.PipelineRunnerId, Err: err}
	}
	resources = append(resources, net.Resource{Closer: pipelineRunnerConn, Name: "PipelineRunner"})
	pipelineRunner := pipeline.NewRunnerClient(pipelineRunnerConn)

	relayersetConn, err := s.Dial(request.RelayerSetId)
	if err != nil {
		s.CloseAll(resources...)
		return nil, net.ErrConnDial{Name: "RelayerSet", ID: request.RelayerSetId, Err: err}
	}
	resources = append(resources, net.Resource{Closer: relayersetConn, Name: "RelayerSet"})
	relayerSet := relayerset.NewRelayerSetClient(s.Logger, s.BrokerExt, relayersetConn)

	oracleFactoryConn, err := s.Dial(request.OracleFactoryId)
	if err != nil {
		s.CloseAll(resources...)
		return nil, net.ErrConnDial{Name: "OracleFactory", ID: request.OracleFactoryId, Err: err}
	}
	resources = append(resources, net.Resource{Closer: oracleFactoryConn, Name: "OracleFactory"})
	oracleFactory := oraclefactory.NewClient(s.Logger, s.BrokerExt, oracleFactoryConn)

	if err = s.impl.Initialise(ctx, request.Config, telemetry, keyValueStore, capabilitiesRegistry, errorLog, pipelineRunner, relayerSet, oracleFactory); err != nil {
		s.CloseAll(resources...)
		return nil, fmt.Errorf("failed to initialise standard capability: %w", err)
	}

	s.resources = resources

	return &emptypb.Empty{}, nil
}

func (s *standardCapabilitiesServer) Infos(ctx context.Context, request *emptypb.Empty) (*capabilitiespb.CapabilityInfosReply, error) {
	infos, err := s.impl.Infos(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get capability infos: %w", err)
	}

	var infosReply []*capabilitiespb.CapabilityInfoReply
	for _, info := range infos {
		infosReply = append(infosReply, capability.InfoToReply(info))
	}

	return &capabilitiespb.CapabilityInfosReply{Infos: infosReply}, nil
}

type resourceClosingServer struct {
	StandardCapabilities
	server *standardCapabilitiesServer
}

func (r *resourceClosingServer) Close() error {
	r.server.CloseAll(r.server.resources...)
	return r.StandardCapabilities.Close()
}
