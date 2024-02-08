package internal

import (
	"context"
	"errors"
	"fmt"
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

func (r *registry) GetAction(ctx context.Context, id string) (capabilities.ActionCapability, error) {
	c, ok := r.caps[id]
	if !ok {
		return nil, errors.New("capability not found")
	}
	ac, ok := c.(capabilities.ActionCapability)
	if !ok {
		return nil, errors.New("not an action capability")
	}

	return ac, nil
}

func (r *registry) GetTrigger(ctx context.Context, id string) (capabilities.TriggerCapability, error) {
	c, ok := r.caps[id]
	if !ok {
		return nil, errors.New("capability not found")
	}
	tc, ok := c.(capabilities.TriggerCapability)
	if !ok {
		return nil, errors.New("not an action capability")
	}

	return tc, nil
}

func (r *registry) GetConsensus(ctx context.Context, id string) (capabilities.ConsensusCapability, error) {
	c, ok := r.caps[id]
	if !ok {
		return nil, errors.New("capability not found")
	}
	tc, ok := c.(capabilities.ConsensusCapability)
	if !ok {
		return nil, errors.New("not an action capability")
	}

	return tc, nil
}

func (r *registry) GetTarget(ctx context.Context, id string) (capabilities.TargetCapability, error) {
	c, ok := r.caps[id]
	if !ok {
		return nil, errors.New("capability not found")
	}
	tc, ok := c.(capabilities.TargetCapability)
	if !ok {
		return nil, errors.New("not an action capability")
	}

	return tc, nil
}

func (r *registry) Add(ctx context.Context, bc capabilities.BaseCapability) error {
	info, err := bc.Info(ctx)
	if err != nil {
		return err
	}

	r.caps[info.ID] = bc
	return nil
}

func (r *registry) List(ctx context.Context) ([]capabilities.BaseCapability, error) {
	return nil, nil
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

func mustMockCallback(t *testing.T, _type capabilities.CapabilityType) *mockCallback {
	return &mockCallback{
		BaseCapability: capabilities.MustNewCapabilityInfo(fmt.Sprintf("callback %s", _type), _type, fmt.Sprintf("a mock %s", _type), "v0.0.1"),
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
	ma := mustMockCallback(t, capabilities.CapabilityTypeAction)
	mcon := mustMockCallback(t, capabilities.CapabilityTypeConsensus)
	mt := mustMockCallback(t, capabilities.CapabilityTypeTarget)
	reg := &registry{
		caps: map[string]capabilities.BaseCapability{
			"trigger":   mtr,
			"action":    ma,
			"consensus": mcon,
			"target":    mt,
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
		tr, err := rc.GetTrigger(ctx, "trigger")
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
		tr, err := rc.GetTrigger(ctx, "trigger")
		require.NoError(t, err)

		ch := make(chan capabilities.CapabilityResponse)
		err = tr.RegisterTrigger(
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
		tr, err := rc.GetTrigger(ctx, "trigger")
		require.NoError(t, err)

		ch := make(chan capabilities.CapabilityResponse)
		err = tr.RegisterTrigger(
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
		tr, err := rc.GetAction(ctx, "action")
		require.NoError(t, err)

		vmap, err := values.NewMap(map[string]any{"foo": "bar"})
		require.NoError(t, err)
		expectedRequest := capabilities.RegisterToWorkflowRequest{
			Config: vmap,
		}
		err = tr.RegisterToWorkflow(
			ctx,
			expectedRequest)
		require.NoError(t, err)

		assert.Equal(t, expectedRequest, ma.regRequest)

		expectedUnrRequest := capabilities.UnregisterFromWorkflowRequest{
			Config: vmap,
		}
		err = tr.(capabilities.ActionCapability).UnregisterFromWorkflow(
			ctx,
			expectedUnrRequest)
		require.NoError(t, err)

		assert.Equal(t, expectedUnrRequest, ma.unregRequest)
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

		ma.callback <- expectedResp
		assert.Equal(t, expectedResp, <-ch)
	})

	t.Run("fetching an action capability, and closing it", func(t *testing.T) {
		tr, err := rc.GetAction(ctx, "action")
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
		err = tr.Execute(
			ctx,
			ch,
			expectedRequest)
		require.NoError(t, err)

		close(ma.callback)
		_, isOpen := <-ch
		assert.False(t, isOpen)
	})

	t.Run("getting an capability via Get", func(t *testing.T) {
		ac, err := rc.Get(ctx, "action")
		require.NoError(t, err)
		_, ok := ac.(capabilities.ActionCapability)
		assert.True(t, ok)

		tc, err := rc.Get(ctx, "trigger")
		require.NoError(t, err)
		_, ok = tc.(capabilities.TriggerCapability)
		assert.True(t, ok)

		cc, err := rc.Get(ctx, "consensus")
		require.NoError(t, err)
		_, ok = cc.(capabilities.ConsensusCapability)
		assert.True(t, ok)
	})

	t.Run("adding a capability via Add", func(t *testing.T) {
		id := "capToAdd"
		capToAdd := &mockCallback{
			BaseCapability: capabilities.MustNewCapabilityInfo(id, capabilities.CapabilityTypeAction, "capability to add description", "v0.0.1"),
		}

		err := rc.Add(ctx, capToAdd)
		require.NoError(t, err)

		aclient, err := rc.GetAction(ctx, id)
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
		err = aclient.Execute(
			ctx,
			ch,
			expectedRequest)
		require.NoError(t, err)

		close(capToAdd.callback)
		_, isOpen := <-ch
		assert.False(t, isOpen)
	})
}
