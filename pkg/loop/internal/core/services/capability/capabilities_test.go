package capability

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/tests"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type mockTrigger struct {
	capabilities.BaseCapability
	callback        chan capabilities.TriggerResponse
	triggerActive   bool
	unregisterCalls chan bool
	registerCalls   chan bool

	mu sync.Mutex
}

func (m *mockTrigger) RegisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.triggerActive {
		return nil, errors.New("already registered")
	}

	m.triggerActive = true

	m.registerCalls <- true
	return m.callback, nil
}

func (m *mockTrigger) UnregisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.unregisterCalls <- true

	if m.triggerActive {
		close(m.callback)
		m.callback = nil
		m.triggerActive = false
	}

	return nil
}

func (m *mockTrigger) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	close(m.callback)
	m.callback = nil
	m.triggerActive = false
}

func mustMockTrigger(t *testing.T) *mockTrigger {
	return &mockTrigger{
		BaseCapability:  capabilities.MustNewCapabilityInfo("trigger@1.0.0", capabilities.CapabilityTypeTrigger, "a mock trigger"),
		callback:        make(chan capabilities.TriggerResponse, 10),
		unregisterCalls: make(chan bool, 10),
		registerCalls:   make(chan bool, 10),
	}
}

type mockExecutable struct {
	capabilities.BaseCapability
	callback      chan capabilities.CapabilityResponse
	responseError error

	regRequest   capabilities.RegisterToWorkflowRequest
	unregRequest capabilities.UnregisterFromWorkflowRequest
}

func (m *mockExecutable) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	m.regRequest = request
	return nil
}

func (m *mockExecutable) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	m.unregRequest = request
	return nil
}

func (m *mockExecutable) Execute(ctx context.Context, request capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	if m.responseError != nil {
		return capabilities.CapabilityResponse{}, m.responseError
	}

	return <-m.callback, nil
}

func mustMockExecutable(t *testing.T, _type capabilities.CapabilityType) *mockExecutable {
	return &mockExecutable{
		BaseCapability: capabilities.MustNewCapabilityInfo(fmt.Sprintf("callback-%s@1.0.0", _type), _type, fmt.Sprintf("a mock %s", _type)),
		callback:       make(chan capabilities.CapabilityResponse, 10),
	}
}

type capabilityPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	brokerCfg  net.BrokerConfig
	capability capabilities.BaseCapability
}

func (c *capabilityPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, client *grpc.ClientConn) (any, error) {
	bext := &net.BrokerExt{
		BrokerConfig: c.brokerCfg,
		Broker:       broker,
	}
	switch c.capability.(type) {
	case capabilities.TriggerExecutable:
		return NewTriggerCapabilityClient(bext, client), nil
	case capabilities.Executable:
		return NewExecutableCapabilityClient(bext, client), nil
	}

	panic(fmt.Sprintf("unexpected capability type %T", c.capability))
}

func (c *capabilityPlugin) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	switch tc := c.capability.(type) {
	case capabilities.TriggerCapability:
		return RegisterTriggerCapabilityServer(server, broker, c.brokerCfg, tc)
	case ExecutableCapability:
		return RegisterExecutableCapabilityServer(server, broker, c.brokerCfg, tc)
	}

	return nil
}

func newCapabilityPlugin(t *testing.T, capability capabilities.BaseCapability) (capabilities.BaseCapability,
	*plugin.GRPCClient, *plugin.GRPCServer, error) {
	stopCh := make(chan struct{})
	logger := logger.Test(t)
	pluginName := "registry"

	client, server := plugin.TestPluginGRPCConn(
		t,
		false,
		map[string]plugin.Plugin{
			pluginName: &capabilityPlugin{
				brokerCfg: net.BrokerConfig{
					StopCh: stopCh,
					Logger: logger,
				},
				capability: capability,
			},
		},
	)

	regClient, err := client.Dispense(pluginName)
	require.NoError(t, err)

	return regClient.(capabilities.BaseCapability), client, server, nil
}

func Test_Capabilities(t *testing.T) {
	testContext := tests.Context(t)

	t.Run("fetching a trigger capability and sending responses propagate to client", func(t *testing.T) {
		ctx, cancel := context.WithCancel(testContext)
		defer cancel()

		mtr := mustMockTrigger(t)
		tr, _, _, err := newCapabilityPlugin(t, mtr)
		require.NoError(t, err)

		ctr := tr.(capabilities.TriggerCapability)

		ch, err := ctr.RegisterTrigger(
			ctx,
			capabilities.TriggerRegistrationRequest{})
		require.NoError(t, err)

		v, err := values.NewMap(map[string]any{"hello": "world"})
		require.NoError(t, err)

		cr1 := capabilities.TriggerResponse{
			Event: capabilities.TriggerEvent{
				Outputs: v,
			},
		}
		mtr.callback <- cr1

		v, err = values.NewMap(map[string]any{"hello": "world"})
		require.NoError(t, err)

		cr2 := capabilities.TriggerResponse{
			Event: capabilities.TriggerEvent{
				Outputs: v,
			},
		}
		mtr.callback <- cr2

		assert.Equal(t, cr1, <-ch)
	})

	t.Run("fetching a trigger capability and stopping the underlying trigger closes the client channel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(testContext)
		defer cancel()

		mtr := mustMockTrigger(t)
		tr, _, _, err := newCapabilityPlugin(t, mtr)
		require.NoError(t, err)

		ctr := tr.(capabilities.TriggerCapability)

		ch, err := ctr.RegisterTrigger(
			ctx,
			capabilities.TriggerRegistrationRequest{})
		require.NoError(t, err)

		// Wait for registration to complete
		<-mtr.registerCalls

		// Stop the trigger
		mtr.Stop()

		// This should propagate to the client.
		_, isOpen := <-ch
		assert.False(t, isOpen)
	})

	t.Run("fetching a trigger capability and closing the client connection should close the client channel and unregister the trigger", func(t *testing.T) {
		ctx, cancel := context.WithCancel(testContext)
		defer cancel()

		mtr := mustMockTrigger(t)
		tr, client, _, err := newCapabilityPlugin(t, mtr)
		require.NoError(t, err)

		ctr := tr.(capabilities.TriggerCapability)

		ch, err := ctr.RegisterTrigger(
			ctx,
			capabilities.TriggerRegistrationRequest{})
		require.NoError(t, err)

		// Wait for registration to complete
		<-mtr.registerCalls
		assert.NotNil(t, mtr.callback)

		err = client.Close()
		require.NoError(t, err)

		// Closing the client will result in an error being
		// bubbled back to the client.
		resp := <-ch
		assert.Equal(t, status.Code(resp.Err), codes.Unavailable)

		resp, isOpen := <-ch
		assert.False(t, isOpen)

		<-mtr.unregisterCalls
	})

	t.Run("fetching a trigger capability and stopping the server should close the client channel and unregister the trigger", func(t *testing.T) {
		ctx, cancel := context.WithCancel(testContext)
		defer cancel()

		mtr := mustMockTrigger(t)
		tr, _, server, err := newCapabilityPlugin(t, mtr)
		require.NoError(t, err)

		ctr := tr.(capabilities.TriggerCapability)

		ch, err := ctr.RegisterTrigger(
			ctx,
			capabilities.TriggerRegistrationRequest{})
		require.NoError(t, err)

		// Wait for registration to complete
		<-mtr.registerCalls
		assert.NotNil(t, mtr.callback)

		server.Stop()

		// Closing the client will result in an error being
		// bubbled back to the client.
		resp := <-ch
		assert.Equal(t, status.Code(resp.Err), codes.Unavailable)

		_, isOpen := <-ch
		assert.False(t, isOpen)

		<-mtr.unregisterCalls
	})

	t.Run("fetching a trigger capability and unregistering should close client channel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(testContext)
		defer cancel()

		mtr := mustMockTrigger(t)
		tr, _, _, err := newCapabilityPlugin(t, mtr)
		require.NoError(t, err)

		ctr := tr.(capabilities.TriggerCapability)

		ch, err := ctr.RegisterTrigger(
			ctx,
			capabilities.TriggerRegistrationRequest{})
		require.NoError(t, err)

		// Wait for registration to complete
		<-mtr.registerCalls
		assert.NotNil(t, mtr.callback)

		err = ctr.UnregisterTrigger(
			ctx,
			capabilities.TriggerRegistrationRequest{})

		require.NoError(t, err)

		<-mtr.unregisterCalls

		_, isOpen := <-ch
		assert.False(t, isOpen)
	})

	t.Run("fetching a trigger capability and cancelling context should unregister trigger and close client channel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(testContext)
		defer cancel()

		mtr := mustMockTrigger(t)
		tr, _, _, err := newCapabilityPlugin(t, mtr)
		require.NoError(t, err)

		ctr := tr.(capabilities.TriggerCapability)

		ctxWithCancel, cancel := context.WithCancel(ctx)

		ch, err := ctr.RegisterTrigger(
			ctxWithCancel,
			capabilities.TriggerRegistrationRequest{})
		require.NoError(t, err)

		// Wait for registration to complete
		<-mtr.registerCalls
		assert.NotNil(t, mtr.callback)

		cancel()

		<-mtr.unregisterCalls

		_, isOpen := <-ch
		assert.False(t, isOpen)
	})

	t.Run("fetching a trigger capability and calling Info", func(t *testing.T) {
		ctx, cancel := context.WithCancel(testContext)
		defer cancel()

		mtr := mustMockTrigger(t)
		tr, _, _, err := newCapabilityPlugin(t, mtr)
		require.NoError(t, err)

		gotInfo, err := tr.Info(ctx)
		require.NoError(t, err)

		expectedInfo, err := mtr.Info(ctx)
		require.NoError(t, err)
		assert.Equal(t, expectedInfo, gotInfo)
	})

	t.Run("fetching an action capability, and (un)registering it", func(t *testing.T) {
		ctx, cancel := context.WithCancel(testContext)
		defer cancel()

		ma := mustMockExecutable(t, capabilities.CapabilityTypeAction)
		c, _, _, err := newCapabilityPlugin(t, ma)
		require.NoError(t, err)

		act := c.(capabilities.ActionCapability)

		vmap, err := values.NewMap(map[string]any{"foo": "bar"})
		require.NoError(t, err)
		expectedRequest := capabilities.RegisterToWorkflowRequest{
			Config: vmap,
		}
		err = act.RegisterToWorkflow(
			ctx,
			expectedRequest)
		require.NoError(t, err)

		assert.Equal(t, expectedRequest, ma.regRequest)

		expectedUnrRequest := capabilities.UnregisterFromWorkflowRequest{
			Config: vmap,
		}
		err = act.UnregisterFromWorkflow(
			ctx,
			expectedUnrRequest)
		require.NoError(t, err)

		assert.Equal(t, expectedUnrRequest, ma.unregRequest)
	})

	t.Run("fetching an action capability, and executing it", func(t *testing.T) {
		ctx, cancel := context.WithCancel(testContext)
		defer cancel()

		ma := mustMockExecutable(t, capabilities.CapabilityTypeAction)
		c, _, _, err := newCapabilityPlugin(t, ma)
		require.NoError(t, err)

		cmap, err := values.NewMap(map[string]any{"foo": "bar"})
		require.NoError(t, err)

		imap, err := values.NewMap(map[string]any{"bar": "baz"})
		require.NoError(t, err)
		expectedRequest := capabilities.CapabilityRequest{
			Config: cmap,
			Inputs: imap,
		}

		expectedResp := capabilities.CapabilityResponse{
			Value: values.EmptyMap(),
		}

		ma.callback <- expectedResp

		resp, err := c.(capabilities.ActionCapability).Execute(
			ctx,
			expectedRequest)
		require.NoError(t, err)

		assert.Equal(t, expectedResp, resp)
	})

	t.Run("fetching an action capability, and executing it with error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(testContext)
		defer cancel()

		ma := mustMockExecutable(t, capabilities.CapabilityTypeAction)
		c, _, _, err := newCapabilityPlugin(t, ma)
		require.NoError(t, err)

		cmap, err := values.NewMap(map[string]any{"foo": "bar"})
		require.NoError(t, err)

		imap, err := values.NewMap(map[string]any{"bar": "baz"})
		require.NoError(t, err)
		expectedRequest := capabilities.CapabilityRequest{
			Config: cmap,
			Inputs: imap,
		}

		ma.responseError = errors.New("bang")

		_, err = c.(capabilities.ActionCapability).Execute(
			ctx,
			expectedRequest)
		require.NotNil(t, err)
		assert.Equal(t, "bang", err.Error())
	})

	t.Run("fetching an action capability, and closing it", func(t *testing.T) {
		ctx, cancel := context.WithCancel(testContext)
		defer cancel()

		ma := mustMockExecutable(t, capabilities.CapabilityTypeAction)
		c, _, _, err := newCapabilityPlugin(t, ma)
		require.NoError(t, err)

		cmap, err := values.NewMap(map[string]any{"foo": "bar"})
		require.NoError(t, err)

		imap, err := values.NewMap(map[string]any{"bar": "baz"})
		require.NoError(t, err)
		expectedRequest := capabilities.CapabilityRequest{
			Config: cmap,
			Inputs: imap,
		}

		ma.callback <- capabilities.CapabilityResponse{}
		_, err = c.(capabilities.ActionCapability).Execute(
			ctx,
			expectedRequest)
		require.NoError(t, err)
	})

	t.Run("calling execute should be synchronous", func(t *testing.T) {
		ctx, cancel := context.WithCancel(testContext)
		defer cancel()

		ma := mustSynchronousCallback(t, capabilities.CapabilityTypeAction)
		ma.callback <- capabilities.CapabilityResponse{}

		c, _, _, err := newCapabilityPlugin(t, ma)
		require.NoError(t, err)

		cmap, err := values.NewMap(map[string]any{"foo": "bar"})
		require.NoError(t, err)

		imap, err := values.NewMap(map[string]any{"bar": "baz"})
		require.NoError(t, err)
		expectedRequest := capabilities.CapabilityRequest{
			Config: cmap,
			Inputs: imap,
		}

		assert.False(t, ma.executeCalled)

		_, err = c.(capabilities.ActionCapability).Execute(
			ctx,
			expectedRequest)
		require.NoError(t, err)

		assert.True(t, ma.executeCalled)
	})
}

type synchronousCallback struct {
	capabilities.BaseCapability
	callback      chan capabilities.CapabilityResponse
	executeCalled bool
}

func (m *synchronousCallback) RegisterToWorkflow(ctx context.Context, request capabilities.RegisterToWorkflowRequest) error {
	return nil
}

func (m *synchronousCallback) UnregisterFromWorkflow(ctx context.Context, request capabilities.UnregisterFromWorkflowRequest) error {
	return nil
}

func (m *synchronousCallback) Execute(ctx context.Context, request capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	m.executeCalled = true
	return <-m.callback, nil
}

func mustSynchronousCallback(t *testing.T, _type capabilities.CapabilityType) *synchronousCallback {
	return &synchronousCallback{
		BaseCapability: capabilities.MustNewCapabilityInfo(fmt.Sprintf("callback-%s@1.0.0", _type), _type, fmt.Sprintf("a mock %s", _type)),
		callback:       make(chan capabilities.CapabilityResponse, 10),
		executeCalled:  false,
	}
}
