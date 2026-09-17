package remote

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

// bufnet stands in for a network entirely in memory: listen hands out a listener
// under a synthetic address, and connect dials whichever listener holds it. This
// keeps the test free of ports and sockets - and, more to the point, free of a
// go-plugin broker, which is what it is here to prove is no longer required.
type bufnet struct {
	mu  sync.Mutex
	n   int
	lis map[string]*bufconn.Listener
}

func newBufnet() *bufnet { return &bufnet{lis: map[string]*bufconn.Listener{}} }

func (b *bufnet) listen() (net.Listener, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.n++
	addr := fmt.Sprintf("bufnet-%d", b.n)
	l := bufconn.Listen(1024 * 1024)
	b.lis[addr] = l
	return addrListener{Listener: l, addr: addr}, nil
}

func (b *bufnet) connect(target string) (*grpc.ClientConn, error) {
	b.mu.Lock()
	l, ok := b.lis[target]
	b.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("no listener at %q", target)
	}
	return grpc.NewClient("passthrough:///"+target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return l.DialContext(ctx)
		}),
	)
}

// addrListener reports the synthetic address bufnet assigned, since a bufconn
// listener's own Addr is a constant and could not tell two capabilities apart.
type addrListener struct {
	net.Listener
	addr string
}

func (l addrListener) Addr() net.Addr { return bufAddr(l.addr) }

type bufAddr string

func (a bufAddr) Network() string { return "bufnet" }
func (a bufAddr) String() string  { return string(a) }

type testTrigger struct {
	info capabilities.CapabilityInfo
	ch   chan capabilities.TriggerResponse

	mu         sync.Mutex
	registered int
	acked      []string
}

func (c *testTrigger) Info(context.Context) (capabilities.CapabilityInfo, error) {
	return c.info, nil
}

func (c *testTrigger) RegisterTrigger(context.Context, capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	c.mu.Lock()
	c.registered++
	c.mu.Unlock()
	return c.ch, nil
}

func (c *testTrigger) UnregisterTrigger(context.Context, capabilities.TriggerRegistrationRequest) error {
	return nil
}

func (c *testTrigger) AckEvent(_ context.Context, _, eventID, _ string) error {
	c.mu.Lock()
	c.acked = append(c.acked, eventID)
	c.mu.Unlock()
	return nil
}

// newRegistryOverBufnet serves a local registry over plain gRPC and returns a client
// for it. No broker is involved on either side.
func newRegistryOverBufnet(t *testing.T) (registry.CapabilitiesRegistry, *registry.Registry) {
	t.Helper()
	lggr := logger.Test(t)
	bn := newBufnet()
	transport, err := NewListenerTransport(ListenerTransportParams{
		Lggr:    lggr,
		Listen:  bn.listen,
		Connect: bn.connect,
	})
	require.NoError(t, err)

	local := registry.NewRegistry(lggr)
	lis, err := bn.listen()
	require.NoError(t, err)

	srv := grpc.NewServer()
	RegisterCapabilitiesRegistryServer(srv, lggr, local, transport)
	var wg sync.WaitGroup
	wg.Go(func() { _ = srv.Serve(lis) })
	t.Cleanup(func() {
		srv.Stop()
		wg.Wait()
	})

	cc, err := bn.connect(lis.Addr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = cc.Close() })

	return NewCapabilitiesRegistryClient(lggr, cc, transport), local
}

func TestListenerTransport_addAndResolveTrigger(t *testing.T) {
	client, local := newRegistryOverBufnet(t)
	ctx := t.Context()

	trigger := &testTrigger{
		info: capabilities.CapabilityInfo{
			ID:             "trigger-one@1.0.0",
			CapabilityType: capabilities.CapabilityTypeTrigger,
			Description:    "a trigger",
		},
		ch: make(chan capabilities.TriggerResponse, 1),
	}

	// Add publishes the capability on this side and hands the server a target to
	// dial back on - the direction that used to need a broker connection id.
	require.NoError(t, client.Add(ctx, trigger))

	// The server-side registry now holds a wrapper reaching back over gRPC.
	stored, err := local.GetTrigger(ctx, "trigger-one@1.0.0")
	require.NoError(t, err)
	info, err := stored.Info(ctx)
	require.NoError(t, err)
	assert.Equal(t, "trigger-one@1.0.0", info.ID)
	assert.Equal(t, capabilities.CapabilityTypeTrigger, info.CapabilityType)

	// And resolving it back through the client works in the other direction.
	resolved, err := client.GetTrigger(ctx, "trigger-one@1.0.0")
	require.NoError(t, err)
	resolvedInfo, err := resolved.Info(ctx)
	require.NoError(t, err)
	assert.Equal(t, "trigger-one@1.0.0", resolvedInfo.ID)
}

func TestListenerTransport_triggerEventsFlowBack(t *testing.T) {
	client, _ := newRegistryOverBufnet(t)
	ctx := t.Context()

	trigger := &testTrigger{
		info: capabilities.CapabilityInfo{
			ID:             "trigger-two@1.0.0",
			CapabilityType: capabilities.CapabilityTypeTrigger,
		},
		ch: make(chan capabilities.TriggerResponse, 1),
	}
	require.NoError(t, client.Add(ctx, trigger))

	resolved, err := client.GetTrigger(ctx, "trigger-two@1.0.0")
	require.NoError(t, err)

	events, err := resolved.RegisterTrigger(ctx, capabilities.TriggerRegistrationRequest{TriggerID: "t1"})
	require.NoError(t, err)

	trigger.ch <- capabilities.TriggerResponse{
		Event: capabilities.TriggerEvent{TriggerType: "trigger-two", ID: "event-1"},
	}

	got := <-events
	require.NoError(t, got.Err)
	assert.Equal(t, "event-1", got.Event.ID)
}

func TestListenerTransport_list(t *testing.T) {
	client, _ := newRegistryOverBufnet(t)
	ctx := t.Context()

	for _, id := range []string{"trigger-a@1.0.0", "trigger-b@1.0.0"} {
		require.NoError(t, client.Add(ctx, &testTrigger{
			info: capabilities.CapabilityInfo{ID: id, CapabilityType: capabilities.CapabilityTypeTrigger},
			ch:   make(chan capabilities.TriggerResponse, 1),
		}))
	}

	listed, err := client.List(ctx)
	require.NoError(t, err)
	require.Len(t, listed, 2)

	var ids []string
	for _, c := range listed {
		info, err := c.Info(ctx)
		require.NoError(t, err)
		ids = append(ids, info.ID)
	}
	assert.ElementsMatch(t, []string{"trigger-a@1.0.0", "trigger-b@1.0.0"}, ids)
}

func TestListenerTransport_rejectsLegacyHandle(t *testing.T) {
	// A peer that names a capability only by the numeric handle wants a broker this
	// transport does not have, and must be told so rather than silently dialling
	// something else.
	transport, err := NewListenerTransport(ListenerTransportParams{
		Lggr:   logger.Test(t),
		Listen: newBufnet().listen,
	})
	require.NoError(t, err)
	conn := transport.Dial("Capability", func(context.Context) (Locator, error) {
		return Locator{LegacyHandle: 7}, nil
	})
	err = conn.Invoke(t.Context(), "/whatever", nil, nil)
	require.ErrorIs(t, err, ErrNoLegacyHandle)
}

type testAction struct {
	info capabilities.CapabilityInfo

	mu       sync.Mutex
	gotInput string
}

func (c *testAction) Info(context.Context) (capabilities.CapabilityInfo, error) {
	return c.info, nil
}

func (c *testAction) RegisterToWorkflow(context.Context, capabilities.RegisterToWorkflowRequest) error {
	return nil
}

func (c *testAction) UnregisterFromWorkflow(context.Context, capabilities.UnregisterFromWorkflowRequest) error {
	return nil
}

func (c *testAction) Execute(_ context.Context, req capabilities.CapabilityRequest) (capabilities.CapabilityResponse, error) {
	in, err := req.Inputs.Underlying["in"].Unwrap()
	if err != nil {
		return capabilities.CapabilityResponse{}, err
	}
	c.mu.Lock()
	c.gotInput = in.(string)
	c.mu.Unlock()

	out, err := values.NewMap(map[string]any{"out": "echo:" + in.(string)})
	if err != nil {
		return capabilities.CapabilityResponse{}, err
	}
	return capabilities.CapabilityResponse{Value: out}, nil
}

func TestListenerTransport_executeAction(t *testing.T) {
	client, _ := newRegistryOverBufnet(t)
	ctx := t.Context()

	action := &testAction{info: capabilities.CapabilityInfo{
		ID:             "action-one@1.0.0",
		CapabilityType: capabilities.CapabilityTypeAction,
	}}
	require.NoError(t, client.Add(ctx, action))

	resolved, err := client.GetExecutable(ctx, "action-one@1.0.0")
	require.NoError(t, err)

	inputs, err := values.NewMap(map[string]any{"in": "hello"})
	require.NoError(t, err)

	resp, err := resolved.Execute(ctx, capabilities.CapabilityRequest{Inputs: inputs})
	require.NoError(t, err)

	got, err := resp.Value.Underlying["out"].Unwrap()
	require.NoError(t, err)
	assert.Equal(t, "echo:hello", got)

	action.mu.Lock()
	defer action.mu.Unlock()
	assert.Equal(t, "hello", action.gotInput, "the request reached the capability itself")
}

func TestListenerTransport_getBaseCapability(t *testing.T) {
	client, _ := newRegistryOverBufnet(t)
	ctx := t.Context()

	require.NoError(t, client.Add(ctx, &testAction{info: capabilities.CapabilityInfo{
		ID:             "action-two@1.0.0",
		CapabilityType: capabilities.CapabilityTypeAction,
	}}))

	got, err := client.Get(ctx, "action-two@1.0.0")
	require.NoError(t, err)
	info, err := got.Info(ctx)
	require.NoError(t, err)
	assert.Equal(t, "action-two@1.0.0", info.ID)
}
