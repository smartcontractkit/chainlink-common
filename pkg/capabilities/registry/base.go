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
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

var (
	ErrCapabilityAlreadyExists = errors.New("capability already exists")
)

type baseRegistry struct {
	m    map[string]capabilities.BaseCapability
	lggr logger.Logger
	mu   sync.RWMutex
}

var _ core.CapabilitiesRegistryBase = (*baseRegistry)(nil)

func NewBaseRegistry(lggr logger.Logger) core.CapabilitiesRegistryBase {
	return &baseRegistry{
		m:    map[string]capabilities.BaseCapability{},
		lggr: logger.Named(lggr, "registries.basic"),
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

func (r *baseRegistry) Add(ctx context.Context, c capabilities.BaseCapability) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	info, err := c.Info(ctx)
	if err != nil {
		return err
	}

	switch info.CapabilityType {
	case capabilities.CapabilityTypeTrigger:
		_, ok := c.(capabilities.TriggerCapability)
		if !ok {
			return errors.New("trigger capability does not satisfy TriggerCapability interface")
		}
	case capabilities.CapabilityTypeAction, capabilities.CapabilityTypeConsensus, capabilities.CapabilityTypeTarget:
		_, ok := c.(capabilities.ExecutableCapability)
		if !ok {
			return errors.New("action does not satisfy ExecutableCapability interface")
		}
	case capabilities.CapabilityTypeCombined:
		_, ok := c.(capabilities.ExecutableAndTriggerCapability)
		if !ok {
			return errors.New("target capability does not satisfy ExecutableAndTriggerCapability interface")
		}
	default:
		return fmt.Errorf("unknown capability type: %s", info.CapabilityType)
	}

	id := info.ID
	_, ok := r.m[id]
	if ok {
		return fmt.Errorf("%w: id %s found in registry", ErrCapabilityAlreadyExists, id)
	}

	r.m[id] = c
	r.lggr.Infow("capability added", "id", id, "type", info.CapabilityType, "description", info.Description, "version", info.Version())
	return nil
}

func (r *baseRegistry) Remove(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.m[id]
	if !ok {
		return fmt.Errorf("unable to remove, capability not found: %s", id)
	}

	delete(r.m, id)
	r.lggr.Infow("capability removed", "id", id)
	return nil
}
