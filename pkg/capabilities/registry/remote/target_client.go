package remote

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

// TargetClient is a client for a registry that holds the addresses capabilities
// are served at rather than the capabilities themselves. It dials what it is
// told about, so registration names an address instead of passing a value.
type TargetClient struct {
	lggr logger.Logger
	grpc registrypb.CapabilitiesRegistryClient

	// dialOpts are applied when dialing capability addresses. They do not apply
	// to the registry connection, which the caller supplies already dialed.
	dialOpts []grpc.DialOption

	mu    sync.Mutex
	conns map[string]*grpc.ClientConn
}

// NewTargetClient returns a client for the registry on the other end of cc,
// which must be already dialed. capabilityDialOpts apply to the capability
// addresses this client resolves, not to cc.
func NewTargetClient(lggr logger.Logger, cc grpc.ClientConnInterface, capabilityDialOpts ...grpc.DialOption) *TargetClient {
	return &TargetClient{
		lggr:     logger.Named(lggr, "RemoteRegistry"),
		grpc:     registrypb.NewCapabilitiesRegistryClient(cc),
		dialOpts: capabilityDialOpts,
		conns:    map[string]*grpc.ClientConn{},
	}
}

// Close tears down every cached capability connection. The registry connection
// it was given is the caller's to close.
func (c *TargetClient) Close() error {
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

func (c *TargetClient) connFor(addr string) (*grpc.ClientConn, error) {
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
func (c *TargetClient) wrap(h *registrypb.CapabilityHandle) (capabilities.BaseCapability, capabilities.CapabilityType, error) {
	capType, err := capabilitiespb.CapabilityTypeFromProto(h.GetType())
	if err != nil {
		return nil, capabilities.CapabilityTypeUnknown, fmt.Errorf("capability %s: %w", h.GetCapabilityId(), err)
	}

	conn, err := c.connFor(h.GetCallbackUrl())
	if err != nil {
		return nil, capType, err
	}

	return Wrap(c.lggr, conn, capType), capType, nil
}

// lookupFn is the shape of the registry's three handle-resolving RPCs.
type lookupFn func(context.Context, *registrypb.GetRequest, ...grpc.CallOption) (*registrypb.CapabilityHandle, error)

// resolve looks a capability up and returns it as the surface T.
//
// The three Get* methods differ only in which RPC they call and which surface
// they demand, so the lookup, wrap and assert steps live here once. The assert
// is [Resolve], which both registry transports share.
func resolve[T capabilities.BaseCapability](ctx context.Context, c *TargetClient, id, surface string, lookup lookupFn) (T, error) {
	var zero T

	h, err := lookup(ctx, &registrypb.GetRequest{CapabilityId: id})
	if err != nil {
		return zero, err
	}

	capType, err := capabilitiespb.CapabilityTypeFromProto(h.GetType())
	if err != nil {
		return zero, fmt.Errorf("capability %s: %w", id, err)
	}

	conn, err := c.connFor(h.GetCallbackUrl())
	if err != nil {
		return zero, err
	}

	return Resolve[T](c.lggr, conn, capType, id, surface)
}

func (c *TargetClient) Get(ctx context.Context, id string) (capabilities.BaseCapability, error) {
	return resolve[capabilities.BaseCapability](ctx, c, id, "base", c.grpc.Get)
}

func (c *TargetClient) GetTrigger(ctx context.Context, id string) (capabilities.TriggerCapability, error) {
	return resolve[capabilities.TriggerCapability](ctx, c, id, "trigger", c.grpc.GetTrigger)
}

func (c *TargetClient) GetExecutable(ctx context.Context, id string) (capabilities.ExecutableCapability, error) {
	return resolve[capabilities.ExecutableCapability](ctx, c, id, "executable", c.grpc.GetExecutable)
}

func (c *TargetClient) List(ctx context.Context) ([]capabilities.BaseCapability, error) {
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
func (c *TargetClient) OCRConfig(ctx context.Context, capabilityID string, donID uint32, key string) (ocrtypes.ContractConfig, error) {
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
func (c *TargetClient) AddAt(ctx context.Context, id string, capType capabilities.CapabilityType, addr string) error {
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

func (c *TargetClient) Remove(ctx context.Context, id string) error {
	_, err := c.grpc.Remove(ctx, &registrypb.RemoveRequest{CapabilityId: id})
	return err
}

// --- metadata ---

func (c *TargetClient) LocalNode(ctx context.Context) (capabilities.Node, error) {
	reply, err := c.grpc.LocalNode(ctx, &emptypb.Empty{})
	if err != nil {
		return capabilities.Node{}, err
	}
	return NodeFromProto(reply, reply.GetWorkflowDon(), reply.GetCapabilityDons())
}

func (c *TargetClient) NodeByPeerID(ctx context.Context, peerID ragetypes.PeerID) (capabilities.Node, error) {
	reply, err := c.grpc.NodeByPeerID(ctx, &registrypb.NodeRequest{PeerId: peerID[:]})
	if err != nil {
		return capabilities.Node{}, err
	}
	return NodeFromProto(reply, reply.GetWorkflowDon(), reply.GetCapabilityDons())
}

// ConfigForCapability decodes the capability configuration the registry serves.
func (c *TargetClient) ConfigForCapability(ctx context.Context, capabilityID string, donID uint32) (capabilities.CapabilityConfiguration, error) {
	reply, err := c.grpc.ConfigForCapability(ctx, &registrypb.ConfigForCapabilityRequest{
		CapabilityId: capabilityID,
		DonId:        donID,
	})
	if err != nil {
		return capabilities.CapabilityConfiguration{}, err
	}

	return capabilitiespb.CapabilityConfigFromProto(reply.GetCapabilityConfig())
}

func (c *TargetClient) DONsForCapability(ctx context.Context, capabilityID string) ([]capabilities.DONWithNodes, error) {
	reply, err := c.grpc.DONsForCapability(ctx, &registrypb.DONsForCapabilityRequest{CapabilityId: capabilityID})
	if err != nil {
		return nil, err
	}

	out := make([]capabilities.DONWithNodes, 0, len(reply.GetDons()))
	for _, d := range reply.GetDons() {
		nodes := make([]capabilities.Node, 0, len(d.GetNodes()))
		for _, n := range d.GetNodes() {
			node, err := NodeFromProto(n, n.GetWorkflowDon(), n.GetCapabilityDons())
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, node)
		}
		out = append(out, capabilities.DONWithNodes{DON: DONFromProto(d.GetDon()), Nodes: nodes})
	}
	return out, nil
}

func (c *TargetClient) DONByID(ctx context.Context, donID uint32) (capabilities.DON, error) {
	reply, err := c.grpc.DONByID(ctx, &registrypb.DONByIDRequest{DonId: donID})
	if err != nil {
		return capabilities.DON{}, err
	}
	return DONFromProto(reply.GetDon()), nil
}
