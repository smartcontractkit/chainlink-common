package client

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	ragetypes "github.com/smartcontractkit/libocr/ragep2p/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	capabilitiespb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/pb"
	registrypb "github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

type Client struct {
	lggr logger.Logger
	grpc registrypb.CapabilitiesRegistryClient

	// dialOpts are applied when dialing capability addresses. They do not apply
	// to the registry connection, which the caller supplies already dialed.
	dialOpts []grpc.DialOption

	mu    sync.Mutex
	conns map[string]*grpc.ClientConn
}

func New(lggr logger.Logger, cc grpc.ClientConnInterface, capabilityDialOpts ...grpc.DialOption) *Client {
	return &Client{
		lggr:     logger.Named(lggr, "CapabilitiesRegistryClient"),
		grpc:     registrypb.NewCapabilitiesRegistryClient(cc),
		dialOpts: capabilityDialOpts,
		conns:    map[string]*grpc.ClientConn{},
	}
}

// Close tears down every cached capability connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var errs []error
	for addr, conn := range c.conns {
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("closing conn to %s: %w", addr, err))
		}
	}
	c.conns = map[string]*grpc.ClientConn{}
	return errors.Join(errs...)
}

func (c *Client) connFor(addr string) (*grpc.ClientConn, error) {
	if addr == "" {
		return nil, errors.New("registry returned an empty callback address")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if conn, ok := c.conns[addr]; ok {
		return conn, nil
	}

	conn, err := grpc.NewClient(addr, c.dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial capability at %s: %w", addr, err)
	}
	c.conns[addr] = conn
	return conn, nil
}

// wrap turns a handle into the capability surface its type promises, and returns
// that type so callers can report what a capability actually is.
func (c *Client) wrap(h *registrypb.CapabilityHandle) (capabilities.BaseCapability, capabilities.CapabilityType, error) {
	capType, err := capabilitiespb.CapabilityTypeFromProto(h.GetType())
	if err != nil {
		return nil, capabilities.CapabilityTypeUnknown, fmt.Errorf("capability %s: %w", h.GetCapabilityId(), err)
	}

	conn, err := c.connFor(h.GetCallbackUrl())
	if err != nil {
		return nil, capType, err
	}

	wrapped, err := Wrap(c.lggr, conn, capType)
	if err != nil {
		return nil, capType, fmt.Errorf("capability %s: %w", h.GetCapabilityId(), err)
	}
	return wrapped, capType, nil
}

// Wrap builds the capability surface capType promises, served over conn.
//
// Exported so a registry holding a Handle (an ID, a type and a callback address) can dial that
// address itself and get back a real capabilities.BaseCapability, the same way this Client does for
// its own callers - server.Registry uses this to back Add with an actual value instead of only the
// address.
func Wrap(lggr logger.Logger, conn grpc.ClientConnInterface, capType capabilities.CapabilityType) (capabilities.BaseCapability, error) {
	base := newBaseCapabilityClient(conn)
	// Mirrors RegisterCapability on the serving side: same type, same services.
	switch capType {
	case capabilities.CapabilityTypeTrigger:
		return &triggerCapabilityClient{
			baseCapabilityClient:    base,
			triggerExecutableClient: newTriggerExecutableClient(lggr, conn),
		}, nil
	case capabilities.CapabilityTypeAction,
		capabilities.CapabilityTypeTarget,
		capabilities.CapabilityTypeConsensus:
		return &executableCapabilityClient{
			baseCapabilityClient: base,
			executableClient:     newExecutableClient(conn),
		}, nil
	case capabilities.CapabilityTypeCombined:
		return &combinedCapabilityClient{
			baseCapabilityClient:    base,
			executableClient:        newExecutableClient(conn),
			triggerExecutableClient: newTriggerExecutableClient(lggr, conn),
		}, nil
	default:
		return nil, fmt.Errorf("no usable capability type (%s)", capType)
	}
}

// lookupFn is the shape of the registry's three handle-resolving RPCs.
type lookupFn func(context.Context, *registrypb.GetRequest, ...grpc.CallOption) (*registrypb.CapabilityHandle, error)

// resolve looks a capability up and returns it as the surface T.
//
// The three Get* methods differ only in which RPC they call and which surface
// they demand, so the lookup, wrap and assert steps live here once.
func resolve[T capabilities.BaseCapability](ctx context.Context, c *Client, id, surface string, lookup lookupFn) (T, error) {
	var zero T

	h, err := lookup(ctx, &registrypb.GetRequest{CapabilityId: id})
	if err != nil {
		return zero, err
	}

	wrapped, capType, err := c.wrap(h)
	if err != nil {
		return zero, err
	}

	// The assertion and the declared type cannot disagree: wrap builds the surface
	// from that same type. It is kept as the mechanism so the two stay in step, but
	// the message reports the type, which is what a caller can act on.
	typed, ok := wrapped.(T)
	if !ok {
		return zero, fmt.Errorf("capability %s is a %s, so it does not serve the %s API", id, capType, surface)
	}
	return typed, nil
}

func (c *Client) Get(ctx context.Context, id string) (capabilities.BaseCapability, error) {
	return resolve[capabilities.BaseCapability](ctx, c, id, "base", c.grpc.Get)
}

func (c *Client) GetTrigger(ctx context.Context, id string) (capabilities.TriggerCapability, error) {
	return resolve[capabilities.TriggerCapability](ctx, c, id, "trigger", c.grpc.GetTrigger)
}

func (c *Client) GetExecutable(ctx context.Context, id string) (capabilities.ExecutableCapability, error) {
	return resolve[capabilities.ExecutableCapability](ctx, c, id, "executable", c.grpc.GetExecutable)
}

func (c *Client) List(ctx context.Context) ([]capabilities.BaseCapability, error) {
	reply, err := c.grpc.List(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	out := make([]capabilities.BaseCapability, 0, len(reply.GetHandles()))
	for _, h := range reply.GetHandles() {
		wrapped, _, err := c.wrap(h)
		if err != nil {
			// One unreachable capability must not blank the whole list; the
			// caller can still use the rest.
			c.lggr.Warnw("skipping capability that could not be wrapped",
				"capabilityID", h.GetCapabilityId(), "address", h.GetCallbackUrl(), "err", err)
			continue
		}
		out = append(out, wrapped)
	}
	return out, nil
}

// OCRConfig returns the OCR3 configuration a capability runs under, digest
// included.
//
// The digest is computed by the registry rather than here: it covers the
// configuration together with the chain and address the registry was read from,
// which only that process knows. See core.OCRConfigRegistry.
func (c *Client) OCRConfig(ctx context.Context, capabilityID string, donID uint32, key string) (ocrtypes.ContractConfig, error) {
	reply, err := c.grpc.OCRConfig(ctx, &registrypb.OCRConfigRequest{
		CapabilityId: capabilityID,
		DonId:        donID,
		Key:          key,
	})
	if err != nil {
		return ocrtypes.ContractConfig{}, fmt.Errorf("failed to read the OCR config of capability %s on DON %d: %w", capabilityID, donID, err)
	}

	digest, err := ocrtypes.BytesToConfigDigest(reply.GetConfigDigest())
	if err != nil {
		return ocrtypes.ContractConfig{}, fmt.Errorf("capability %s on DON %d has an unusable config digest: %w", capabilityID, donID, err)
	}

	return capabilitiespb.OCR3ConfigFromProto(reply.GetConfig(), digest)
}

// AddAt registers a capability served at addr.
func (c *Client) AddAt(ctx context.Context, id string, capType capabilities.CapabilityType, addr string) error {
	if capType == capabilities.CapabilityTypeUnknown {
		return fmt.Errorf("cannot register capability %s: no capability type, so the registry cannot know which services it serves", id)
	}
	pbType, err := capabilitiespb.CapabilityTypeToProto(capType)
	if err != nil {
		return err
	}
	if addr == "" {
		return fmt.Errorf("cannot register capability %s with an empty address", id)
	}
	if _, err := c.grpc.Add(ctx, &registrypb.AddRequest{
		CapabilityId: id,
		Type:         pbType,
		CallbackUrl:  addr,
	}); err != nil {
		return fmt.Errorf("failed to register capability %s at %s: %w", id, addr, err)
	}
	c.lggr.Infow("registered capability", "capabilityID", id, "address", addr)
	return nil
}

func (c *Client) Remove(ctx context.Context, id string) error {
	_, err := c.grpc.Remove(ctx, &registrypb.RemoveRequest{CapabilityId: id})
	return err
}

// --- metadata ---

func (c *Client) LocalNode(ctx context.Context) (capabilities.Node, error) {
	reply, err := c.grpc.LocalNode(ctx, &emptypb.Empty{})
	if err != nil {
		return capabilities.Node{}, err
	}
	return nodeFromProto(reply)
}

func (c *Client) NodeByPeerID(ctx context.Context, peerID ragetypes.PeerID) (capabilities.Node, error) {
	reply, err := c.grpc.NodeByPeerID(ctx, &registrypb.NodeRequest{PeerId: peerID[:]})
	if err != nil {
		return capabilities.Node{}, err
	}
	return nodeFromProto(reply)
}

// ConfigForCapability decodes the capability configuration the registry serves.
func (c *Client) ConfigForCapability(ctx context.Context, capabilityID string, donID uint32) (capabilities.CapabilityConfiguration, error) {
	reply, err := c.grpc.ConfigForCapability(ctx, &registrypb.ConfigForCapabilityRequest{
		CapabilityId: capabilityID,
		DonId:        donID,
	})
	if err != nil {
		return capabilities.CapabilityConfiguration{}, err
	}

	return capabilitiespb.CapabilityConfigFromProto(reply.GetCapabilityConfig())
}

func (c *Client) DONsForCapability(ctx context.Context, capabilityID string) ([]capabilities.DONWithNodes, error) {
	reply, err := c.grpc.DONsForCapability(ctx, &registrypb.DONsForCapabilityRequest{CapabilityId: capabilityID})
	if err != nil {
		return nil, err
	}

	out := make([]capabilities.DONWithNodes, 0, len(reply.GetDons()))
	for _, d := range reply.GetDons() {
		nodes := make([]capabilities.Node, 0, len(d.GetNodes()))
		for _, n := range d.GetNodes() {
			node, err := nodeFromProto(n)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, node)
		}
		out = append(out, capabilities.DONWithNodes{DON: DONFromProto(d.GetDon()), Nodes: nodes})
	}
	return out, nil
}

func (c *Client) DONByID(ctx context.Context, donID uint32) (capabilities.DON, error) {
	reply, err := c.grpc.DONByID(ctx, &registrypb.DONByIDRequest{DonId: donID})
	if err != nil {
		return capabilities.DON{}, err
	}
	return DONFromProto(reply.GetDon()), nil
}

// --- capability surfaces ---

// triggerCapabilityClient serves BaseCapability + TriggerExecutable.
type triggerCapabilityClient struct {
	*baseCapabilityClient
	*triggerExecutableClient
}

var _ capabilities.TriggerCapability = (*triggerCapabilityClient)(nil)

// executableCapabilityClient serves BaseCapability + Executable.
type executableCapabilityClient struct {
	*baseCapabilityClient
	*executableClient
}

var _ capabilities.ExecutableCapability = (*executableCapabilityClient)(nil)

// combinedCapabilityClient serves all three.
type combinedCapabilityClient struct {
	*baseCapabilityClient
	*executableClient
	*triggerExecutableClient
}

var _ capabilities.ExecutableAndTriggerCapability = (*combinedCapabilityClient)(nil)

// --- conversions ---

// DONFromProto converts a wire DON to the Go type.
func DONFromProto(d *registrypb.DON) capabilities.DON {
	if d == nil {
		return capabilities.DON{}
	}
	members := make([]ragetypes.PeerID, 0, len(d.GetMembers()))
	for _, m := range d.GetMembers() {
		var peerID ragetypes.PeerID
		copy(peerID[:], m)
		members = append(members, peerID)
	}
	return capabilities.DON{
		ID:               d.GetId(),
		Name:             d.GetName(),
		Members:          members,
		F:                uint8(d.GetF()),
		ConfigVersion:    d.GetConfigVersion(),
		Families:         d.GetFamilies(),
		Config:           d.GetConfig(),
		IsPublic:         d.GetIsPublic(),
		AcceptsWorkflows: d.GetAcceptsWorkflows(),
	}
}

func nodeFromProto(n *registrypb.NodeReply) (capabilities.Node, error) {
	if n == nil {
		return capabilities.Node{}, errors.New("nil node reply")
	}

	var peerID ragetypes.PeerID
	if len(n.GetPeerId()) != 0 {
		if len(n.GetPeerId()) != len(peerID) {
			return capabilities.Node{}, fmt.Errorf("invalid peer ID length %d", len(n.GetPeerId()))
		}
		copy(peerID[:], n.GetPeerId())
	}

	var signer, encryptionPublicKey [32]byte
	copy(signer[:], n.GetSigner())
	copy(encryptionPublicKey[:], n.GetEncryptionPublicKey())

	capabilityDONs := make([]capabilities.DON, 0, len(n.GetCapabilityDons()))
	for _, d := range n.GetCapabilityDons() {
		capabilityDONs = append(capabilityDONs, DONFromProto(d))
	}

	return capabilities.Node{
		PeerID:              &peerID,
		NodeOperatorID:      n.GetNodeOperatorId(),
		Signer:              signer,
		EncryptionPublicKey: encryptionPublicKey,
		WorkflowDON:         DONFromProto(n.GetWorkflowDon()),
		CapabilityDONs:      capabilityDONs,
	}, nil
}
