package capability

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/go-plugin"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	caperrors "github.com/smartcontractkit/chainlink-common/pkg/capabilities/errors"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/loop/internal/net"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

type mockTrigger struct {
	capabilities.BaseCapability
	callback            chan capabilities.TriggerResponse
	triggerActive       bool
	unregisterCalls     chan bool
	registerCalls       chan bool
	failedToRegisterErr *string

	mu sync.Mutex
}

func (m *mockTrigger) RegisterTrigger(ctx context.Context, request capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.triggerActive {
		return nil, errors.New("already registered")
	}

	if m.failedToRegisterErr != nil {
		return nil, errors.New(*m.failedToRegisterErr)
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

func (m *mockTrigger) AckEvent(ctx context.Context, triggerId string, eventId string, method string) error {
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
	callback       chan capabilities.CapabilityResponse
	responseError  error
	executeEntered chan struct{}

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
	if m.executeEntered != nil {
		select {
		case m.executeEntered <- struct{}{}:
		default:
		}
	}
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

// malformedExecutableServer emits a raw pb.CapabilityResponse that the
// generated CapabilityResponseFromProto rejects (ConfigDigest is not 32
// bytes). It is used to drive the unmarshal-failure path in
// executableClient.Execute without going through executableServer's
// CapabilityResponseToProto conversion.
type malformedExecutableServer struct {
	pb.UnimplementedExecutableServer
}

func (m *malformedExecutableServer) Execute(_ *pb.CapabilityRequest, stream grpc.ServerStreamingServer[pb.CapabilityResponse]) error {
	return stream.Send(&pb.CapabilityResponse{
		OcrAttestation: &pb.OCRAttestation{
			ConfigDigest: []byte{1}, // 1 byte, CapabilityResponseFromProto() requires 32
		},
	})
}

type malformedExecutablePlugin struct {
	plugin.NetRPCUnsupportedPlugin
	brokerCfg net.BrokerConfig
}

func (p *malformedExecutablePlugin) GRPCClient(_ context.Context, broker *plugin.GRPCBroker, client *grpc.ClientConn) (any, error) {
	bext := &net.BrokerExt{BrokerConfig: p.brokerCfg, Broker: broker}
	return NewExecutableCapabilityClient(bext, client), nil
}

func (p *malformedExecutablePlugin) GRPCServer(_ *plugin.GRPCBroker, server *grpc.Server) error {
	pb.RegisterExecutableServer(server, &malformedExecutableServer{})
	return nil
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

func Test_RegisterTrigger(t *testing.T) {
	t.Run("async RegisterTrigger implementation returns error to server", func(t *testing.T) {
		errMsg := "boom"
		mtr := mustMockTrigger(t)
		mtr.failedToRegisterErr = &errMsg

		tr, _, _, err := newCapabilityPlugin(t, mtr)
		require.NoError(t, err)

		ctr := tr.(capabilities.TriggerCapability)

		_, err = ctr.RegisterTrigger(
			t.Context(),
			capabilities.TriggerRegistrationRequest{})
		require.ErrorContains(t, err, errMsg)
	})
}

func Test_Capabilities(t *testing.T) {
	t.Run("fetching a trigger capability and sending responses propagate to client", func(t *testing.T) {
		mtr := mustMockTrigger(t)
		tr, _, _, err := newCapabilityPlugin(t, mtr)
		require.NoError(t, err)

		ctr := tr.(capabilities.TriggerCapability)

		ch, err := ctr.RegisterTrigger(
			t.Context(),
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
		mtr := mustMockTrigger(t)
		tr, _, _, err := newCapabilityPlugin(t, mtr)
		require.NoError(t, err)

		ctr := tr.(capabilities.TriggerCapability)

		ch, err := ctr.RegisterTrigger(
			t.Context(),
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
		ctx, cancel := context.WithCancel(t.Context())
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
		assert.Equal(t, codes.Unavailable, status.Code(resp.Err))

		resp, isOpen := <-ch
		assert.False(t, isOpen)

		<-mtr.unregisterCalls
	})

	t.Run("fetching a trigger capability and stopping the server should close the client channel and unregister the trigger", func(t *testing.T) {
		mtr := mustMockTrigger(t)
		tr, _, server, err := newCapabilityPlugin(t, mtr)
		require.NoError(t, err)

		ctr := tr.(capabilities.TriggerCapability)

		ch, err := ctr.RegisterTrigger(
			t.Context(),
			capabilities.TriggerRegistrationRequest{})
		require.NoError(t, err)

		// Wait for registration to complete
		<-mtr.registerCalls
		assert.NotNil(t, mtr.callback)

		server.Stop()

		// Closing the client will result in an error being
		// bubbled back to the client.
		resp := <-ch
		assert.Equal(t, codes.Unavailable, status.Code(resp.Err))

		_, isOpen := <-ch
		assert.False(t, isOpen)

		<-mtr.unregisterCalls
	})

	t.Run("fetching a trigger capability and unregistering should close client channel", func(t *testing.T) {
		mtr := mustMockTrigger(t)
		tr, _, _, err := newCapabilityPlugin(t, mtr)
		require.NoError(t, err)

		ctr := tr.(capabilities.TriggerCapability)

		ch, err := ctr.RegisterTrigger(
			t.Context(),
			capabilities.TriggerRegistrationRequest{})
		require.NoError(t, err)

		// Wait for registration to complete
		<-mtr.registerCalls
		assert.NotNil(t, mtr.callback)

		err = ctr.UnregisterTrigger(
			t.Context(),
			capabilities.TriggerRegistrationRequest{})

		require.NoError(t, err)

		<-mtr.unregisterCalls

		_, isOpen := <-ch
		assert.False(t, isOpen)
	})

	t.Run("fetching a trigger capability and cancelling context does not close client channel", func(t *testing.T) {
		mtr := mustMockTrigger(t)
		tr, _, _, err := newCapabilityPlugin(t, mtr)
		require.NoError(t, err)

		ctr := tr.(capabilities.TriggerCapability)

		ctxWithCancel, cancel := context.WithCancel(t.Context())

		ch, err := ctr.RegisterTrigger(
			ctxWithCancel,
			capabilities.TriggerRegistrationRequest{})
		require.NoError(t, err)

		// Wait for registration to complete
		<-mtr.registerCalls
		assert.NotNil(t, mtr.callback)

		// cancel originating context
		cancel()

		// send response on stream
		mtr.callback <- capabilities.TriggerResponse{
			Event: capabilities.TriggerEvent{
				ID: "test-event",
			},
		}
		gotTrigger, isOpen := <-ch
		assert.True(t, isOpen)
		assert.Equal(t, "test-event", gotTrigger.Event.ID)

		// call unregister to unregister trigger and close stream
		err = ctr.UnregisterTrigger(t.Context(), capabilities.TriggerRegistrationRequest{})
		require.NoError(t, err)

		<-mtr.unregisterCalls

		_, isOpen = <-ch
		assert.False(t, isOpen)
		assert.Nil(t, mtr.callback)
	})

	t.Run("fetching a trigger capability and calling Info", func(t *testing.T) {
		mtr := mustMockTrigger(t)
		tr, _, _, err := newCapabilityPlugin(t, mtr)
		require.NoError(t, err)

		gotInfo, err := tr.Info(t.Context())
		require.NoError(t, err)

		expectedInfo, err := mtr.Info(t.Context())
		require.NoError(t, err)
		assert.Equal(t, expectedInfo, gotInfo)
	})

	t.Run("fetching an action capability, and (un)registering it", func(t *testing.T) {
		ma := mustMockExecutable(t, capabilities.CapabilityTypeAction)
		c, _, _, err := newCapabilityPlugin(t, ma)
		require.NoError(t, err)

		act := c.(capabilities.ExecutableCapability)

		vmap, err := values.NewMap(map[string]any{"foo": "bar"})
		require.NoError(t, err)
		expectedRequest := capabilities.RegisterToWorkflowRequest{
			Config: vmap,
		}
		err = act.RegisterToWorkflow(
			t.Context(),
			expectedRequest)
		require.NoError(t, err)

		assert.Equal(t, expectedRequest, ma.regRequest)

		expectedUnrRequest := capabilities.UnregisterFromWorkflowRequest{
			Config: vmap,
		}
		err = act.UnregisterFromWorkflow(
			t.Context(),
			expectedUnrRequest)
		require.NoError(t, err)

		assert.Equal(t, expectedUnrRequest, ma.unregRequest)
	})

	t.Run("fetching an action capability, and executing it", func(t *testing.T) {
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
			Metadata: capabilities.ResponseMetadata{
				Metering: []capabilities.MeteringNodeDetail{},
			},
		}

		ma.callback <- expectedResp

		resp, err := c.(capabilities.ExecutableCapability).Execute(
			t.Context(),
			expectedRequest)
		require.NoError(t, err)

		assert.Equal(t, expectedResp, resp)
	})

	t.Run("fetching an action capability, and executing it with reportable error", func(t *testing.T) {
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

		ma.responseError = caperrors.NewPublicSystemError(errors.New("bang"), caperrors.DeadlineExceeded)

		_, err = c.(capabilities.ActionCapability).Execute(
			t.Context(),
			expectedRequest)
		require.Error(t, err)
		capErr, ok := errors.AsType[caperrors.Error](err)
		require.True(t, ok)
		require.Equal(t, "[4]DeadlineExceeded: bang", capErr.Error())
		require.Equal(t, caperrors.DeadlineExceeded, capErr.Code())
		require.Equal(t, caperrors.VisibilityPublic, capErr.Visibility())
		require.Equal(t, caperrors.OriginSystem, capErr.Origin())
	})

	t.Run("fetching an action capability, and executing it with reportable user error", func(t *testing.T) {
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

		ma.responseError = caperrors.NewPublicUserError(errors.New("bang"), caperrors.NotFound)

		_, err = c.(capabilities.ActionCapability).Execute(
			t.Context(),
			expectedRequest)
		require.Error(t, err)
		capErr, ok := errors.AsType[caperrors.Error](err)
		require.True(t, ok)
		require.Equal(t, "[5]NotFound: bang", capErr.Error())
		require.Equal(t, caperrors.NotFound, capErr.Code())
		require.Equal(t, caperrors.VisibilityPublic, capErr.Visibility())
		require.Equal(t, caperrors.OriginUser, capErr.Origin())
	})

	t.Run("fetching an action capability, and executing it with private system error", func(t *testing.T) {
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

		ma.responseError = caperrors.NewPrivateSystemError(errors.New("bang"), caperrors.DeadlineExceeded)

		_, err = c.(capabilities.ActionCapability).Execute(
			t.Context(),
			expectedRequest)
		require.Error(t, err)
		capErr, ok := errors.AsType[caperrors.Error](err)
		require.True(t, ok)
		require.Equal(t, "[4]DeadlineExceeded: bang", capErr.Error())
		require.Equal(t, caperrors.DeadlineExceeded, capErr.Code())
		require.Equal(t, caperrors.VisibilityPrivate, capErr.Visibility())
		require.Equal(t, caperrors.OriginSystem, capErr.Origin())
	})

	// This will only happen a local capability has not had it's API migrated to always return capability.Error
	t.Run("fetching an action capability, and executing it without capability error", func(t *testing.T) {
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
			t.Context(),
			expectedRequest)
		require.Error(t, err)
		capErr, ok := errors.AsType[caperrors.Error](err)
		require.True(t, ok)
		require.Equal(t, "[2]Unknown: Private:bang", capErr.Error())
		require.Equal(t, caperrors.Unknown, capErr.Code())
		require.Equal(t, caperrors.VisibilityPrivate, capErr.Visibility())
		require.Equal(t, caperrors.OriginSystem, capErr.Origin())
	})

	t.Run("fetching an action capability, and closing it", func(t *testing.T) {
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
		_, err = c.(capabilities.ExecutableCapability).Execute(
			t.Context(),
			expectedRequest)
		require.NoError(t, err)
	})

	t.Run("calling execute should be synchronous", func(t *testing.T) {
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

		_, err = c.(capabilities.ExecutableCapability).Execute(
			t.Context(),
			expectedRequest)
		require.NoError(t, err)

		assert.True(t, ma.executeCalled)
	})

	t.Run("Execute wraps transport error when client connection is closed before call", func(t *testing.T) {
		ma := mustMockExecutable(t, capabilities.CapabilityTypeAction)
		c, client, _, err := newCapabilityPlugin(t, ma)
		require.NoError(t, err)

		// Close the underlying client connection so c.grpc.Execute fails at the
		// initial gRPC call site.
		require.NoError(t, client.Close())

		cmap, err := values.NewMap(map[string]any{"foo": "bar"})
		require.NoError(t, err)
		imap, err := values.NewMap(map[string]any{"bar": "baz"})
		require.NoError(t, err)

		_, err = c.(capabilities.ExecutableCapability).Execute(
			t.Context(),
			capabilities.CapabilityRequest{Config: cmap, Inputs: imap})
		require.Error(t, err)

		capErr, ok := errors.AsType[caperrors.Error](err)
		require.True(t, ok, "expected caperrors.Error, got %T: %v", err, err)
		require.Equal(t, caperrors.Unavailable, capErr.Code())
		require.Equal(t, caperrors.VisibilityPublic, capErr.Visibility())
		require.Equal(t, caperrors.OriginSystem, capErr.Origin())
		require.Contains(t, capErr.Error(), "[14]Unavailable: error executing capability request:") // capping the error as it might change in CI
	})

	t.Run("Execute wraps responseStream.Recv error when stream breaks mid-flight", func(t *testing.T) {
		ma := mustMockExecutable(t, capabilities.CapabilityTypeAction)
		ma.executeEntered = make(chan struct{}, 1)
		// Do NOT preload ma.callback — the server-side impl.Execute will block
		// on `<-m.callback` after signalling executeEntered, leaving the client
		// parked in responseStream.Recv() so we can break the stream below.
		c, _, server, err := newCapabilityPlugin(t, ma)
		require.NoError(t, err)

		cmap, err := values.NewMap(map[string]any{"foo": "bar"})
		require.NoError(t, err)
		imap, err := values.NewMap(map[string]any{"bar": "baz"})
		require.NoError(t, err)

		done := make(chan error, 1)
		go func() {
			_, execErr := c.(capabilities.ExecutableCapability).Execute(
				t.Context(),
				capabilities.CapabilityRequest{Config: cmap, Inputs: imap})
			done <- execErr
		}()

		select {
		case <-ma.executeEntered:
		case <-time.After(5 * time.Second):
			t.Fatal("server-side Execute was never invoked")
		}

		// Stream is established and client is blocked in Recv(); tear the
		// server down so Recv() returns an error.
		server.Stop()

		var execErr error
		select {
		case execErr = <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Execute did not return after server stop")
		}
		require.Error(t, execErr)

		capErr, ok := errors.AsType[caperrors.Error](execErr)
		require.True(t, ok, "expected caperrors.Error, got %T: %v", execErr, execErr)
		require.Equal(t, caperrors.Unavailable, capErr.Code())
		require.Equal(t, caperrors.VisibilityPublic, capErr.Visibility())
		require.Equal(t, caperrors.OriginSystem, capErr.Origin())
		require.Contains(t, capErr.Error(), "[14]Unavailable: error waiting for response message: rpc error:") // capping the error as it might change in CI
	})

	t.Run("Execute wraps CapabilityResponseFromProto unmarshal error as caperrors.Error", func(t *testing.T) {
		stopCh := make(chan struct{})
		pluginName := "malformed"
		cl, _ := plugin.TestPluginGRPCConn(
			t,
			false,
			map[string]plugin.Plugin{
				pluginName: &malformedExecutablePlugin{
					brokerCfg: net.BrokerConfig{
						StopCh: stopCh,
						Logger: logger.Test(t),
					},
				},
			},
		)

		raw, err := cl.Dispense(pluginName)
		require.NoError(t, err)

		cmap, err := values.NewMap(map[string]any{"foo": "bar"})
		require.NoError(t, err)
		imap, err := values.NewMap(map[string]any{"bar": "baz"})
		require.NoError(t, err)

		_, err = raw.(capabilities.ExecutableCapability).Execute(
			t.Context(),
			capabilities.CapabilityRequest{Config: cmap, Inputs: imap})
		require.Error(t, err)

		capErr, ok := errors.AsType[caperrors.Error](err)
		require.True(t, ok, "expected caperrors.Error, got %T: %v", err, err)
		require.Equal(t, caperrors.Internal, capErr.Code())
		require.Equal(t, caperrors.VisibilityPublic, capErr.Visibility())
		require.Equal(t, caperrors.OriginSystem, capErr.Origin())
		require.Contains(t, capErr.Error(), "[13]Internal: could not unmarshal response: invalid config digest length: expected 32 bytes, got 1")
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
