package loop

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

const PluginEmptyName = "empty"

func EmptyHandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		MagicCookieKey:   "CL_PLUGIN_EMPTY_MAGIC_COOKIE",
		MagicCookieValue: "2543d46d-e317-4bc4-a6be-1c018bf3a8c7",
	}
}

// EmptyLoop is a LOOP whose only gRPC service is the [services.Service] lifecycle of the process. The host
// supervises the process. The process gets what it needs from the host in other ways, for example from env vars.
type EmptyLoop struct {
	plugin.NetRPCUnsupportedPlugin
	BrokerConfig
	PluginServer services.Service

	pluginClient *emptyClient
}

func (p *EmptyLoop) GRPCServer(_ *plugin.GRPCBroker, server *grpc.Server) error {
	pb.RegisterServiceServer(server, &goplugin.ServiceServer{Srv: p.PluginServer})
	return nil
}

// GRPCClient reuses one client across launches, so that the client held by the plugin service always
// talks to the current process.
func (p *EmptyLoop) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (any, error) {
	if p.pluginClient == nil {
		p.pluginClient = newEmptyClient(p.BrokerConfig)
	}
	p.pluginClient.Refresh(broker, conn)
	return p.pluginClient, nil
}

func (p *EmptyLoop) ClientConfig() *plugin.ClientConfig {
	return ManagedGRPCClientConfig(&plugin.ClientConfig{
		HandshakeConfig: EmptyHandshakeConfig(),
		Plugins:         map[string]plugin.Plugin{PluginEmptyName: p},
	}, p.BrokerConfig)
}

// ServeEmpty serves an [EmptyLoop] until the host stops this process. impl must already be started; the
// host reports its health and closes it. newServer makes the gRPC server.
func ServeEmpty(impl services.Service, newServer func([]grpc.ServerOption) *grpc.Server) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: EmptyHandshakeConfig(),
		Plugins:         map[string]plugin.Plugin{PluginEmptyName: &EmptyLoop{PluginServer: impl}},
		GRPCServer:      newServer,
	})
}

// EmptyService launches an [EmptyLoop] process, and launches it again when it stops or does not
// answer pings.
type EmptyService struct {
	goplugin.PluginService[*EmptyLoop, services.Service]
}

func NewEmptyService(lggr logger.Logger, grpcOpts GRPCOpts, cmd func() *exec.Cmd) *EmptyService {
	newService := func(_ context.Context, instance any) (services.Service, services.HealthReporter, error) {
		c, ok := instance.(*emptyClient)
		if !ok {
			return nil, nil, fmt.Errorf("expected *emptyClient but got %T", instance)
		}
		return c, c, nil
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "EmptyService")
	var es EmptyService
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	es.Init(PluginEmptyName, &EmptyLoop{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &es
}

// emptyClient is the host side of an [EmptyLoop]. Its health comes from the process, so it reports an
// error while the process is down.
type emptyClient struct {
	*goplugin.PluginClient
	*goplugin.ServiceClient
}

func newEmptyClient(brokerCfg BrokerConfig) *emptyClient {
	brokerCfg.Logger = logger.Named(brokerCfg.Logger, "EmptyClient")
	pc := goplugin.NewPluginClient(brokerCfg)
	return &emptyClient{PluginClient: pc, ServiceClient: goplugin.NewServiceClient(pc.BrokerExt, pc)}
}
