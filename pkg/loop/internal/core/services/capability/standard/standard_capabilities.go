package capability

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/capability"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/errorlog"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/gateway"
	keystoreservice "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/keystore"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/keyvalue"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/oraclefactory"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/orgresolver"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/pipeline"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/settings"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/telemetry"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	gatewayconnectorpb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/gatewayconnector"
	oraclefactorypb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/oraclefactory"
	relayersetpb "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb/relayerset"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/relayerset"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

type StandardCapabilities interface {
	services.Service
	Initialise(ctx context.Context, dependencies core.StandardCapabilitiesDependencies) error
	Infos(ctx context.Context) ([]capabilities.CapabilityInfo, error)
}

type StandardCapabilitiesClient struct {
	*goplugin.PluginClient
	capabilitiespb.StandardCapabilitiesClient
	*goplugin.ServiceClient
	*net.BrokerExt

	resources      []net.Resource
	initializeDeps *core.StandardCapabilitiesDependencies
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

// Reinitialise calls Initialise with cached deps from the previous call, if one was already made.
func (c *StandardCapabilitiesClient) Reinitialise(ctx context.Context) error {
	if c.initializeDeps == nil {
		c.Logger.Debug("No dependencies to re-initialise")
		return nil
	}
	c.CloseAll(c.resources...)
	c.resources = nil
	c.Logger.Info("Re-initialising dependencies")
	return c.Initialise(ctx, *c.initializeDeps)
}

func (c *StandardCapabilitiesClient) Initialise(ctx context.Context, dependencies core.StandardCapabilitiesDependencies) error {
	config := dependencies.Config
	telemetryService := dependencies.TelemetryService
	keyValueStore := dependencies.Store
	capabilitiesRegistry := dependencies.CapabilityRegistry
	errorLog := dependencies.ErrorLog
	pipelineRunner := dependencies.PipelineRunner
	relayerSet := dependencies.RelayerSet
	oracleFactory := dependencies.OracleFactory
	gatewayConnector := dependencies.GatewayConnector
	p2pKeystore := dependencies.P2PKeystore
	orgResolver := dependencies.OrgResolver
	creSettings := dependencies.CRESettings
	telemetryID, telemetryRes, err := c.ServeNew("Telemetry", func(s *grpc.Server) {
		pb.RegisterTelemetryServer(s, telemetry.NewTelemetryServer(telemetryService))
	})

	if err != nil {
		return fmt.Errorf("failed to serve new telemetry: %w", err)
	}
	var resources []net.Resource
	resources = append(resources, telemetryRes)

	keyStoreID, keyStoreRes, err := c.ServeNew("KeyStore", func(s *grpc.Server) {
		pb.RegisterKeystoreServer(s, keystoreservice.NewServer(p2pKeystore))
	})
	if err != nil {
		c.CloseAll(resources...)
		return fmt.Errorf("failed to serve new keyStore: %w", err)
	}
	resources = append(resources, keyStoreRes)

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
		relayersetpb.RegisterRelayerSetServerWithDependants(s, relayerSetServer)
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

	gatewayConnectorID, gatewayConnectorRes, err := c.ServeNew("GatewayConnector", func(s *grpc.Server) {
		gatewayconnectorpb.RegisterGatewayConnectorServer(s, gateway.NewGatewayConnectorServer(c.BrokerExt, gatewayConnector))
	})
	if err != nil {
		c.CloseAll(resources...)
		return fmt.Errorf("failed to serve gateway connector: %w", err)
	}
	resources = append(resources, gatewayConnectorRes)

	orgResolverID, orgResolverRes, err := c.ServeNew("OrgResolver", func(s *grpc.Server) {
		pb.RegisterOrgResolverServer(s, orgresolver.NewServer(orgResolver))
	})
	if err != nil {
		c.CloseAll(resources...)
		return fmt.Errorf("failed to serve org resolver: %w", err)
	}
	resources = append(resources, orgResolverRes)

	var creSettingsID uint32
	if creSettings != nil {
		var creSettingsRes net.Resource
		creSettingsID, creSettingsRes, err = c.ServeNew("CRESettings", func(s *grpc.Server) {
			capabilitiespb.RegisterSettingsServer(s, settings.NewServer(creSettings))
		})
		if err != nil {
			c.CloseAll(resources...)
			return fmt.Errorf("failed to serve cre settings: %w", err)
		}
		resources = append(resources, creSettingsRes)
	}

	_, err = c.StandardCapabilitiesClient.Initialise(ctx, &capabilitiespb.InitialiseRequest{
		Config:             config,
		ErrorLogId:         errorLogID,
		PipelineRunnerId:   pipelineRunnerID,
		TelemetryId:        telemetryID,
		CapRegistryId:      capabilitiesRegistryID,
		KeyValueStoreId:    keyValueStoreID,
		RelayerSetId:       relayerSetID,
		OracleFactoryId:    oracleFactoryID,
		GatewayConnectorId: gatewayConnectorID,
		KeystoreId:         keyStoreID,
		OrgResolverId:      orgResolverID,
		CreSettingsId:      creSettingsID,
	})

	if err != nil {
		c.CloseAll(resources...)
		return fmt.Errorf("failed to initialise standard capability: %w", err)
	}

	c.resources = resources
	c.initializeDeps = &dependencies

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

	keystoreConn, err := s.Dial(request.KeystoreId)
	if err != nil {
		s.CloseAll(resources...)
		return nil, net.ErrConnDial{Name: "Keystore", ID: request.KeystoreId, Err: err}
	}
	resources = append(resources, net.Resource{Closer: keystoreConn, Name: "KeystoreConn"})
	keyStore := keystoreservice.NewClient(keystoreConn)

	// Sets the auth header signing mechanism
	beholder.GetClient().SetSigner(keyStore)

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

	gatewayConnectorConn, err := s.Dial(request.GatewayConnectorId)
	if err != nil {
		s.CloseAll(resources...)
		return nil, net.ErrConnDial{Name: "GatewayConnector", ID: request.GatewayConnectorId, Err: err}
	}
	resources = append(resources, net.Resource{Closer: gatewayConnectorConn, Name: "GatewayConnector"})
	gatewayConnector := gateway.NewGatewayConnectorClient(gatewayConnectorConn, s.BrokerExt)

	orgResolverConn, err := s.Dial(request.OrgResolverId)
	if err != nil {
		s.CloseAll(resources...)
		return nil, net.ErrConnDial{Name: "OrgResolver", ID: request.OrgResolverId, Err: err}
	}
	resources = append(resources, net.Resource{Closer: orgResolverConn, Name: "OrgResolver"})
	orgResolver := orgresolver.NewClient(s.Logger, orgResolverConn)

	var creSettings core.SettingsBroadcaster
	if request.CreSettingsId > 0 {
		creSettingsConn, err := s.Dial(request.CreSettingsId)
		if err != nil {
			s.CloseAll(resources...)
			return nil, net.ErrConnDial{Name: "CRESettings", ID: request.OrgResolverId, Err: err}
		}
		resources = append(resources, net.Resource{Closer: orgResolverConn, Name: "CRESettings"})
		creSettings = settings.NewClient(s.Logger, creSettingsConn)
	}

	dependencies := core.StandardCapabilitiesDependencies{
		Config:             request.Config,
		TelemetryService:   telemetry,
		Store:              keyValueStore,
		CapabilityRegistry: capabilitiesRegistry,
		ErrorLog:           errorLog,
		PipelineRunner:     pipelineRunner,
		RelayerSet:         relayerSet,
		OracleFactory:      oracleFactory,
		GatewayConnector:   gatewayConnector,
		P2PKeystore:        keyStore,
		OrgResolver:        orgResolver,
		CRESettings:        creSettings,
	}

	if err = s.impl.Initialise(ctx, dependencies); err != nil {
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
