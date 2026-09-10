package registry

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/grpc"

	ocrtypes "github.com/smartcontractkit/libocr/offchainreporting2plus/types"
	ragetypes "github.com/smartcontractkit/libocr/ragep2p/types"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

// Registry is a capabilities registry that can also say what OCR configuration a capability runs
// under.
//
// The two travel together because one process answers both: whoever read the registry contract is
// the only one that can compute a configuration's digest, which covers the chain and address it was
// read from. See core.OCRConfigRegistry.
type Registry interface {
	core.CapabilitiesRegistry

	// Also addressable: a registry with something behind it holds addresses as well as values, so a
	// caller that knows where a capability is served can say so directly rather than through Add.
	core.AddressableRegistryBase

	core.OCRConfigRegistry

	// Close releases whatever this registry dialled of its own accord - the capability addresses it
	// resolved, not the connection it was handed. A registry with nothing behind it closes nothing.
	io.Closer
}

// Local returns the registry of the capabilities this process holds itself.
//
// On its own it is the whole registry of a process with nothing behind it: it resolves what was
// added to it and nothing else, and the metadata calls - which DONs exist, which nodes are in them,
// what OCR configuration a capability runs under - fail, because a process holding capability values
// has no way to know any of it. WithRemote is what puts a registry behind it.
func Local(lggr logger.Logger) *LocalRegistry {
	return &LocalRegistry{base: NewBaseRegistry(lggr), lggr: lggr}
}

// LocalRegistry holds capability values in this process. Compose it with WithRemote to reach the
// ones it does not hold.
type LocalRegistry struct {
	base core.CapabilitiesRegistryBase
	lggr logger.Logger
}

var _ Registry = (*LocalRegistry)(nil)

func (r *LocalRegistry) Add(ctx context.Context, c capabilities.BaseCapability) error {
	return r.base.Add(ctx, c)
}

func (r *LocalRegistry) Remove(ctx context.Context, id string) error { return r.base.Remove(ctx, id) }

func (r *LocalRegistry) Get(ctx context.Context, id string) (capabilities.BaseCapability, error) {
	return r.base.Get(ctx, id)
}

func (r *LocalRegistry) GetTrigger(ctx context.Context, id string) (capabilities.TriggerCapability, error) {
	return r.base.GetTrigger(ctx, id)
}

func (r *LocalRegistry) GetExecutable(ctx context.Context, id string) (capabilities.ExecutableCapability, error) {
	return r.base.GetExecutable(ctx, id)
}

func (r *LocalRegistry) List(ctx context.Context) ([]capabilities.BaseCapability, error) {
	return r.base.List(ctx)
}

// The metadata a local registry cannot answer. It holds capability values, which say nothing about
// the registry that named them - so it says so, rather than answering with something invented that a
// caller would act on.

func (r *LocalRegistry) LocalNode(context.Context) (capabilities.Node, error) {
	return capabilities.Node{}, errLocalOnly("which node this is")
}

func (r *LocalRegistry) NodeByPeerID(context.Context, ragetypes.PeerID) (capabilities.Node, error) {
	return capabilities.Node{}, errLocalOnly("which node a peer is")
}

func (r *LocalRegistry) ConfigForCapability(context.Context, string, uint32) (capabilities.CapabilityConfiguration, error) {
	return capabilities.CapabilityConfiguration{}, errLocalOnly("how a capability is configured")
}

func (r *LocalRegistry) DONsForCapability(context.Context, string) ([]capabilities.DONWithNodes, error) {
	return nil, errLocalOnly("which DONs host a capability")
}

func (r *LocalRegistry) DONByID(context.Context, uint32) (capabilities.DON, error) {
	return capabilities.DON{}, errLocalOnly("which DON an ID names")
}

func (r *LocalRegistry) OCRConfig(context.Context, string, uint32, string) (ocrtypes.ContractConfig, error) {
	return ocrtypes.ContractConfig{}, errLocalOnly("what OCR configuration a capability runs under")
}

// AddAt is the announcement a local registry cannot make: holding a value is all it does, and there
// is nothing behind it for an address to mean anything to.
func (r *LocalRegistry) AddAt(context.Context, string, capabilities.CapabilityType, string) error {
	return errLocalOnly("where a capability can be reached, having nothing to announce it to")
}

// Close is a no-op: a local registry dials nothing.
func (r *LocalRegistry) Close() error { return nil }

func errLocalOnly(question string) error {
	return fmt.Errorf("this registry holds only the capabilities this process registered, so it cannot say %s", question)
}

// WithRemote returns a registry that resolves what this process holds before asking the registry on
// the other end of conn for anything else - and that answers, from there, everything a local
// registry cannot.
//
// Local first is not an optimisation. A capability this process hosts is a value it already holds,
// so resolving it locally hands back the implementation rather than a client looping back into this
// same process - and it works before this process has announced anything, which is what lets one
// capability here call another during startup.
//
// addresses is where the capabilities this process holds are served, by capability ID. Add uses it
// to announce: a remote registry holds addresses rather than values, so registering with one means
// naming where the value can be reached. A capability added with no entry there is held locally and
// not announced, which is what a process serving nothing wants.
//
// conn is taken already dialled rather than as an address, so that a caller reaching one process for
// several things - its registry and its OCR configurations, say - reaches it over one connection.
// capabilityDialOpts are for the capability addresses this registry resolves, not for conn.
func (r *LocalRegistry) WithRemote(conn grpc.ClientConnInterface, addresses map[string]string, capabilityDialOpts ...grpc.DialOption) Registry {
	return &proxiedRegistry{
		lggr:      logger.Named(r.lggr, "ProxiedRegistry"),
		local:     r,
		remote:    newRemote(r.lggr, conn, capabilityDialOpts...),
		addresses: addresses,
	}
}

// proxiedRegistry is a local registry with a remote one behind it.
type proxiedRegistry struct {
	lggr      logger.Logger
	local     *LocalRegistry
	remote    *remote
	addresses map[string]string
}

var _ Registry = (*proxiedRegistry)(nil)

// Add holds c in this process and, when something serves it, announces where.
//
// Both halves are the same registration seen from two sides: holding the value is what makes it
// resolvable here, and announcing the address is what makes it resolvable anywhere else.
func (r *proxiedRegistry) Add(ctx context.Context, c capabilities.BaseCapability) error {
	if err := r.local.Add(ctx, c); err != nil {
		return err
	}

	info, err := c.Info(ctx)
	if err != nil {
		return err
	}
	address, ok := r.addresses[info.ID]
	if !ok {
		// Held, not announced: nothing serves it, so there is no address to invite traffic to.
		return nil
	}
	return r.remote.AddAt(ctx, info.ID, info.CapabilityType, address)
}

// Remove drops a capability from this process and from the remote registry, and fails only if
// neither had it.
//
// Either half alone missing it is ordinary rather than exceptional: a capability registered by
// address was never held here, one never announced was never there, and a caller asking for it to be
// gone wants it gone rather than a report on which of the two places it was in.
func (r *proxiedRegistry) Remove(ctx context.Context, id string) error {
	localErr := r.local.Remove(ctx, id)
	remoteErr := r.remote.Remove(ctx, id)

	switch {
	case localErr == nil && remoteErr == nil:
		return nil
	case localErr == nil:
		r.lggr.Debugw("capability not removed from the remote registry", "capabilityID", id, "err", remoteErr)
		return nil
	case remoteErr == nil:
		r.lggr.Debugw("capability was not held in this process", "capabilityID", id, "err", localErr)
		return nil
	default:
		return fmt.Errorf("capability %s removed from neither this process (%w) nor the remote registry: %w", id, localErr, remoteErr)
	}
}

// AddAt announces a capability served at addr, for a caller holding the address rather than the
// value - the shape a registry of addresses is written to.
func (r *proxiedRegistry) AddAt(ctx context.Context, id string, capType capabilities.CapabilityType, addr string) error {
	return r.remote.AddAt(ctx, id, capType, addr)
}

func (r *proxiedRegistry) Get(ctx context.Context, id string) (capabilities.BaseCapability, error) {
	return resolveLocalFirst(ctx, id, r.local.Get, r.remote.Get)
}

func (r *proxiedRegistry) GetTrigger(ctx context.Context, id string) (capabilities.TriggerCapability, error) {
	return resolveLocalFirst(ctx, id, r.local.GetTrigger, r.remote.GetTrigger)
}

func (r *proxiedRegistry) GetExecutable(ctx context.Context, id string) (capabilities.ExecutableCapability, error) {
	return resolveLocalFirst(ctx, id, r.local.GetExecutable, r.remote.GetExecutable)
}

// List returns everything this process holds plus everything the remote registry knows about, local
// entries winning on ID so a capability held here comes back as the value rather than as a client
// dialling back into this process.
func (r *proxiedRegistry) List(ctx context.Context) ([]capabilities.BaseCapability, error) {
	local, err := r.local.List(ctx)
	if err != nil {
		return nil, err
	}

	remote, err := r.remote.List(ctx)
	if err != nil {
		// The local half is still usable and still correct; a registry that cannot be reached should
		// not blank the capabilities this process holds.
		r.lggr.Warnw("failed to list capabilities from the remote registry", "err", err)
		return local, nil
	}

	seen := make(map[string]bool, len(local))
	for _, c := range local {
		info, ierr := c.Info(ctx)
		if ierr != nil {
			return nil, fmt.Errorf("failed to read local capability info: %w", ierr)
		}
		seen[info.ID] = true
	}
	for _, c := range remote {
		info, ierr := c.Info(ctx)
		if ierr != nil {
			r.lggr.Warnw("skipping remote capability whose info could not be read", "err", ierr)
			continue
		}
		if !seen[info.ID] {
			local = append(local, c)
		}
	}
	return local, nil
}

// --- everything only the registry on the other end knows ---

func (r *proxiedRegistry) LocalNode(ctx context.Context) (capabilities.Node, error) {
	return r.remote.LocalNode(ctx)
}

func (r *proxiedRegistry) NodeByPeerID(ctx context.Context, peerID ragetypes.PeerID) (capabilities.Node, error) {
	return r.remote.NodeByPeerID(ctx, peerID)
}

func (r *proxiedRegistry) ConfigForCapability(ctx context.Context, capabilityID string, donID uint32) (capabilities.CapabilityConfiguration, error) {
	return r.remote.ConfigForCapability(ctx, capabilityID, donID)
}

func (r *proxiedRegistry) DONsForCapability(ctx context.Context, capabilityID string) ([]capabilities.DONWithNodes, error) {
	return r.remote.DONsForCapability(ctx, capabilityID)
}

func (r *proxiedRegistry) DONByID(ctx context.Context, donID uint32) (capabilities.DON, error) {
	return r.remote.DONByID(ctx, donID)
}

func (r *proxiedRegistry) OCRConfig(ctx context.Context, capabilityID string, donID uint32, key string) (ocrtypes.ContractConfig, error) {
	return r.remote.OCRConfig(ctx, capabilityID, donID, key)
}

// Close tears down the connections the remote half opened to capability addresses. The connection it
// was given is the caller's to close.
func (r *proxiedRegistry) Close() error { return r.remote.Close() }

// resolveLocalFirst tries this process, then the remote registry. A local miss is expected rather
// than exceptional - most capabilities live elsewhere - so its error is only reported if the remote
// registry cannot resolve the ID either, where it is the more useful half of the answer: it says
// what this process does hold.
func resolveLocalFirst[T capabilities.BaseCapability](
	ctx context.Context,
	id string,
	local, remote func(context.Context, string) (T, error),
) (T, error) {
	var zero T

	got, localErr := local(ctx, id)
	if localErr == nil {
		return got, nil
	}

	got, remoteErr := remote(ctx, id)
	if remoteErr == nil {
		return got, nil
	}
	return zero, fmt.Errorf("capability %s not found locally (%w) or in the remote registry: %w", id, localErr, remoteErr)
}
