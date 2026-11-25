package registry

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Masterminds/semver/v3"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

var (
	ErrCapabilityAlreadyExists = errors.New("capability already exists")
)

type baseRegistry struct {
	m    map[string]atomicBaseCapability
	lggr logger.Logger
	// mu guards m. It is not held while a capability is replaced or removed: that can
	// wait for calls in flight on that one capability, and must not block the others.
	mu      sync.RWMutex
	timings replaceTimings
}

var _ CapabilitiesRegistryBase = (*baseRegistry)(nil)

func NewBaseRegistry(lggr logger.Logger) CapabilitiesRegistryBase {
	return &baseRegistry{
		m:       map[string]atomicBaseCapability{},
		lggr:    logger.Named(lggr, "registries.basic"),
		timings: replaceTimings{grace: defaultReplaceGracePeriod, replay: defaultReplayTimeout},
	}
}

// Get gets a capability from the registry.
func (r *baseRegistry) Get(_ context.Context, id string) (capabilities.BaseCapability, error) {
	r.lggr.Debugw("get capability", "id", id)
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.m[id]
	if ok {
		return c, nil
	}

	// Find compatible version (>= requested version with same major)
	parts := strings.Split(id, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid capability id format: %s", id)
	}
	name, verStr := parts[0], parts[1]

	reqVer, err := semver.NewVersion(verStr)
	if err != nil {
		return nil, fmt.Errorf("invalid version in capability id %q: %w", id, err)
	}
	reqIsPrerelease := reqVer.Prerelease() != ""

	var bestCap capabilities.BaseCapability
	var bestVer *semver.Version
	for key, cap := range r.m {
		p := strings.Split(key, "@")
		if len(p) != 2 {
			continue
		}
		if p[0] != name {
			continue
		}
		v, err := semver.NewVersion(p[1])
		if err != nil {
			continue
		}
		if v.Major() != reqVer.Major() {
			continue
		}
		// If the request is stable, skip pre-release candidates
		if !reqIsPrerelease && v.Prerelease() != "" {
			continue
		}

		if v.GreaterThan(reqVer) {
			if bestVer == nil || v.LessThan(bestVer) {
				bestCap = cap
				bestVer = v
			}
		}
	}

	if bestCap != nil {
		r.lggr.Debugw("found compatible capability", "id", name+"@"+bestVer.String())
		return bestCap, nil
	}
	return nil, fmt.Errorf("no compatible capability found for id %s", id)
}

// GetTrigger gets a capability from the registry and tries to coerce it to the TriggerCapability interface.
func (r *baseRegistry) GetTrigger(ctx context.Context, id string) (capabilities.TriggerCapability, error) {
	c, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	tc, ok := c.(capabilities.TriggerCapability)
	if !ok {
		return nil, fmt.Errorf("capability with id: %s does not satisfy the capability interface", id)
	}

	return tc, nil
}

// GetExecutable gets a capability from the registry and tries to coerce it to the ExecutableCapability interface.
func (r *baseRegistry) GetExecutable(ctx context.Context, id string) (capabilities.ExecutableCapability, error) {
	c, err := r.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	ac, ok := c.(capabilities.ExecutableCapability)
	if !ok {
		return nil, fmt.Errorf("capability with id: %s does not satisfy the capability interface", id)
	}

	return ac, nil
}

func (r *baseRegistry) List(_ context.Context) ([]capabilities.BaseCapability, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var cl []capabilities.BaseCapability
	for _, v := range r.m {
		cl = append(cl, v)
	}

	return cl, nil
}

// Add adds a capability to the registry, or replaces the one with the same ID when the
// connection of that one is down.
//
// A replaced capability gets a grace period for its calls in flight; after that they are
// cancelled with [ErrCapabilityReplaced]. Its trigger registrations are replayed onto the
// new capability in the background, after Add has returned.
func (r *baseRegistry) Add(ctx context.Context, c capabilities.BaseCapability) error {
	if c == nil {
		return errors.New("cannot add a nil capability to the registry")
	}
	// Info is a call into the capability, so it is made before the registry is locked.
	info, err := c.Info(ctx)
	if err != nil {
		return err
	}
	id := info.ID

	existing, err := r.addIfNew(ctx, id, info.CapabilityType, c)
	if err != nil {
		return err
	}
	if existing != nil {
		if err := existing.Replace(ctx, c); err != nil {
			var up *stillUpError
			if errors.As(err, &up) {
				return fmt.Errorf("%w: id %s found in registry: state %s", ErrCapabilityAlreadyExists, id, up.state)
			}
			return fmt.Errorf("failed to update capability %s: %w", id, err)
		}
	}
	r.lggr.Infow("capability added", "id", id, "type", info.CapabilityType, "description", info.Description, "version", info.Version())
	return nil
}

// addIfNew adds c under id when the registry has no entry for id. Otherwise it changes
// nothing and returns the entry.
func (r *baseRegistry) addIfNew(ctx context.Context, id string, typ capabilities.CapabilityType, c capabilities.BaseCapability) (atomicBaseCapability, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if existing, ok := r.m[id]; ok {
		return existing, nil
	}

	var ac atomicBaseCapability
	lggr := logger.With(r.lggr, "capabilityID", id)
	switch typ {
	case capabilities.CapabilityTypeTrigger:
		ac = newAtomicTriggerCapability(lggr, r.timings)
	case capabilities.CapabilityTypeAction, capabilities.CapabilityTypeConsensus, capabilities.CapabilityTypeTarget:
		ac = newAtomicExecuteCapability(r.timings)
	case capabilities.CapabilityTypeCombined:
		ac = newAtomicExecuteAndTriggerCapability(lggr, r.timings)
	default:
		return nil, fmt.Errorf("unknown capability type: %s", typ)
	}
	// Nothing uses a new entry yet, so this does not wait.
	if err := ac.Update(ctx, c); err != nil {
		return nil, err
	}
	r.m[id] = ac
	return nil, nil
}

// Remove takes the capability out of use. Its calls in flight get a grace period and are
// then cancelled with [ErrCapabilityReplaced]. The entry keeps its trigger registrations,
// for the capability that is added under the same ID next.
func (r *baseRegistry) Remove(ctx context.Context, id string) error {
	r.mu.RLock()
	ac, ok := r.m[id]
	r.mu.RUnlock()
	if !ok {
		return fmt.Errorf("unable to remove, capability not found: %s", id)
	}
	if err := ac.Update(ctx, nil); err != nil {
		return fmt.Errorf("failed to remove capability %s: %w", id, err)
	}
	r.lggr.Infow("capability removed", "id", id)
	return nil
}
