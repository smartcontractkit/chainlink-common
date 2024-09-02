package store

import (
	"fmt"
	"sync"
)

type capabilitiesStore[T any] struct {
	mu           sync.RWMutex
	capabilities map[string]T
}

type CapabilitiesStore[T any] interface {
	Read(capabilityID string) (value T, ok bool)
	ReadAll() (values map[string]T)
	Write(capabilityID string, value T)
	Delete(capabilityID string)
}

var _ CapabilitiesStore[string] = (CapabilitiesStore[string])(nil)

func NewCapabilitiesStore[T any]() CapabilitiesStore[T] {
	return &capabilitiesStore[T]{
		capabilities: map[string]T{},
	}
}

func (cs *capabilitiesStore[T]) Read(capabilityID string) (value T, ok bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	trigger, ok := cs.capabilities[capabilityID]
	return trigger, ok
}

func (cs *capabilitiesStore[T]) ReadAll() (values map[string]T) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.capabilities
}

func (cs *capabilitiesStore[T]) Write(capabilityID string, value T) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.capabilities[capabilityID] = value
}

func (cs *capabilitiesStore[T]) InsertIfNotExists(capabilityID string, value T) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if _, ok := cs.capabilities[capabilityID]; ok {
		return fmt.Errorf("capabilityID %v already exists in CapabilitiesStore", capabilityID)
	}
	cs.capabilities[capabilityID] = value
	return nil
}

func (cs *capabilitiesStore[T]) Delete(capabilityID string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.capabilities, capabilityID)
}
