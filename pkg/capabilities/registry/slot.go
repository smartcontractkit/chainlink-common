package registry

import (
	"context"
	"errors"
	"sync"
	"time"

	"google.golang.org/grpc/connectivity"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

const (
	// defaultReplaceGracePeriod is how long the calls in flight on a capability get to
	// finish once that capability is replaced or removed. After it they are cancelled.
	defaultReplaceGracePeriod = 5 * time.Second
	// defaultReplayTimeout bounds each trigger registration that is replayed onto a new
	// capability.
	defaultReplayTimeout = 30 * time.Second
)

// ErrCapabilityReplaced is the cause of the cancellation of a call that was still in
// flight on a capability after the registry replaced or removed that capability.
var ErrCapabilityReplaced = errors.New("capability was replaced or removed")

var errCapabilityUnavailable = errors.New("capability unavailable")

// stillUpError is returned by slot.replace when the capability must be down to be
// replaced, and it is not.
type stillUpError struct {
	state connectivity.State
}

func (e *stillUpError) Error() string {
	return "capability is still in use: state " + e.state.String()
}

// ctxMutex is a mutex whose lock gives up when a context ends. Waiters get it in the
// order they arrived.
type ctxMutex chan struct{}

func newCtxMutex() ctxMutex { return make(chan struct{}, 1) }

func (m ctxMutex) lock(ctx context.Context) error {
	select {
	case m <- struct{}{}:
		return nil
	default:
	}
	select {
	case m <- struct{}{}:
		return nil
	case <-ctx.Done():
		return context.Cause(ctx)
	}
}

func (m ctxMutex) unlock() { <-m }

// callScope ends the calls that are in flight on one capability.
type callScope struct {
	// ctx is the lifetime of the calls on one capability. It is not the context of a call.
	ctx    context.Context //nolint:containedctx // see the comment above
	cancel context.CancelCauseFunc
}

func newCallScope() *callScope {
	ctx, cancel := context.WithCancelCause(context.Background())
	return &callScope{ctx: ctx, cancel: cancel}
}

// bind returns a context that ends when ctx ends, and also when the scope is ended.
// The returned func must be called when the call is over.
func (s *callScope) bind(ctx context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancelCause(ctx)
	stop := context.AfterFunc(s.ctx, func() { cancel(context.Cause(s.ctx)) })
	return ctx, func() {
		stop()
		cancel(nil)
	}
}

// slot holds one capability that can be replaced while it is in use.
//
// A call into a capability that no longer answers lasts as long as its caller allows.
// That can be forever: a LOOP client keeps reconnecting to a plugin that is gone, and
// not every caller sets a deadline. So every call made through a slot is bound to a
// scope, and the scope is ended when the capability is replaced.
type slot[T capabilities.BaseCapability] struct {
	// replacing serialises replacements.
	replacing ctxMutex

	// mu guards the fields below. It is never held across a call into the capability.
	mu      sync.RWMutex
	current T
	set     bool
	scope   *callScope
}

func newSlot[T capabilities.BaseCapability]() *slot[T] {
	return &slot[T]{replacing: newCtxMutex(), scope: newCallScope()}
}

// load returns the capability and the scope for a call into it. ok is false when the
// slot is empty.
func (s *slot[T]) load() (current T, scope *callScope, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.current, s.scope, s.set
}

// call returns the capability and the context for one call into it. The context also
// ends when the capability is replaced: see callScope. done must be called when the call
// is over.
//
// call fails when the slot is empty. It also fails, with the cause, when the calls of the
// capability have been ended already: the capability is being replaced, and a new call
// into it could only wait for its own cancellation. The client of a plugin that is gone
// finishes an attempt to connect before it observes its context, so calls that wait in
// line for a capability must not each take that turn.
func (s *slot[T]) call(ctx context.Context) (current T, callCtx context.Context, done context.CancelFunc, err error) {
	current, scope, ok := s.load()
	if !ok {
		return current, nil, nil, errCapabilityUnavailable
	}
	if scope.ctx.Err() != nil {
		return current, nil, nil, context.Cause(scope.ctx)
	}
	callCtx, done = scope.bind(ctx)
	return current, callCtx, done, nil
}

func (s *slot[T]) state() connectivity.State {
	current, _, ok := s.load()
	if !ok {
		return connectivity.Shutdown
	}
	if sg, ok := any(current).(StateGetter); ok {
		return sg.GetState()
	}
	return connectivity.State(-1) // unknown
}

// replace puts next in the slot, or empties the slot when set is false.
//
// The calls in flight on the current capability get grace to finish. After that they
// are cancelled with [ErrCapabilityReplaced], so a capability that no longer answers
// cannot hold up its replacement.
//
// regLock, when it is not nil, is the lock that trigger registration calls hold while
// they are in flight. replace waits for it, and on success returns with it held.
//
// With onlyIfDown, a capability whose connection is up (or whose state is not known) is
// left alone, and replace returns a *stillUpError.
func (s *slot[T]) replace(ctx context.Context, next T, set, onlyIfDown bool, regLock ctxMutex, grace time.Duration) error {
	if err := s.replacing.lock(ctx); err != nil {
		return err
	}
	defer s.replacing.unlock()

	if onlyIfDown {
		switch state := s.state(); state {
		case connectivity.Shutdown, connectivity.TransientFailure, connectivity.Idle, connectivity.Connecting:
			// Allow replace. Connecting is in the list because a client whose peer is gone
			// alternates between TransientFailure and Connecting for as long as it retries:
			// without it, whether the peer's replacement is accepted depends on the instant.
		default:
			return &stillUpError{state: state}
		}
	}

	_, old, _ := s.load()
	abort := time.AfterFunc(grace, func() { old.cancel(ErrCapabilityReplaced) })
	if regLock != nil {
		if err := regLock.lock(ctx); err != nil {
			// Not replaced. If the calls in flight were ended, later calls need a new scope.
			if !abort.Stop() {
				s.mu.Lock()
				s.scope = newCallScope()
				s.mu.Unlock()
			}
			return err
		}
	}

	// Calls that are still in flight on the old capability keep the rest of the grace
	// period: abort is left to fire.
	s.mu.Lock()
	s.current, s.set, s.scope = next, set, newCallScope()
	s.mu.Unlock()
	return nil
}
