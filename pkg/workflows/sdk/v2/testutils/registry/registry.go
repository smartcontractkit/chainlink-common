package registry

import (
	"sync"
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/registry"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

var testRegistries = map[testing.TB]*Registry{}

var registryLock sync.Mutex

func GetRegistry(tb testing.TB) *Registry {
	registryLock.Lock()
	defer registryLock.Unlock()
	if r, ok := testRegistries[tb]; ok {
		return r
	}

	r := &Registry{tb: tb, CapabilitiesRegistryBase: registry.NewBaseRegistry(logger.Test(tb))}
	testRegistries[tb] = r
	tb.Cleanup(func() {
		delete(testRegistries, tb)
	})
	return r
}

// Registry is meant to be used with GetRegistry, do not use it directly.
type Registry struct {
	core.CapabilitiesRegistryBase
	tb testing.TB
}

func (r *Registry) RegisterCapability(c Capability) error {
	return r.Add(r.tb.Context(), &CapabilityWrapper{Capability: c})
}

func (r *Registry) GetCapability(id string) (Capability, error) {
	v1Capability, err := r.Get(r.tb.Context(), id)
	if err != nil {
		return nil, err
	}

	capability, ok := v1Capability.(Capability)
	if ok {
		return capability, nil
	}

	return &FakeWrapper{BaseCapability: v1Capability, tb: r.tb}, nil
}
