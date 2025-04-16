package testutils

import (
	"fmt"
	"testing"
)

var testRegistries = map[testing.TB]*Registry{}

func GetRegistry(t testing.TB) *Registry {
	if r, ok := testRegistries[t]; ok {
		return r
	}
	r := &Registry{}
	testRegistries[t] = r
	t.Cleanup(func() {
		delete(testRegistries, t)
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
