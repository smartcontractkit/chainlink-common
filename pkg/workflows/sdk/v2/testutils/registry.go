package testutils

import (
	"fmt"
	"sync"
	"testing"
)

var testRegistries = map[testing.TB]*Registry{}

var registryLock sync.Mutex

func GetRegistry(tb testing.TB) *Registry {
	registryLock.Lock()
	defer registryLock.Unlock()
	if r, ok := testRegistries[tb]; ok {
		return r
	}

	r := &Registry{}
	testRegistries[tb] = r
	tb.Cleanup(func() {
		delete(testRegistries, tb)
	})
	return r
}

type Registry struct {
	capabilities map[string]Capability
}

func (r *Registry) RegisterCapability(c Capability) error {
	if r.capabilities == nil {
		r.capabilities = map[string]Capability{}
	}

	if _, ok := r.capabilities[c.ID()]; ok {
		return fmt.Errorf("capability %s already registered", c.ID())
	}

	r.capabilities[c.ID()] = c
	return nil
}

func (r *Registry) GetCapability(id string) (Capability, error) {
	c, ok := r.capabilities[id]
	if !ok {
		return nil, NoCapability(id)
	}

	return c, nil
}
