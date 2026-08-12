// Package client_test, not client: server now dials capabilities through this package's Wrap to
// back Add with a real value, so an internal test here (package client) importing server would
// cycle back to client.
package client_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	ragetypes "github.com/smartcontractkit/libocr/ragep2p/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/client"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/registrytest"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/server"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
)

// These tests drive the real registry and its real gRPC adapter over an in-memory
// listener. Only the on-chain metadata read is substituted, since a test cannot
// run a chain; everything else the client talks to is production code.

// newRegistry builds the registry under test. Add now dials whatever address it is given to
// resolve a real value, so the registry needs the same capability dial options the client tests
// pass it - book, when the test serves fakes through one; otherwise, just credentials, since
// Add's dial-out is a production requirement, not conditional on what the test's own client does.
func newRegistry(t *testing.T, book *registrytest.AddrBook) *server.Registry {
	t.Helper()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if book != nil {
		opts = append(opts, book.DialOption())
	}
	return server.New(logger.Test(t), opts...)
}

// newClient serves reg and returns a client for it, with capability dials
// resolved through book when one is given.
func newClient(t *testing.T, reg *server.Registry, book *registrytest.AddrBook) *client.Client {
	t.Helper()

	var opts []grpc.DialOption
	if book != nil {
		opts = append(opts, book.DialOption(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	return client.New(logger.Test(t), registrytest.Serve(t, reg), opts...)
}

// serveCapabilityAt registers impl in book under name and returns its address.
func serveCapabilityAt(t *testing.T, book *registrytest.AddrBook, name string,
	impl capabilities.BaseCapability, capType capabilities.CapabilityType) string {
	t.Helper()

	srv := grpc.NewServer()
	require.NoError(t, client.RegisterCapability(logger.Test(t), srv, impl, capType))
	return book.Serve(t, name, srv)
}

func peer(b byte) ragetypes.PeerID {
	var p ragetypes.PeerID
	p[0] = b
	return p
}

// fakeExecutable and fakeTrigger are local copies of client's own capability_test.go doubles: that
// file stays an internal test (package client) since it does not need server, and its unexported
// types are not reachable from here, an external test package.

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

// --- registration ---

func TestAddAt_CarriesCapabilityTypeVerbatim(t *testing.T) {
	ctx := context.Background()

	// The wire carries the capability's own type, so nothing is collapsed on the
	// way out and nothing has to be guessed back on the way in.
	for _, capType := range []capabilities.CapabilityType{
		capabilities.CapabilityTypeTrigger,
		capabilities.CapabilityTypeAction,
		capabilities.CapabilityTypeTarget,
		capabilities.CapabilityTypeConsensus,
		capabilities.CapabilityTypeCombined,
	} {
		t.Run(string(capType), func(t *testing.T) {
			book := registrytest.NewAddrBook()
			info := capabilities.MustNewCapabilityInfo("c@1.0.0", capType, "c")
			impl := &fakeCombined{
				fakeExecutable: &fakeExecutable{info: info},
				fakeTrigger:    newFakeTrigger(info),
			}
			addr := serveCapabilityAt(t, book, "c", impl, capType)

			reg := newRegistry(t, book)
			c := newClient(t, reg, book)

			require.NoError(t, c.AddAt(ctx, "c@1.0.0", capType, addr))

			got, err := reg.Get(ctx, "c@1.0.0")
			require.NoError(t, err)
			assert.Equal(t, capType, got.Type)
			assert.Equal(t, addr, got.URL)
		})
	}
}

func TestAddAt_RejectsUnknownType(t *testing.T) {
	ctx := context.Background()

	reg := newRegistry(t, nil)
	c := newClient(t, reg, nil)

	// Without a type the registry cannot know which services live at the address.
	require.Error(t, c.AddAt(ctx, "c@1.0.0", capabilities.CapabilityTypeUnknown, "addr:1"))
	assert.Empty(t, reg.List(ctx))
}

func TestAddAt_RejectsEmptyAddress(t *testing.T) {
	ctx := context.Background()

	reg := newRegistry(t, nil)
	c := newClient(t, reg, nil)

	require.Error(t, c.AddAt(ctx, "act@1.0.0", capabilities.CapabilityTypeAction, ""))
	assert.Empty(t, reg.List(ctx))
}

func TestClient_Remove(t *testing.T) {
	ctx := context.Background()

	book := registrytest.NewAddrBook()
	impl := &fakeExecutable{
		info: capabilities.MustNewCapabilityInfo("act@1.0.0", capabilities.CapabilityTypeAction, "act"),
	}
	addr := serveCapabilityAt(t, book, "cap-a", impl, capabilities.CapabilityTypeAction)

	reg := newRegistry(t, book)
	c := newClient(t, reg, book)

	require.NoError(t, c.AddAt(ctx, "act@1.0.0", capabilities.CapabilityTypeAction, addr))
	require.NoError(t, c.Remove(ctx, "act@1.0.0"))

	_, err := reg.Get(ctx, "act@1.0.0")
	require.Error(t, err)

	// Removing what is not there is an error, not a silent no-op.
	require.Error(t, c.Remove(ctx, "act@1.0.0"))
}

// --- resolution ---

func TestClient_GetExecutableDialsHandleAddress(t *testing.T) {
	ctx := context.Background()

	outputs, err := values.NewMap(map[string]any{"out": int64(3)})
	require.NoError(t, err)

	book := registrytest.NewAddrBook()
	impl := &fakeExecutable{
		info:     capabilities.MustNewCapabilityInfo("act@1.0.0", capabilities.CapabilityTypeAction, "act"),
		response: capabilities.CapabilityResponse{Value: outputs},
	}
	addr := serveCapabilityAt(t, book, "cap-a", impl, capabilities.CapabilityTypeAction)

	reg := newRegistry(t, book)
	c := newClient(t, reg, book)
	require.NoError(t, c.AddAt(ctx, "act@1.0.0", capabilities.CapabilityTypeAction, addr))

	exec, err := c.GetExecutable(ctx, "act@1.0.0")
	require.NoError(t, err)

	resp, err := exec.Execute(ctx, capabilities.CapabilityRequest{})
	require.NoError(t, err)
	assert.Equal(t, outputs, resp.Value)
}

func TestClient_ReusesConnectionPerAddress(t *testing.T) {
	ctx := context.Background()

	book := registrytest.NewAddrBook()
	impl := &fakeExecutable{
		info: capabilities.MustNewCapabilityInfo("act@1.0.0", capabilities.CapabilityTypeAction, "act"),
	}
	addr := serveCapabilityAt(t, book, "cap-a", impl, capabilities.CapabilityTypeAction)

	reg := newRegistry(t, book)
	c := newClient(t, reg, book)
	require.NoError(t, c.AddAt(ctx, "act@1.0.0", capabilities.CapabilityTypeAction, addr))
	// Add itself dials once, server-side, to resolve the handle into a real value.
	dialsBeforeLookups := book.DialCount("cap-a")

	// GetExecutable is on the per-invocation path for workflow steps, so a fresh
	// dial per call would add a handshake to every capability call.
	for i := 0; i < 5; i++ {
		exec, err := c.GetExecutable(ctx, "act@1.0.0")
		require.NoError(t, err)
		_, err = exec.Execute(ctx, capabilities.CapabilityRequest{})
		require.NoError(t, err)
	}

	assert.Equal(t, dialsBeforeLookups+1, book.DialCount("cap-a"), "expected one transport dial across repeated lookups")
}

func TestClient_GetTriggerRejectsExecutableOnlyCapability(t *testing.T) {
	ctx := context.Background()

	book := registrytest.NewAddrBook()
	impl := &fakeExecutable{
		info: capabilities.MustNewCapabilityInfo("act@1.0.0", capabilities.CapabilityTypeAction, "act"),
	}
	addr := serveCapabilityAt(t, book, "cap-a", impl, capabilities.CapabilityTypeAction)

	reg := newRegistry(t, book)
	c := newClient(t, reg, book)
	require.NoError(t, c.AddAt(ctx, "act@1.0.0", capabilities.CapabilityTypeAction, addr))

	// The registry gates this server-side, so the client never even gets a handle.
	_, err := c.GetTrigger(ctx, "act@1.0.0")
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestClient_CombinedCapabilityServesBothSurfaces(t *testing.T) {
	ctx := context.Background()

	book := registrytest.NewAddrBook()
	info := capabilities.MustNewCapabilityInfo("both@1.0.0", capabilities.CapabilityTypeCombined, "both")
	impl := &fakeCombined{
		fakeExecutable: &fakeExecutable{info: info},
		fakeTrigger:    newFakeTrigger(info),
	}
	addr := serveCapabilityAt(t, book, "cap-b", impl, capabilities.CapabilityTypeCombined)

	reg := newRegistry(t, book)
	c := newClient(t, reg, book)
	require.NoError(t, c.AddAt(ctx, "both@1.0.0", capabilities.CapabilityTypeCombined, addr))

	_, err := c.GetExecutable(ctx, "both@1.0.0")
	require.NoError(t, err)
	_, err = c.GetTrigger(ctx, "both@1.0.0")
	require.NoError(t, err)
}

func TestClient_GetUnknownCapability(t *testing.T) {
	ctx := context.Background()

	c := newClient(t, newRegistry(t, nil), nil)

	_, err := c.Get(ctx, "nope@1.0.0")
	require.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestClient_List(t *testing.T) {
	ctx := context.Background()

	book := registrytest.NewAddrBook()
	impl := &fakeExecutable{
		info: capabilities.MustNewCapabilityInfo("good@1.0.0", capabilities.CapabilityTypeAction, "good"),
	}
	addr := serveCapabilityAt(t, book, "cap-good", impl, capabilities.CapabilityTypeAction)

	reg := newRegistry(t, book)
	c := newClient(t, reg, book)
	require.NoError(t, c.AddAt(ctx, "good@1.0.0", capabilities.CapabilityTypeAction, addr))

	got, err := c.List(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)

	info, err := got[0].Info(ctx)
	require.NoError(t, err)
	assert.Equal(t, "good@1.0.0", info.ID)
}

func TestClient_MissingCredentialsFailsRatherThanFallingBackToInsecure(t *testing.T) {
	ctx := context.Background()

	book := registrytest.NewAddrBook()
	impl := &fakeExecutable{
		info: capabilities.MustNewCapabilityInfo("act@1.0.0", capabilities.CapabilityTypeAction, "act"),
	}
	addr := serveCapabilityAt(t, book, "cap-a", impl, capabilities.CapabilityTypeAction)

	// No capability dial options, so no transport credentials. gRPC must refuse the
	// dial: defaulting to insecure would silently make every capability call
	// unauthenticated and unencrypted for a caller that forgot to say.
	reg := newRegistry(t, book)
	c := newClient(t, reg, nil)
	require.NoError(t, c.AddAt(ctx, "act@1.0.0", capabilities.CapabilityTypeAction, addr))

	_, err := c.GetExecutable(ctx, "act@1.0.0")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "transport security",
		"expected gRPC's no-transport-security error, not a successful insecure dial")
}

// --- metadata ---

func TestClient_LocalNodeAndNodeByPeerID(t *testing.T) {
	ctx := context.Background()

	self := peer(0xab)
	node := capabilities.Node{
		PeerID:              &self,
		NodeOperatorID:      9,
		Signer:              [32]byte{0xcd},
		EncryptionPublicKey: [32]byte{0xef},
		WorkflowDON: capabilities.DON{
			ID: 1, Name: "wf", F: 1, ConfigVersion: 4,
			Members:          []ragetypes.PeerID{self},
			Families:         []string{"zone-a"},
			AcceptsWorkflows: true,
		},
		CapabilityDONs: []capabilities.DON{{ID: 2, Name: "cap", F: 2, ConfigVersion: 5}},
	}

	reg := newRegistry(t, nil)
	reg.SetMetadata(&registrytest.Metadata{
		Node:  &node,
		Nodes: map[ragetypes.PeerID]capabilities.Node{self: node},
	})
	c := newClient(t, reg, nil)

	// Round-trips through the real Go -> proto -> Go conversions on both sides.
	got, err := c.LocalNode(ctx)
	require.NoError(t, err)
	require.NotNil(t, got.PeerID)
	assert.Equal(t, self, *got.PeerID)
	assert.Equal(t, uint32(9), got.NodeOperatorID)
	assert.Equal(t, [32]byte{0xcd}, got.Signer)
	assert.Equal(t, [32]byte{0xef}, got.EncryptionPublicKey)
	assert.Equal(t, uint32(1), got.WorkflowDON.ID)
	assert.Equal(t, uint8(1), got.WorkflowDON.F)
	assert.Equal(t, uint32(4), got.WorkflowDON.ConfigVersion)
	assert.Equal(t, []string{"zone-a"}, got.WorkflowDON.Families)
	assert.True(t, got.WorkflowDON.AcceptsWorkflows)
	require.Len(t, got.WorkflowDON.Members, 1)
	assert.Equal(t, self, got.WorkflowDON.Members[0])
	require.Len(t, got.CapabilityDONs, 1)
	assert.Equal(t, uint32(2), got.CapabilityDONs[0].ID)

	byPeer, err := c.NodeByPeerID(ctx, self)
	require.NoError(t, err)
	assert.Equal(t, got.NodeOperatorID, byPeer.NodeOperatorID)
}

func TestClient_MetadataBeforeFirstSync(t *testing.T) {
	ctx := context.Background()

	// Metadata is a live read: before the registry has a source there is no answer
	// to cache or fabricate, and the caller has to see that.
	c := newClient(t, newRegistry(t, nil), nil)

	_, err := c.LocalNode(ctx)
	require.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))
}

func TestClient_ConfigForCapability(t *testing.T) {
	ctx := context.Background()

	defaultCfg, err := values.NewMap(map[string]any{"VaultPublicKey": "0xabc"})
	require.NoError(t, err)

	// Stored as the contract stores it: wire-encoded CapabilityConfig bytes. The
	// server parses them and the client decodes, so the on-chain encoding and the
	// wire message must agree end to end.
	wire, err := proto.Marshal(&capabilitiespb.CapabilityConfig{
		DefaultConfig: values.ProtoMap(defaultCfg),
		LocalOnly:     true,
	})
	require.NoError(t, err)

	reg := newRegistry(t, nil)
	reg.SetMetadata(&registrytest.Metadata{
		Configs: map[string][]byte{"vault@1.0.0": wire},
	})
	c := newClient(t, reg, nil)

	got, err := c.ConfigForCapability(ctx, "vault@1.0.0", 1)
	require.NoError(t, err)
	assert.Equal(t, defaultCfg, got.DefaultConfig)
	assert.True(t, got.LocalOnly)
}

func TestClient_ConfigForCapabilityRejectsGarbageBytes(t *testing.T) {
	ctx := context.Background()

	reg := newRegistry(t, nil)
	reg.SetMetadata(&registrytest.Metadata{
		Configs: map[string][]byte{"vault@1.0.0": []byte("not a proto")},
	})
	c := newClient(t, reg, nil)

	_, err := c.ConfigForCapability(ctx, "vault@1.0.0", 1)
	require.Error(t, err)
	assert.Equal(t, codes.Internal, status.Code(err))
}

func TestClient_DONsForCapabilityAndDONByID(t *testing.T) {
	ctx := context.Background()

	member := peer(0x01)
	don := capabilities.DON{
		ID: 7, Name: "cap-don", F: 1,
		Members:  []ragetypes.PeerID{member},
		Families: []string{"zone-b"},
	}

	reg := newRegistry(t, nil)
	reg.SetMetadata(&registrytest.Metadata{
		DONsForCap: map[string][]capabilities.DONWithNodes{
			"act@1.0.0": {{DON: don, Nodes: []capabilities.Node{{PeerID: &member}}}},
		},
		DONsByID: map[uint32]capabilities.DON{7: don},
	})
	c := newClient(t, reg, nil)

	dons, err := c.DONsForCapability(ctx, "act@1.0.0")
	require.NoError(t, err)
	require.Len(t, dons, 1)
	assert.Equal(t, uint32(7), dons[0].DON.ID)
	require.Len(t, dons[0].Nodes, 1)
	require.NotNil(t, dons[0].Nodes[0].PeerID)
	assert.Equal(t, member, *dons[0].Nodes[0].PeerID)

	got, err := c.DONByID(ctx, 7)
	require.NoError(t, err)
	assert.Equal(t, "cap-don", got.Name)
	assert.Equal(t, []string{"zone-b"}, got.Families)
}

// --- conversions ---

func TestCapabilityTypeConvertersRoundTrip(t *testing.T) {
	// Round-tripping must be lossless in both directions; the wire type is the
	// shared capabilities.pb enum, so there is no collapsing step to undo.
	for _, in := range []capabilities.CapabilityType{
		capabilities.CapabilityTypeTrigger,
		capabilities.CapabilityTypeAction,
		capabilities.CapabilityTypeConsensus,
		capabilities.CapabilityTypeTarget,
		capabilities.CapabilityTypeCombined,
		capabilities.CapabilityTypeUnknown,
	} {
		pbType, err := capabilitiespb.CapabilityTypeToProto(in)
		require.NoError(t, err)

		back, err := capabilitiespb.CapabilityTypeFromProto(pbType)
		require.NoError(t, err)
		assert.Equal(t, in, back)
	}
}

func TestDONFromProto_Nil(t *testing.T) {
	assert.Equal(t, capabilities.DON{}, client.DONFromProto(nil))
}

// fakeCombined serves both the executable and trigger surfaces.
type fakeCombined struct {
	*fakeExecutable
	*fakeTrigger
}

func (f *fakeCombined) Info(ctx context.Context) (capabilities.CapabilityInfo, error) {
	return f.fakeExecutable.Info(ctx)
}
