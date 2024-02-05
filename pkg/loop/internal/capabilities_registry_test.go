package internal

import (
	"context"
	"errors"
	"testing"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type registry struct {
	caps map[string]capabilities.BaseCapability
}

func (r *registry) Get(ctx context.Context, id string) (capabilities.BaseCapability, error) {
	c, ok := r.caps[id]
	if !ok {
		return nil, errors.New("capability not found")
	}
	return c, nil
}

type mockTrigger struct {
	capabilities.BaseCapability
	callback chan<- capabilities.CapabilityResponse
}

func (m *mockTrigger) RegisterTrigger(ctx context.Context, callback chan<- capabilities.CapabilityResponse, request capabilities.CapabilityRequest) error {
	m.callback = callback
	return nil
}

func (m *mockTrigger) UnregisterTrigger(ctx context.Context, request capabilities.CapabilityRequest) error {
	m.callback = nil
	return nil
}

func mustMockTrigger(t *testing.T) *mockTrigger {
	return &mockTrigger{
		BaseCapability: capabilities.MustNewCapabilityInfo("trigger", capabilities.CapabilityTypeTrigger, "a mock trigger", "v0.0.1"),
	}
}

type mockCallback struct {
	capabilities.BaseCapability
	callback     chan<- capabilities.CapabilityResponse
	regRequest   capabilities.RegisterToWorkflowRequest
	unregRequest capabilities.UnregisterFromWorkflowRequest
}

func (m *mockCallback) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	m.regRequest = request
	return nil
}
func (m *mockCallback) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	m.unregRequest = request
	return nil
}

func (m *mockCallback) Execute(ctx context.Context, callback chan<- capabilities.CapabilityResponse, request capabilities.CapabilityRequest) error {
	m.callback = callback
	return nil
}

func mustMockCallback(t *testing.T) *mockCallback {
	return &mockCallback{
		BaseCapability: capabilities.MustNewCapabilityInfo("callback", capabilities.CapabilityTypeAction, "a mock action", "v0.0.1"),
	}
}

type registryPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	brokerCfg BrokerConfig
	reg       *registry
}

func (r *registryPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, client *grpc.ClientConn) (any, error) {
	return NewCapabilitiesRegistryClient(broker, r.brokerCfg, client)
}

func (r *registryPlugin) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	RegisterCapabilitiesRegistryServer(server, broker, r.brokerCfg, r.reg)
	return nil
}

func newRegistryPlugin(t *testing.T, reg *registry) (*CapabilitiesRegistryClient, error) {
	stopCh := make(chan struct{})
	logger := logger.Test(t)
	pluginName := "registry"

	client, _ := plugin.TestPluginGRPCConn(
		t,
		map[string]plugin.Plugin{
			pluginName: &registryPlugin{
				brokerCfg: BrokerConfig{
					StopCh: stopCh,
					Logger: logger,
				},
				reg: reg,
			},
		},
	)

	regClient, err := client.Dispense(pluginName)
	require.NoError(t, err)

	return regClient.(*CapabilitiesRegistryClient), nil
}

func Test_CapabilitiesRegistry(t *testing.T) {
	mtr := mustMockTrigger(t)
	mcb := mustMockCallback(t)
	reg := &registry{
		caps: map[string]capabilities.BaseCapability{
			"trigger": mtr,
			"action":  mcb,
		},
	}
	ctx := tests.Context(t)

	rc, err := newRegistryPlugin(t, reg)
	require.NoError(t, err)

	t.Run("capability not found", func(t *testing.T) {
		_, err = rc.Get(ctx, "foo")
		assert.ErrorContains(t, err, "capability not found")
	})

	t.Run("fetching a trigger capability, and executing it", func(t *testing.T) {
		tr, err := rc.Get(ctx, "trigger")
		require.NoError(t, err)

		ch := make(chan capabilities.CapabilityResponse)
		err = tr.(capabilities.TriggerCapability).RegisterTrigger(
			ctx,
			ch,
			capabilities.CapabilityRequest{})
		require.NoError(t, err)

		vs, err := values.NewString("hello")
		require.NoError(t, err)
		cr := capabilities.CapabilityResponse{
			Value: vs,
		}
		mtr.callback <- cr
		assert.Equal(t, cr, <-ch)
	})

	t.Run("fetching a trigger capability, and closing the channel", func(t *testing.T) {
		tr, err := rc.Get(ctx, "trigger")
		require.NoError(t, err)

		ch := make(chan capabilities.CapabilityResponse)
		err = tr.(capabilities.TriggerCapability).RegisterTrigger(
			ctx,
			ch,
			capabilities.CapabilityRequest{})
		require.NoError(t, err)

		// Close the channel from the server, to signal no further results.
		close(mtr.callback)

		// This should propagate to the client.
		_, isOpen := <-ch
		assert.False(t, isOpen)
	})

	t.Run("fetching a trigger capability, and unregistering", func(t *testing.T) {
		tr, err := rc.Get(ctx, "trigger")
		require.NoError(t, err)

		ch := make(chan capabilities.CapabilityResponse)
		err = tr.(capabilities.TriggerCapability).RegisterTrigger(
			ctx,
			ch,
			capabilities.CapabilityRequest{})
		require.NoError(t, err)
		assert.NotNil(t, mtr.callback)

		err = tr.(capabilities.TriggerCapability).UnregisterTrigger(
			ctx,
			capabilities.CapabilityRequest{})
		require.NoError(t, err)

		assert.Nil(t, mtr.callback)
	})

	t.Run("fetching a trigger capability and calling Info", func(t *testing.T) {
		tr, err := rc.Get(tests.Context(t), "trigger")
		require.NoError(t, err)

		gotInfo, err := tr.Info(ctx)
		require.NoError(t, err)

		expectedInfo, err := mtr.Info(ctx)
		require.NoError(t, err)
		assert.Equal(t, expectedInfo, gotInfo)
	})

	t.Run("fetching an action capability, and (un)registering it", func(t *testing.T) {
		tr, err := rc.Get(ctx, "action")
		require.NoError(t, err)

		vmap, err := values.NewMap(map[string]any{"foo": "bar"})
		require.NoError(t, err)
		expectedRequest := capabilities.RegisterToWorkflowRequest{
			Config: vmap,
		}
		err = tr.(capabilities.ActionCapability).RegisterToWorkflow(
			ctx,
			expectedRequest)
		require.NoError(t, err)

		assert.Equal(t, expectedRequest, mcb.regRequest)

		expectedUnrRequest := capabilities.UnregisterFromWorkflowRequest{
			Config: vmap,
		}
		err = tr.(capabilities.ActionCapability).UnregisterFromWorkflow(
			ctx,
			expectedUnrRequest)
		require.NoError(t, err)

		assert.Equal(t, expectedUnrRequest, mcb.unregRequest)
	})

	t.Run("fetching an action capability, and executing it", func(t *testing.T) {
		tr, err := rc.Get(ctx, "action")
		require.NoError(t, err)

		cmap, err := values.NewMap(map[string]any{"foo": "bar"})
		require.NoError(t, err)

		imap, err := values.NewMap(map[string]any{"bar": "baz"})
		require.NoError(t, err)
		expectedRequest := capabilities.CapabilityRequest{
			Config: cmap,
			Inputs: imap,
		}
		ch := make(chan capabilities.CapabilityResponse)
		err = tr.(capabilities.ActionCapability).Execute(
			ctx,
			ch,
			expectedRequest)
		require.NoError(t, err)

		expectedErr := errors.New("an error")
		expectedResp := capabilities.CapabilityResponse{
			Err: expectedErr,
		}

		mcb.callback <- expectedResp
		assert.Equal(t, expectedResp, <-ch)
	})

	t.Run("fetching an action capability, and closing it", func(t *testing.T) {
		tr, err := rc.Get(ctx, "action")
		require.NoError(t, err)

		cmap, err := values.NewMap(map[string]any{"foo": "bar"})
		require.NoError(t, err)

		imap, err := values.NewMap(map[string]any{"bar": "baz"})
		require.NoError(t, err)
		expectedRequest := capabilities.CapabilityRequest{
			Config: cmap,
			Inputs: imap,
		}
		ch := make(chan capabilities.CapabilityResponse)
		err = tr.(capabilities.ActionCapability).Execute(
			ctx,
			ch,
			expectedRequest)
		require.NoError(t, err)

		close(mcb.callback)
		_, isOpen := <-ch
		assert.False(t, isOpen)
	})
}
