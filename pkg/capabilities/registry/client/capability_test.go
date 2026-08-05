package client

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	caperrors "github.com/smartcontractkit/chainlink-common/pkg/capabilities/errors"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

// serveCapability starts a gRPC server for impl over an in-memory listener and
// returns a connection to it.
//
// bufconn rather than a loopback port: these tests exercise the capability wire
// protocol, not address resolution, and an in-memory pipe needs no port and no
// listen permission. Address handling is covered separately in client_test.go,
// which drives the registry client's dial path with named fake addresses.
func serveCapability(t *testing.T, impl capabilities.BaseCapability, capType capabilities.CapabilityType) *grpc.ClientConn {
	t.Helper()

	srv := grpc.NewServer()
	require.NoError(t, RegisterCapability(logger.Test(t), srv, impl, capType))

	lis := bufconn.Listen(1 << 20)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	t.Cleanup(func() { _ = conn.Close() })

	return conn
}

// --- fakes ---

type fakeExecutable struct {
	info capabilities.CapabilityInfo

	executeErr error
	response   capabilities.CapabilityResponse

	registeredTo   []capabilities.RegistrationMetadata
	unregisteredTo []capabilities.RegistrationMetadata
	lastRequest    capabilities.CapabilityRequest
}

func (f *fakeExecutable) Info(context.Context) (capabilities.CapabilityInfo, error) {
	return f.info, nil
}

func (f *fakeExecutable) Execute(_ context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	f.lastRequest = req
	if f.executeErr != nil {
		return capabilities.CapabilityResponse{}, f.executeErr
	}
	return f.response, nil
}

func (f *fakeExecutable) RegisterToWorkflow(_ context.Context, req capabilities.RegisterToWorkflowRequest) error {
	f.registeredTo = append(f.registeredTo, req.Metadata)
	return nil
}

func (f *fakeExecutable) UnregisterFromWorkflow(_ context.Context, req capabilities.UnregisterFromWorkflowRequest) error {
	f.unregisteredTo = append(f.unregisteredTo, req.Metadata)
	return nil
}

type fakeTrigger struct {
	info capabilities.CapabilityInfo

	registerErr error
	events      []capabilities.TriggerResponse

	unregistered chan capabilities.TriggerRegistrationRequest
	acked        chan [3]string
}

func newFakeTrigger(info capabilities.CapabilityInfo) *fakeTrigger {
	return &fakeTrigger{
		info:         info,
		unregistered: make(chan capabilities.TriggerRegistrationRequest, 4),
		acked:        make(chan [3]string, 4),
	}
}

func (f *fakeTrigger) Info(context.Context) (capabilities.CapabilityInfo, error) {
	return f.info, nil
}

func (f *fakeTrigger) RegisterTrigger(_ context.Context, _ capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	if f.registerErr != nil {
		return nil, f.registerErr
	}
	ch := make(chan capabilities.TriggerResponse, len(f.events))
	for _, e := range f.events {
		ch <- e
	}
	close(ch)
	return ch, nil
}

func (f *fakeTrigger) UnregisterTrigger(_ context.Context, req capabilities.TriggerRegistrationRequest) error {
	select {
	case f.unregistered <- req:
	default:
	}
	return nil
}

func (f *fakeTrigger) AckEvent(_ context.Context, triggerID, eventID, method string) error {
	select {
	case f.acked <- [3]string{triggerID, eventID, method}:
	default:
	}
	return nil
}

// --- tests ---

func TestCapabilityRoundTrip_InfoAndExecute(t *testing.T) {
	ctx := context.Background()

	outputs, err := values.NewMap(map[string]any{"out": int64(7)})
	require.NoError(t, err)

	impl := &fakeExecutable{
		info:     capabilities.MustNewCapabilityInfo("write-chain@1.0.0", capabilities.CapabilityTypeTarget, "a target"),
		response: capabilities.CapabilityResponse{Value: outputs},
	}
	conn := serveCapability(t, impl, capabilities.CapabilityTypeTarget)

	base := newBaseCapabilityClient(conn)
	gotInfo, err := base.Info(ctx)
	require.NoError(t, err)
	assert.Equal(t, "write-chain@1.0.0", gotInfo.ID)
	assert.Equal(t, capabilities.CapabilityTypeTarget, gotInfo.CapabilityType)
	assert.Equal(t, "a target", gotInfo.Description)

	exec := newExecutableClient(conn)
	resp, err := exec.Execute(ctx, capabilities.CapabilityRequest{
		Metadata: capabilities.RequestMetadata{WorkflowID: "wf-1"},
		Method:   "some.method",
	})
	require.NoError(t, err)
	assert.Equal(t, outputs, resp.Value)
	assert.Equal(t, "wf-1", impl.lastRequest.Metadata.WorkflowID)
	assert.Equal(t, "some.method", impl.lastRequest.Method)
}

func TestCapabilityRoundTrip_ExecutePreservesCapabilityError(t *testing.T) {
	ctx := context.Background()

	// A classified capability error must survive the round trip, because callers
	// (e.g. the confidential-compute retry loop) branch on Origin to decide
	// whether retrying can possibly help.
	userErr := caperrors.NewPublicUserError(errors.New("bad key"), caperrors.InvalidArgument)
	impl := &fakeExecutable{
		info:       capabilities.MustNewCapabilityInfo("vault@1.0.0", capabilities.CapabilityTypeAction, "vault"),
		executeErr: userErr,
	}
	conn := serveCapability(t, impl, capabilities.CapabilityTypeAction)

	_, err := newExecutableClient(conn).Execute(ctx, capabilities.CapabilityRequest{})
	require.Error(t, err)

	var capErr caperrors.Error
	require.ErrorAs(t, err, &capErr)
	assert.Equal(t, caperrors.OriginUser, capErr.Origin())
	assert.Contains(t, err.Error(), "bad key")
}

func TestCapabilityRoundTrip_ExecuteHidesUnclassifiedError(t *testing.T) {
	ctx := context.Background()

	impl := &fakeExecutable{
		info:       capabilities.MustNewCapabilityInfo("act@1.0.0", capabilities.CapabilityTypeAction, "act"),
		executeErr: errors.New("connection string user:hunter2@db"),
	}
	conn := serveCapability(t, impl, capabilities.CapabilityTypeAction)

	_, err := newExecutableClient(conn).Execute(ctx, capabilities.CapabilityRequest{})
	require.Error(t, err)

	// Unclassified errors are marked private so they cannot be reported onward
	// verbatim; the marker is what downstream checks, so assert it is present.
	assert.Contains(t, err.Error(), caperrors.PrePendPrivateVisibilityIdentifier(""))
}

func TestCapabilityRoundTrip_RegisterAndUnregisterToWorkflow(t *testing.T) {
	ctx := context.Background()

	impl := &fakeExecutable{
		info: capabilities.MustNewCapabilityInfo("act@1.0.0", capabilities.CapabilityTypeAction, "act"),
	}
	conn := serveCapability(t, impl, capabilities.CapabilityTypeAction)
	exec := newExecutableClient(conn)

	md := capabilities.RegistrationMetadata{WorkflowID: "wf-1", ReferenceID: "ref-1"}

	// Nil config must not panic: callers routinely omit it.
	require.NoError(t, exec.RegisterToWorkflow(ctx, capabilities.RegisterToWorkflowRequest{Metadata: md}))
	require.NoError(t, exec.UnregisterFromWorkflow(ctx, capabilities.UnregisterFromWorkflowRequest{Metadata: md}))

	require.Len(t, impl.registeredTo, 1)
	assert.Equal(t, "wf-1", impl.registeredTo[0].WorkflowID)
	assert.Equal(t, "ref-1", impl.registeredTo[0].ReferenceID)
	require.Len(t, impl.unregisteredTo, 1)
	assert.Equal(t, "wf-1", impl.unregisteredTo[0].WorkflowID)
}

func TestCapabilityRoundTrip_RegisterTriggerStreamsEvents(t *testing.T) {
	ctx := context.Background()

	outputs, err := values.NewMap(map[string]any{"n": int64(1)})
	require.NoError(t, err)

	impl := newFakeTrigger(capabilities.MustNewCapabilityInfo("cron@1.0.0", capabilities.CapabilityTypeTrigger, "cron"))
	impl.events = []capabilities.TriggerResponse{
		{Event: capabilities.TriggerEvent{TriggerType: "cron@1.0.0", ID: "e1", Outputs: outputs}},
		{Event: capabilities.TriggerEvent{TriggerType: "cron@1.0.0", ID: "e2", Outputs: outputs}},
	}
	conn := serveCapability(t, impl, capabilities.CapabilityTypeTrigger)

	trig := newTriggerExecutableClient(logger.Test(t), conn)
	req := capabilities.TriggerRegistrationRequest{TriggerID: "trig-1"}

	ch, err := trig.RegisterTrigger(ctx, req)
	require.NoError(t, err)

	var ids []string
	for i := 0; i < 2; i++ {
		select {
		case resp, ok := <-ch:
			require.True(t, ok, "channel closed early")
			require.NoError(t, resp.Err)
			ids = append(ids, resp.Event.ID)
		case <-time.After(10 * time.Second):
			t.Fatal("timed out waiting for trigger event")
		}
	}
	assert.Equal(t, []string{"e1", "e2"}, ids)

	require.NoError(t, trig.UnregisterTrigger(ctx, req))

	select {
	case got := <-impl.unregistered:
		assert.Equal(t, "trig-1", got.TriggerID)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for unregister")
	}
}

func TestCapabilityRoundTrip_RegisterTriggerSurfacesRegistrationError(t *testing.T) {
	ctx := context.Background()

	// Registration failure must come back from RegisterTrigger itself, not as a
	// silently empty stream, or the caller would later try to unregister a
	// trigger that never existed.
	impl := newFakeTrigger(capabilities.MustNewCapabilityInfo("cron@1.0.0", capabilities.CapabilityTypeTrigger, "cron"))
	impl.registerErr = caperrors.NewPublicUserError(errors.New("bad schedule"), caperrors.InvalidArgument)
	conn := serveCapability(t, impl, capabilities.CapabilityTypeTrigger)

	trig := newTriggerExecutableClient(logger.Test(t), conn)
	ch, err := trig.RegisterTrigger(ctx, capabilities.TriggerRegistrationRequest{TriggerID: "trig-1"})
	require.Error(t, err)
	assert.Nil(t, ch)
	assert.Contains(t, err.Error(), "bad schedule")
}

func TestCapabilityRoundTrip_AckEvent(t *testing.T) {
	ctx := context.Background()

	impl := newFakeTrigger(capabilities.MustNewCapabilityInfo("cron@1.0.0", capabilities.CapabilityTypeTrigger, "cron"))
	conn := serveCapability(t, impl, capabilities.CapabilityTypeTrigger)

	require.NoError(t, newTriggerExecutableClient(logger.Test(t), conn).AckEvent(ctx, "trig-1", "e1", "m"))

	select {
	case got := <-impl.acked:
		assert.Equal(t, [3]string{"trig-1", "e1", "m"}, got)
	case <-time.After(10 * time.Second):
		t.Fatal("timed out waiting for ack")
	}
}

func TestRegisterCapability_TypeMismatch(t *testing.T) {
	// An executable value declared as a trigger must be rejected at registration
	// rather than producing a capability that fails only when first called.
	impl := &fakeExecutable{
		info: capabilities.MustNewCapabilityInfo("act@1.0.0", capabilities.CapabilityTypeAction, "act"),
	}
	err := RegisterCapability(logger.Test(t), grpc.NewServer(), impl, capabilities.CapabilityTypeTrigger)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TriggerCapability")
}

func TestRegisterCapability_UnknownTypeRegistersBaseOnly(t *testing.T) {
	ctx := context.Background()

	impl := &fakeExecutable{
		info: capabilities.MustNewCapabilityInfo("act@1.0.0", capabilities.CapabilityTypeAction, "act"),
	}
	conn := serveCapability(t, impl, capabilities.CapabilityTypeUnknown)

	// Base works...
	_, err := newBaseCapabilityClient(conn).Info(ctx)
	require.NoError(t, err)

	// ...Executable was never registered.
	_, err = newExecutableClient(conn).Execute(ctx, capabilities.CapabilityRequest{})
	require.Error(t, err)
}
