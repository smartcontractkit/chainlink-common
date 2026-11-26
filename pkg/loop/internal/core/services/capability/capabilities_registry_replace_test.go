package capability

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
)

// registryPlugin connects one plugin process to a node's registry over a go-plugin
// broker. The registry is served on the go-plugin server side, and the plugin's view
// of it is the go-plugin client side.
type registryPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	brokerExt *net.BrokerExt
	impl      core.CapabilitiesRegistry
}

func (r *registryPlugin) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, client *grpc.ClientConn) (any, error) {
	r.brokerExt.Broker = broker
	return NewCapabilitiesRegistryClient(client, r.brokerExt), nil
}

func (r *registryPlugin) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	r.brokerExt.Broker = broker
	RegisterCapabilitiesRegistryServer(server, r.brokerExt, r.impl)
	return nil
}

// startRegistryPlugin returns a plugin's view of reg, and a func that stops the plugin
// the way a process ends: nothing is removed from the registry.
func startRegistryPlugin(t *testing.T, lggr logger.Logger, reg core.CapabilitiesRegistry) (core.CapabilitiesRegistry, func()) {
	t.Helper()
	stopCh := make(chan struct{})
	const name = "registry"
	client, server := plugin.TestPluginGRPCConn(t, true, map[string]plugin.Plugin{
		name: &registryPlugin{impl: reg, brokerExt: &net.BrokerExt{
			BrokerConfig: net.BrokerConfig{StopCh: stopCh, Logger: lggr},
		}},
	})
	dispensed, err := client.Dispense(name)
	require.NoError(t, err)
	stopped := false
	stop := func() {
		if stopped {
			return
		}
		stopped = true
		close(stopCh)
		// The same order as the deferred calls in TestCapabilitiesRegistry: closing the
		// client first makes go-plugin stop the server a second time, concurrently.
		server.Stop()
		_ = client.Close()
	}
	t.Cleanup(stop)
	return dispensed.(core.CapabilitiesRegistry), stop
}

// A plugin stops without removing its capability from the registry, and callers on the
// node are inside that capability when the plugin's replacement adds itself.
//
// This runs the real registry, the registry client and server, and the LOOP client
// connection with its reconnect loop. Before the fix the replacement's Add never
// returned: a registration call sat in the reconnect loop with the entry's lock held,
// and nothing on that path observed a context.
func TestCapabilitiesRegistry_ReplaceCapabilityOfStoppedPlugin(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	// A client whose plugin is gone logs every attempt to reconnect, so the logs are dropped.
	lggr := logger.Nop()
	reg := registry.NewRegistry(lggr)

	info := capabilities.CapabilityInfo{
		ID:             "trigger-1@1.0.0",
		CapabilityType: capabilities.CapabilityTypeTrigger,
		Description:    "trigger",
	}
	newTrigger := func() mockTriggerCapability {
		return mockTriggerCapability{
			mockBaseCapability:    &mockBaseCapability{info: info},
			mockTriggerExecutable: &mockTriggerExecutable{callback: make(chan capabilities.TriggerResponse, 10)},
		}
	}
	request := func(triggerID string) capabilities.TriggerRegistrationRequest {
		return capabilities.TriggerRegistrationRequest{TriggerID: triggerID}
	}
	goCall := func(fn func() error) <-chan error {
		errCh := make(chan error, 1)
		go func() { errCh <- fn() }()
		return errCh
	}
	requireReturns := func(errCh <-chan error, what string) error {
		t.Helper()
		select {
		case err := <-errCh:
			return err
		case <-time.After(tests.WaitTimeout(t)):
			require.FailNow(t, what+" did not return")
			return nil
		}
	}

	oldPlugin, stopOldPlugin := startRegistryPlugin(t, lggr, reg)
	require.NoError(t, oldPlugin.Add(ctx, newTrigger()))
	onNode, err := reg.GetTrigger(ctx, info.ID)
	require.NoError(t, err)
	events, err := onNode.RegisterTrigger(ctx, request("kept"))
	require.NoError(t, err)

	stopOldPlugin()

	// A registration with a deadline ends at its deadline.
	deadlineCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	err = requireReturns(goCall(func() error {
		_, registerErr := onNode.RegisterTrigger(deadlineCtx, request("with-deadline"))
		return registerErr
	}), "a registration with a deadline")
	require.Error(t, err)

	// Callers without a deadline, as the node's remote trigger publisher is. Each one
	// stays in the client's reconnect loop, or waits behind one that does.
	inFlight := []<-chan error{
		goCall(func() error { _, err := onNode.RegisterTrigger(ctx, request("no-deadline")); return err }),
		goCall(func() error { return onNode.AckEvent(ctx, "kept", "event", "method") }),
		goCall(func() error { return onNode.UnregisterTrigger(ctx, request("other")) }),
	}
	time.Sleep(100 * time.Millisecond) // let them reach the capability

	replacement := newTrigger()
	newPlugin, _ := startRegistryPlugin(t, lggr, reg)
	require.NoError(t, requireReturns(goCall(func() error { return newPlugin.Add(ctx, replacement) }), "Add of the replacement"))

	for _, errCh := range inFlight {
		// Each one ends: with an error when it was on the stopped plugin, without one when
		// it waited for the registrations lock and then reached the replacement.
		_ = requireReturns(errCh, "a call without a deadline")
	}

	// The registration made before the plugin stopped is replayed onto the replacement,
	// and its events arrive on the channel the caller already has.
	replacement.XXXTestingPushToCallbackChan(capabilities.TriggerResponse{Event: capabilities.TriggerEvent{ID: "from-replacement"}})
	timeout := time.After(tests.WaitTimeout(t))
	for {
		select {
		case resp, ok := <-events:
			require.True(t, ok, "the events channel was closed")
			if resp.Event.ID == "from-replacement" {
				return
			}
			// The end of the old plugin's stream arrives first, as an error response.
		case <-timeout:
			require.FailNow(t, "no event from the replacement")
		}
	}
}
