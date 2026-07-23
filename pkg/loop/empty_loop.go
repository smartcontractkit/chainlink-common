package loop

import (
	"context"
	"os/exec"

	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/goplugin"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
)

const PluginEmptyName = "empty"

// EmptyHandshakeConfig is the handshake for a plugin that exposes no services.
func EmptyHandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		MagicCookieKey:   "CL_PLUGIN_EMPTY_MAGIC_COOKIE",
		MagicCookieValue: "b9a1e7c4-3f2d-4a6b-8c1e-5d9f0a2b3c4d",
	}
}

// EmptyLoop is a LOOP plugin that registers no services of its own. It exists so
// a binary can run under a host's go-plugin lifecycle — the standard handshake
// plus go-plugin's built-in liveness health check — without exposing any RPCs.
//
// It is used for processes the node only needs to launch and supervise (e.g. a
// standalone networking proxy), while their real work is served elsewhere. As
// with any LOOP, the plugin's health is process liveness: it is healthy for as
// long as the process (and thus the services started alongside it) is running.
type EmptyLoop struct {
	plugin.NetRPCUnsupportedPlugin
	BrokerConfig
}

func (*EmptyLoop) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	return nil
}

func (*EmptyLoop) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, conn *grpc.ClientConn) (any, error) {
	// No services are exposed; the host only supervises liveness.
	return emptyClient{}, nil
}

// ClientConfig returns the go-plugin client config a host uses to launch and
// supervise the empty plugin.
func (p *EmptyLoop) ClientConfig() *plugin.ClientConfig {
	c := &plugin.ClientConfig{
		HandshakeConfig: EmptyHandshakeConfig(),
		Plugins:         map[string]plugin.Plugin{PluginEmptyName: p},
	}
	return ManagedGRPCClientConfig(c, p.BrokerConfig)
}

// EmptyService launches and supervises a binary that serves the empty LOOP,
// relaunching it if it dies. It is a services.Service: healthy while the
// launched process is alive.
type EmptyService struct {
	goplugin.PluginService[*EmptyLoop, services.Service]
}

// NewEmptyService returns a service that launches cmd as an empty LOOP.
func NewEmptyService(lggr logger.Logger, grpcOpts GRPCOpts, cmd func() *exec.Cmd) *EmptyService {
	newService := func(ctx context.Context, instance any) (services.Service, services.HealthReporter, error) {
		s := emptyClient{}
		return s, s, nil
	}
	stopCh := make(chan struct{})
	lggr = logger.Named(lggr, "EmptyLoop")
	var es EmptyService
	broker := BrokerConfig{StopCh: stopCh, Logger: lggr, GRPCOpts: grpcOpts}
	es.Init(PluginEmptyName, &EmptyLoop{BrokerConfig: broker}, newService, lggr, cmd, stopCh)
	return &es
}

// emptyClient is the no-op service dispensed by EmptyLoop; there are no RPCs to
// call, so liveness is tracked by the plugin host via go-plugin's health check.
type emptyClient struct{}

func (emptyClient) Start(context.Context) error    { return nil }
func (emptyClient) Close() error                   { return nil }
func (emptyClient) Ready() error                   { return nil }
func (emptyClient) HealthReport() map[string]error { return map[string]error{"EmptyLoop": nil} }
func (emptyClient) Name() string                   { return "EmptyLoop" }
