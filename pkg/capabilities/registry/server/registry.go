package server

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"

	"google.golang.org/grpc"

	ragetypes "github.com/smartcontractkit/libocr/ragep2p/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry/client"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

// Handle is a registered capability: its ID, which capability services it
// serves, and the gRPC target serving them.
//
// Unlike the go-plugin registry, a Handle holds no connection and owns no
// resources, so registering and resolving a capability allocates nothing that
// has to be reclaimed. Reachability of URL is the registrant's responsibility
// for as long as the capability stays registered.
type Handle struct {
	ID   string
	Type capabilities.CapabilityType
	URL  string
}

// Registry is the CapabilitiesRegistry crecore serves.
//
// It composes two independent sources:
//
//   - Handles, registered at runtime over Add by the processes that host the capabilities. Add
//     dials the handle's address itself and wraps it into a real capabilities.BaseCapability (the
//     same conversion registry/client does for its own callers), then adds that to local, chainlink's
//     ordinary in-process base registry. Add is a drop-in replacement for the in-process
//     Add(BaseCapability) core uses without a proxy: a capability registered this way is just as real
//     and callable as one added directly, whether the caller asking for it is in this process or
//     another one reached the normal way, over the gRPC service Get/GetTrigger/GetExecutable wrap.
//   - Metadata, supplied by a MetadataSource the owner installs. Read-only from
//     here; whoever provides it decides how it is refreshed.
//
// Registry is safe for concurrent use.
type Registry struct {
	lggr logger.Logger

	// metadata is set after construction: the gRPC service is typically registered
	// before whatever reads the chain is ready, so metadata RPCs must be answerable
	// (with an error) before a source exists.
	metadata atomic.Value // MetadataSource

	// local holds the real, callable value Add resolves a Handle to. Anything in this process that
	// wants to call a capability - locally added, or reached only by dialing a Handle's address -
	// goes through this, not through handles below.
	local core.CapabilitiesRegistryBase

	// dialOpts are applied when dialing a Handle's address to resolve it into local.
	dialOpts []grpc.DialOption

	mu      sync.RWMutex
	handles map[string]Handle
	conns   map[string]*grpc.ClientConn // by URL; closed and dropped on Remove
}

func New(lggr logger.Logger, dialOpts ...grpc.DialOption) *Registry {
	return &Registry{
		lggr:     logger.Named(lggr, "CapabilitiesRegistry"),
		local:    registry.NewBaseRegistry(lggr),
		dialOpts: dialOpts,
		handles:  map[string]Handle{},
		conns:    map[string]*grpc.ClientConn{},
	}
}

// Local is the real, in-process registry Add resolves Handles into: whatever in this process wants
// to actually call a capability - rather than tell some other process where to find it - uses this.
func (r *Registry) Local() core.CapabilitiesRegistryBase { return r.local }

// MetadataSource is the registry's view of on-chain state.
//
// It is an interface so the reader that produces it — which needs chain clients
// and contract bindings — can live outside this package, leaving the registry and
// its wire adapter free of those dependencies.
type MetadataSource interface {
	LocalNode(ctx context.Context) (capabilities.Node, error)
	NodeByPeerID(ctx context.Context, peerID ragetypes.PeerID) (capabilities.Node, error)
	DONsForCapability(ctx context.Context, capabilityID string) ([]capabilities.DONWithNodes, error)
	DONByID(ctx context.Context, donID uint32) (capabilities.DON, error)
	RawConfigForCapability(ctx context.Context, capabilityID string, donID uint32) ([]byte, error)
}

// SetMetadata installs the on-chain metadata source. Until it is called, metadata
// RPCs fail with a "not ready" error while Add/Get/List work normally.
func (r *Registry) SetMetadata(src MetadataSource) { r.metadata.Store(src) }

// current returns the metadata source, or an error if none is installed yet.
func (r *Registry) current() (MetadataSource, error) {
	src, _ := r.metadata.Load().(MetadataSource)
	if src == nil {
		return nil, errors.New("registry metadata not ready: no metadata source installed")
	}
	return src, nil
}

// Add registers a capability served at url.
//
// Re-registration is expected rather than exceptional, so it is accepted in both
// forms:
//
//   - Same address as the existing entry: nothing to change, so this succeeds
//     without touching the registry. A capability that serves itself at a fixed
//     address re-announces that same address every time it restarts, and failing
//     the second registration would take the capability out of service even though
//     the entry already points exactly where it should.
//   - Different address: the host moved, so the entry is repointed. Requiring a
//     Remove first would lose the capability for the window between the two calls.
//
// This registry deliberately does not try to distinguish a restart from a second
// process claiming the same ID. It cannot: it holds an address, not a connection,
// so it has no way to tell whether the existing entry is still alive. Ownership of
// a capability ID is settled on chain, and a caller that wants the stricter rule
// has the liveness information locally — see how core gates replacement on
// connection state in chainlink-common's baseRegistry.
func (r *Registry) Add(ctx context.Context, h Handle) error {
	if h.ID == "" {
		return errors.New("capability ID is required")
	}
	if h.URL == "" {
		return fmt.Errorf("callback URL is required to register capability %s", h.ID)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if prev, ok := r.handles[h.ID]; ok && prev.URL == h.URL && prev.Type == h.Type {
		r.lggr.Debugw("capability re-registered at the same address; nothing to do",
			"capabilityID", h.ID, "url", h.URL)
		return nil
	}

	conn, err := grpc.NewClient(h.URL, r.dialOpts...)
	if err != nil {
		return fmt.Errorf("failed to dial capability %s at %s: %w", h.ID, h.URL, err)
	}
	wrapped, err := client.Wrap(r.lggr, conn, h.Type)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to wrap capability %s at %s: %w", h.ID, h.URL, err)
	}

	// The wrapped value cannot report connection state (it is a plain RPC client, not a
	// *grpc.ClientConn), so local's own liveness-gated replace never fires - drop the previous
	// entry explicitly instead so a moved capability is replaced rather than rejected.
	if prevConn, ok := r.conns[h.ID]; ok {
		_ = r.local.Remove(ctx, h.ID)
		_ = prevConn.Close()
	}
	if err := r.local.Add(ctx, wrapped); err != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to add capability %s to local registry: %w", h.ID, err)
	}
	r.conns[h.ID] = conn

	if prev, ok := r.handles[h.ID]; ok {
		r.lggr.Infow("re-registering capability at a new address",
			"capabilityID", h.ID, "previousURL", prev.URL, "url", h.URL)
	}
	r.handles[h.ID] = h
	r.lggr.Infow("capability registered", "capabilityID", h.ID, "type", h.Type, "url", h.URL)
	return nil
}

func (r *Registry) Remove(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if conn, ok := r.conns[id]; ok {
		_ = r.local.Remove(ctx, id)
		_ = conn.Close()
		delete(r.conns, id)
	}

	if _, ok := r.handles[id]; !ok {
		return fmt.Errorf("capability %s not found", id)
	}
	delete(r.handles, id)
	r.lggr.Infow("capability removed", "capabilityID", id)
	return nil
}

// Get resolves a capability of any type.
func (r *Registry) Get(_ context.Context, id string) (Handle, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	h, ok := r.handles[id]
	if !ok {
		return Handle{}, fmt.Errorf("capability %s not found", id)
	}
	return h, nil
}

// GetTrigger resolves a capability that must serve TriggerExecutable.
func (r *Registry) GetTrigger(ctx context.Context, id string) (Handle, error) {
	h, err := r.Get(ctx, id)
	if err != nil {
		return Handle{}, err
	}
	switch h.Type {
	case capabilities.CapabilityTypeTrigger, capabilities.CapabilityTypeCombined:
		return h, nil
	default:
		return Handle{}, fmt.Errorf("capability %s is a %s, not a trigger", id, h.Type)
	}
}

// GetExecutable resolves a capability that must serve Executable.
func (r *Registry) GetExecutable(ctx context.Context, id string) (Handle, error) {
	h, err := r.Get(ctx, id)
	if err != nil {
		return Handle{}, err
	}
	switch h.Type {
	case capabilities.CapabilityTypeAction,
		capabilities.CapabilityTypeTarget,
		capabilities.CapabilityTypeConsensus,
		capabilities.CapabilityTypeCombined:
		return h, nil
	default:
		return Handle{}, fmt.Errorf("capability %s is a %s, not executable", id, h.Type)
	}
}

// List returns every registered capability, ordered by ID so callers and tests
// see a stable sequence.
func (r *Registry) List(_ context.Context) []Handle {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Handle, 0, len(r.handles))
	for _, h := range r.handles {
		out = append(out, h)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// --- metadata, delegated to the current on-chain snapshot ---

func (r *Registry) LocalNode(ctx context.Context) (capabilities.Node, error) {
	lr, err := r.current()
	if err != nil {
		return capabilities.Node{}, err
	}
	return lr.LocalNode(ctx)
}

func (r *Registry) NodeByPeerID(ctx context.Context, peerID ragetypes.PeerID) (capabilities.Node, error) {
	lr, err := r.current()
	if err != nil {
		return capabilities.Node{}, err
	}
	return lr.NodeByPeerID(ctx, peerID)
}

func (r *Registry) DONsForCapability(ctx context.Context, capabilityID string) ([]capabilities.DONWithNodes, error) {
	lr, err := r.current()
	if err != nil {
		return nil, err
	}
	return lr.DONsForCapability(ctx, capabilityID)
}

func (r *Registry) DONByID(ctx context.Context, donID uint32) (capabilities.DON, error) {
	lr, err := r.current()
	if err != nil {
		return capabilities.DON{}, err
	}
	return lr.DONByID(ctx, donID)
}

// RawConfigForCapability returns the undecoded capability config bytes; see
// DON.CapabilityConfigurations for why they are not decoded here.
func (r *Registry) RawConfigForCapability(ctx context.Context, capabilityID string, donID uint32) ([]byte, error) {
	lr, err := r.current()
	if err != nil {
		return nil, err
	}
	return lr.RawConfigForCapability(ctx, capabilityID, donID)
}
