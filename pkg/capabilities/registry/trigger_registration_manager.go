package registry

import (
	"context"
	"errors"
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// Caches all trigger registrations and replays them when the underlying capability is updated.
// Owns channels passed to the higher layer (Engine or Don2Don) and goroutines forwarding events
// from the underlying capability.
// Thread-safe: mu guards the regs map, while each registration owns a separate lock, so that
// (un)registrations of different triggers can proceed concurrently. Operations on the same
// trigger are serialized by its registration lock.
type triggerRegistrationManager struct {
	lggr logger.Logger
	mu   sync.Mutex // guards regs
	regs map[string]*triggerRegistration
}

type triggerRegistration struct {
	mu      sync.Mutex // serializes registration lifecycle operations (register/unregister/rebind) for this trigger and guards the fields below
	request capabilities.TriggerRegistrationRequest
	outCh   chan capabilities.TriggerResponse
	cancel  context.CancelFunc // used to shut down the forwarding goroutine when the trigger is unregistered
	done    chan struct{}      // closed by forwarding goroutine upon exit
}

func newTriggerRegistrationManager(lggr logger.Logger) *triggerRegistrationManager {
	return &triggerRegistrationManager{
		lggr: lggr,
		regs: make(map[string]*triggerRegistration),
	}
}

// getOrAddReg returns the registration stored for id, creating an empty one if none exists yet.
func (m *triggerRegistrationManager) getOrAddReg(id string) *triggerRegistration {
	m.mu.Lock()
	defer m.mu.Unlock()
	reg, ok := m.regs[id]
	if !ok {
		reg = &triggerRegistration{}
		m.regs[id] = reg
	}
	return reg
}

// isCurrentReg reports whether reg is still the registration stored for id.
func (m *triggerRegistrationManager) isCurrentReg(id string, reg *triggerRegistration) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.regs[id] == reg
}

// removeReg removes reg from the manager if it is still the registration stored for id and
// reports whether it was removed. The caller must hold reg.mu.
func (m *triggerRegistrationManager) removeReg(id string, reg *triggerRegistration) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.regs[id] != reg {
		return false
	}
	delete(m.regs, id)
	return true
}

// register registers the trigger with the underlying capability and caches the registration.
// The call to the underlying capability is made while holding only this trigger's registration
// lock, so registrations of different triggers can proceed concurrently.
func (m *triggerRegistrationManager) register(ctx context.Context, underlying capabilities.TriggerExecutable, req capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	reg := m.lockCurrentReg(req.TriggerID) // Engine sets TriggerID to (workflowID, triggerIndex)
	defer reg.mu.Unlock()
	return m.registerUnderlying(ctx, underlying, req, reg)
}

// lockCurrentReg returns the current registration for id, locked. If the registration is
// replaced while waiting for its lock (e.g. concurrent unregister and register for the same
// trigger), it retries with the current one.
func (m *triggerRegistrationManager) lockCurrentReg(id string) *triggerRegistration {
	for {
		reg := m.getOrAddReg(id)
		reg.mu.Lock()
		if m.isCurrentReg(id, reg) {
			return reg
		}
		reg.mu.Unlock()
	}
}

// registerUnderlying registers the trigger with the underlying capability and caches the
// registration. reg must be current and locked.
func (m *triggerRegistrationManager) registerUnderlying(ctx context.Context, underlying capabilities.TriggerExecutable, req capabilities.TriggerRegistrationRequest, reg *triggerRegistration) (<-chan capabilities.TriggerResponse, error) {
	in, err := underlying.RegisterTrigger(ctx, req)
	if err != nil {
		// During migration of the nodes to support trigger registration messages it is possible that the registration status
		// cannot be determined as the remote node may not yet support registration messages.  In this case, to be backwards
		// compatible, we optimistically assume the registration succeeded and return a channel, this matches legacy behaviour.
		if errors.Is(err, capabilities.ErrUnableToDetermineRegistrationStatus) {
			return reg.upsert(req, in), err
		}

		// Drop the registration created by getOrAddReg if no earlier registration is cached,
		// so that it is not replayed on rebind.
		if reg.outCh == nil {
			m.removeReg(req.TriggerID, reg)
		}
		return nil, err
	}
	return reg.upsert(req, in), nil
}

// unregister removes the cached registration (if any) and unregisters the trigger with the
// underlying capability. Only this trigger's registration lock is held, so it does not block
// (un)registrations of other triggers.
func (m *triggerRegistrationManager) unregister(ctx context.Context, underlying capabilities.TriggerExecutable, req capabilities.TriggerRegistrationRequest) error {
	m.mu.Lock()
	reg, ok := m.regs[req.TriggerID]
	m.mu.Unlock()

	if ok {
		reg.mu.Lock()
		removed := m.removeReg(req.TriggerID, reg)
		if removed {
			if reg.cancel != nil {
				reg.cancel()
				<-reg.done
			}
			if reg.outCh != nil {
				close(reg.outCh)
			}
		}
		reg.mu.Unlock()
		if !removed {
			// A concurrent registration has already superseded this one; it now owns the
			// trigger's lifecycle with the underlying capability, so don't tear it down here.
			return nil
		}
	}
	return underlying.UnregisterTrigger(ctx, req)
}

// upsert caches the registration request and (re)starts event forwarding from in to the
// registration's output channel, which is created if it does not exist yet.
// The caller must hold reg.mu.
func (reg *triggerRegistration) upsert(req capabilities.TriggerRegistrationRequest, in <-chan capabilities.TriggerResponse) chan capabilities.TriggerResponse {
	reg.request = req
	if reg.outCh == nil {
		reg.outCh = make(chan capabilities.TriggerResponse)
	}
	if reg.cancel != nil {
		reg.cancel() // shuts down the previous forwarding goroutine
		<-reg.done
		reg.cancel = nil
		reg.done = nil
	}
	if in != nil {
		ctxForward, cancel := context.WithCancel(context.Background())
		reg.cancel = cancel
		reg.done = make(chan struct{})
		go forwardTriggerResponses(ctxForward, in, reg.outCh, func() { close(reg.done) })
	}
	return reg.outCh
}

func (m *triggerRegistrationManager) rebind(newUnderlying capabilities.TriggerExecutable) {
	m.mu.Lock()
	regs := make([]*triggerRegistration, 0, len(m.regs))
	for _, reg := range m.regs {
		regs = append(regs, reg)
	}
	m.mu.Unlock()

	for _, reg := range regs {
		reg.mu.Lock()
		if !m.isCurrentReg(reg.request.TriggerID, reg) {
			// The registration was unregistered while we were rebinding.
			reg.mu.Unlock()
			continue
		}
		if reg.outCh == nil {
			// Never successfully registered (e.g. the registration attempt failed after
			// creating the placeholder); nothing to replay.
			reg.mu.Unlock()
			continue
		}

		var in <-chan capabilities.TriggerResponse
		var err error
		if newUnderlying != nil {
			// NOTE: this is tricky - if an existing registration fails on rebind, there is no way to notify the user ...
			in, err = newUnderlying.RegisterTrigger(context.Background(), reg.request)
		}
		if err != nil {
			m.lggr.Errorw("failed to rebind trigger registration", "triggerID", reg.request.TriggerID, "err", err)
		} else {
			m.lggr.Debugw("rebind trigger registration", "triggerID", reg.request.TriggerID)
		}
		reg.upsert(reg.request, in) // user already has this channel
		reg.mu.Unlock()
	}
}

func forwardTriggerResponses(ctx context.Context, in <-chan capabilities.TriggerResponse, out chan<- capabilities.TriggerResponse, done func()) {
	defer done()
	for {
		select {
		case <-ctx.Done():
			return
		case resp, ok := <-in:
			if !ok {
				return
			}
			select {
			case <-ctx.Done():
				return
			case out <- resp:
			}
		}
	}
}
