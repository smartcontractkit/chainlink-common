package loop

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	standardcapability "github.com/smartcontractkit/chainlink-common/pkg/loop/internal/core/services/capability/standard"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

const PluginStandardCapabilitiesName = "standardcapabilities"

func StandardCapabilitiesHandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		MagicCookieKey:   "CL_PLUGIN_STANDARD_CAPABILITIES_MAGIC_COOKIE",
		MagicCookieValue: "f4df86d3-3552-4231-8206-be0d245b6c67",
	}
}

type StandardCapabilitiesLoop struct {
	Logger logger.Logger
	plugin.NetRPCUnsupportedPlugin
	BrokerConfig
	PluginServer StandardCapabilities
	pluginClient *standardcapability.StandardCapabilitiesClient
}

func (p *StandardCapabilitiesLoop) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	return standardcapability.RegisterStandardCapabilitiesServer(server, broker, p.BrokerConfig, p.PluginServer)
}

func (p *StandardCapabilitiesLoop) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (any, error) {
	if p.pluginClient == nil {
		p.pluginClient = standardcapability.NewStandardCapabilitiesClient(p.BrokerConfig)
	}
	p.pluginClient.Refresh(broker, conn)

	return StandardCapabilities(p.pluginClient), nil
}

func (p *StandardCapabilitiesLoop) ClientConfig() *plugin.ClientConfig {
	clientConfig := &plugin.ClientConfig{
		HandshakeConfig: StandardCapabilitiesHandshakeConfig(),
		Plugins:         map[string]plugin.Plugin{PluginStandardCapabilitiesName: p},
	}
	if p.pluginClient == nil {
		p.pluginClient = standardcapability.NewStandardCapabilitiesClient(p.BrokerConfig)
	}
	return ManagedGRPCClientConfig(clientConfig, p.pluginClient.BrokerConfig)
}

type StandardCapabilities interface {
	services.Service
	Initialise(ctx context.Context, dependencies core.StandardCapabilitiesDependencies) error
	Infos(ctx context.Context) ([]capabilities.CapabilityInfo, error)
}

type StandardCapabilitiesService struct {
	goplugin.PluginService[*StandardCapabilitiesLoop, StandardCapabilities]
}

func NewStandardCapabilitiesService(lggr logger.Logger, grpcOpts GRPCOpts, cmd func() *exec.Cmd) *StandardCapabilitiesService {
	newService := func(ctx context.Context, instance any) (StandardCapabilities, services.HealthReporter, error) {
		scs, ok := instance.(StandardCapabilities)
		if !ok {
			return nil, nil, fmt.Errorf("expected StandardCapabilities but got %T", instance)
		}
		return scs, scs, nil
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "StandardCapabilities")
	var rs StandardCapabilitiesService
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	rs.Init(PluginStandardCapabilitiesName, &StandardCapabilitiesLoop{Logger: lggr, BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &rs
}

func Serve[T StandardCapabilities](serviceName string, createPluginServer func(logger.Logger) T) {
	s := MustNewStartedServer(serviceName)
	defer s.Stop()

	s.Logger.Infof("Starting %s", serviceName)

	stopCh := make(chan struct{})
	defer close(stopCh)

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: StandardCapabilitiesHandshakeConfig(),
		Plugins: map[string]plugin.Plugin{
			PluginStandardCapabilitiesName: &StandardCapabilitiesLoop{
				PluginServer: createPluginServer(s.Logger),
				BrokerConfig: BrokerConfig{Logger: s.Logger, StopCh: stopCh, GRPCOpts: s.GRPCOpts},
			},
		},
		GRPCServer: s.GRPCOpts.NewServer,
	})
}
