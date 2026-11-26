package registry

import (
	"context"
	"errors"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// Caches all trigger registrations and replays them when the underlying capability is updated.
// Owns channels passed to the higher layer (Engine or Don2Don) and goroutines forwarding events
// from the underlying capability.
// NOT thread-safe - the caller is responsible for locking.
type triggerRegistrationManager struct {
	lggr logger.Logger
	regs map[string]*triggerRegistration
	// replayTimeout bounds each registration that is replayed onto a new capability.
	replayTimeout time.Duration
}

type triggerRegistration struct {
	request capabilities.TriggerRegistrationRequest
	outCh   chan capabilities.TriggerResponse
	cancel  context.CancelFunc // used to shut down the forwarding goroutine when the trigger is unregistered
	done    chan struct{}      // closed by forwarding goroutine upon exit
}

func newTriggerRegistrationManager(lggr logger.Logger, replayTimeout time.Duration) *triggerRegistrationManager {
	return &triggerRegistrationManager{
		lggr:          lggr,
		regs:          make(map[string]*triggerRegistration),
		replayTimeout: replayTimeout,
	}
}

func (m *triggerRegistrationManager) register(ctx context.Context, underlying capabilities.TriggerExecutable, req capabilities.TriggerRegistrationRequest) (<-chan capabilities.TriggerResponse, error) {
	in, err := underlying.RegisterTrigger(ctx, req)
	if err != nil {
		// During migration of the nodes to support trigger registration messages it is possible that the registration status
		// cannot be determined as the remote node may not yet support registration messages.  In this case, to be backwards
		// compatible, we optimistically assume the registration succeeded and return a channel, this matches legacy behaviour.
		if errors.Is(err, capabilities.ErrUnableToDetermineRegistrationStatus) {
			return m.upsertRegistration(req, nil, in), err
		}

		return nil, err
	}
	return m.upsertRegistration(req, nil, in), nil
}

func (m *triggerRegistrationManager) unregister(ctx context.Context, underlying capabilities.TriggerExecutable, req capabilities.TriggerRegistrationRequest) error {
	if reg, ok := m.regs[req.TriggerID]; ok {
		if reg.cancel != nil {
			reg.cancel()
			<-reg.done
		}
		if reg.outCh != nil {
			close(reg.outCh)
		}
		delete(m.regs, req.TriggerID)
	}
	return underlying.UnregisterTrigger(ctx, req)
}

func (m *triggerRegistrationManager) upsertRegistration(req capabilities.TriggerRegistrationRequest, outCh chan capabilities.TriggerResponse, in <-chan capabilities.TriggerResponse) chan capabilities.TriggerResponse {
	registrationID := req.TriggerID // Engine sets it to (workflowID, triggerIndex)
	regInMap, ok := m.regs[registrationID]
	if !ok {
		if outCh == nil {
			outCh = make(chan capabilities.TriggerResponse)
		}
		regInMap = &triggerRegistration{
			request: req,
			outCh:   outCh,
		}
		m.regs[registrationID] = regInMap
	} else {
		regInMap.request = req
		if outCh != nil {
			regInMap.outCh = outCh
		}
		if regInMap.cancel != nil {
			regInMap.cancel() // shuts down the previous forwarding goroutine
			<-regInMap.done
			regInMap.cancel = nil
			regInMap.done = nil
		}
	}
	if in != nil {
		ctxForward, cancel := context.WithCancel(context.Background())
		regInMap.cancel = cancel
		regInMap.done = make(chan struct{})
		go forwardTriggerResponses(ctxForward, in, regInMap.outCh, func() { close(regInMap.done) })
	}
	return regInMap.outCh
}

// rebind points the cached registrations at a new underlying capability, or at none.
//
// The caller holds lock, and rebind releases it: at once when there is nothing to replay,
// otherwise when the replay is over. The replay runs in the background, so that adding a
// capability does not take longer the more registrations it inherits. Until it is over,
// registration changes wait for lock. scope is the scope of the calls into newUnderlying:
// the replay stops when it ends, which is when newUnderlying is itself replaced.
func (m *triggerRegistrationManager) rebind(lock ctxMutex, scope *callScope, newUnderlying capabilities.TriggerExecutable) {
	if newUnderlying == nil {
		for _, reg := range m.regs {
			_ = m.upsertRegistration(reg.request, reg.outCh, nil) // stops forwarding from the previous capability
		}
		lock.unlock()
		return
	}
	if len(m.regs) == 0 {
		lock.unlock()
		return
	}
	go func() {
		defer lock.unlock()
		m.replay(scope, newUnderlying)
	}()
}

func (m *triggerRegistrationManager) replay(scope *callScope, newUnderlying capabilities.TriggerExecutable) {
	replayed := 0
	for _, reg := range m.regs {
		if scope.ctx.Err() != nil {
			m.lggr.Warnw("stopped rebinding trigger registrations: the capability was replaced again",
				"rebound", replayed, "total", len(m.regs))
			return
		}
		ctx, cancel := context.WithTimeout(scope.ctx, m.replayTimeout)
		// NOTE: this is tricky - if an existing registration fails on rebind, there is no way to notify the user ...
		in, err := newUnderlying.RegisterTrigger(ctx, reg.request)
		cancel()
		if err != nil {
			m.lggr.Errorw("failed to rebind trigger registration", "triggerID", reg.request.TriggerID, "err", err)
		} else {
			m.lggr.Debugw("rebind trigger registration", "triggerID", reg.request.TriggerID)
		}
		_ = m.upsertRegistration(reg.request, reg.outCh, in) // user already has this channel
		replayed++
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
