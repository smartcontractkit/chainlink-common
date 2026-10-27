package loop

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

const PluginEmptyName = "empty"

func EmptyHandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		MagicCookieKey:   "CL_PLUGIN_EMPTY_MAGIC_COOKIE",
		MagicCookieValue: "2543d46d-e317-4bc4-a6be-1c018bf3a8c7",
	}
}

// EmptyLoop is a LOOP with no gRPC services. The host only supervises the process. The process gets
// what it needs from the host in other ways, for example from env vars.
type EmptyLoop struct {
	plugin.NetRPCUnsupportedPlugin
	BrokerConfig
}

func (*EmptyLoop) GRPCServer(*plugin.GRPCBroker, *grpc.Server) error {
	return nil
}

func (p *EmptyLoop) GRPCClient(context.Context, *plugin.GRPCBroker, *grpc.ClientConn) (any, error) {
	return &emptyClient{name: logger.Named(p.Logger, "EmptyClient").Name()}, nil
}

func (p *EmptyLoop) ClientConfig() *plugin.ClientConfig {
	return ManagedGRPCClientConfig(&plugin.ClientConfig{
		HandshakeConfig: EmptyHandshakeConfig(),
		Plugins:         map[string]plugin.Plugin{PluginEmptyName: p},
	}, p.BrokerConfig)
}

// ServeEmpty serves an [EmptyLoop] until the host stops this process. newServer makes the gRPC server
// that go-plugin uses for its own services, such as the health check.
func ServeEmpty(newServer func([]grpc.ServerOption) *grpc.Server) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: EmptyHandshakeConfig(),
		Plugins:         map[string]plugin.Plugin{PluginEmptyName: &EmptyLoop{}},
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

// emptyClient is the host side of an [EmptyLoop]. The process has no state to report: the plugin
// service pings the process to know that it is alive.
type emptyClient struct {
	name string
}

func (c *emptyClient) Start(context.Context) error { return nil }

func (c *emptyClient) Close() error { return nil }

func (c *emptyClient) Ready() error { return nil }

func (c *emptyClient) HealthReport() map[string]error { return map[string]error{c.name: nil} }

func (c *emptyClient) Name() string { return c.name }
