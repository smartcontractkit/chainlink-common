// Package di is a dependency injection container whose derived containers can override what their
// ancestors registered, all the way down a dependency chain.
//
// samber/do resolves a provider against the scope it was registered in rather than the scope it was
// invoked from, so a child scope cannot redirect a parent's provider: a parent-provided Chain built
// over a DB keeps the parent's DB however the child overrides it. uber-go/dig behaves the same way.
// That makes the obvious layering - register everything once, override a few things per layer -
// impossible with either of them directly.
//
// Container is that layering. Derive gives a child its own copy of the whole graph, so every
// registration is resolved from the container holding it, and an override reaches arbitrarily far
// down the chains that depend on it.
package di

import (
	"errors"
	"fmt"
	"maps"
	"sync"

	"github.com/samber/do/v2"
)

// Container is a [do.RootScope] plus a live inheritance rule.
//
// [do.Scope] resolves a provider against the scope it was registered in, not the
// scope it was invoked from, so a child scope cannot redirect a parent's
// provider. Container sidesteps that by giving every derived container its own
// copy of the graph via [do.RootScope.Clone], so a registration is always
// resolved from the container holding it, however deeply nested its dependents
// are. Derive then keeps that copy current: later registrations on an ancestor
// still reach it, unless a nearer container has already claimed the type.
//
// The underlying injector is deliberately unexported. Container's own
// bookkeeping decides what is registered where, and handing out the injector
// would let callers bypass it.
type Container struct {
	mu  sync.Mutex
	inj *do.RootScope

	// depth is 0 at the root and one deeper per Derive.
	depth int

	// setAt records, per service name, the depth of the container whose
	// registration currently owns it. A registration cascading down from depth d
	// is applied wherever the resident registration is no deeper than d, so a
	// nearer ancestor always outranks a further one and a container's own
	// registration outranks every ancestor.
	setAt map[string]int

	children []*Container
}

// Resolver is the resolution context passed to a Provider. It carries the
// [do.Injector] samber/do is currently building in, which is what keeps
// samber's circular dependency detection and shutdown ordering working.
type Resolver struct {
	i do.Injector
}

// Scope is something Invoke can resolve from: a Container, or the Resolver a
// Provider is running in.
//
// Resolving through the Resolver is what a Provider should do, so that what it
// builds joins the graph being built rather than starting a second one.
type Scope interface{ injector() do.Injector }

func (c *Container) injector() do.Injector { return c.inj }
func (r Resolver) injector() do.Injector   { return r.i }

// Provider is [do.Provider] over a Resolver.
type Provider[T any] func(Resolver) (T, error)

func New() *Container {
	return &Container{inj: do.New(), setAt: map[string]int{}}
}

// Derive returns a child seeded with everything this container currently has,
// which also keeps receiving this container's later registrations. It is
// [do.RootScope.Clone] plus that ongoing subscription.
func (c *Container) Derive() *Container {
	c.mu.Lock()
	defer c.mu.Unlock()

	child := &Container{
		inj:   c.inj.Clone(),
		depth: c.depth + 1,
		setAt: maps.Clone(c.setAt),
	}
	c.children = append(c.children, child)
	return child
}

// Provide registers a lazy provider for T. See [do.Provide].
//
// Unlike [do.Provide], shadowing a T that came from an ancestor is not an error,
// so callers never have to know whether an ancestor happened to register it.
// Only registering T twice on this same container panics.
func (c *Container) Provide[T any](provider Provider[T]) {
	c.register(do.NameOf[T](), applyProvider(provider), true)
}

// ProvideValue registers an already built T. See [do.ProvideValue].
//
// The instance is shared with every container this cascades into.
func (c *Container) ProvideValue[T any](value T) {
	c.register(do.NameOf[T](), applyValue(value), true)
}

// Override registers a lazy provider for T, replacing whatever is resident. See
// [do.Override].
//
// It also cascades to derived containers that have not registered their own T,
// and, like [do.Override], will not revisit a T that has already been built.
func (c *Container) Override[T any](provider Provider[T]) {
	c.register(do.NameOf[T](), applyProvider(provider), false)
}

// OverrideValue registers an already built T, replacing whatever is resident.
// See [do.OverrideValue]. It cascades on the same terms as Override.
func (c *Container) OverrideValue[T any](value T) {
	c.register(do.NameOf[T](), applyValue(value), false)
}

// Invoke resolves T from this container. See [do.Invoke].
func (c *Container) Invoke[T any]() (T, error) { return do.Invoke[T](c.inj) }

// MustInvoke resolves T from this container. See [do.MustInvoke].
func (c *Container) MustInvoke[T any]() T { return do.MustInvoke[T](c.inj) }

// Invoke resolves T from the container currently being built. See [do.Invoke].
func (r Resolver) Invoke[T any]() (T, error) { return do.Invoke[T](r.i) }

// MustInvoke resolves T from the container currently being built. See [do.MustInvoke].
func (r Resolver) MustInvoke[T any]() T { return do.MustInvoke[T](r.i) }

// Resolve is Invoke for a caller holding a Scope rather than a concrete Container or Resolver.
//
// It exists because Go has no generic methods on interfaces: a helper written against Scope - one
// that serves a provider and a registration alike, as the standalone package's do - cannot call
// s.Invoke. Prefer the methods; reach for this only when the receiver is a Scope.
func Resolve[T any](s Scope) (T, error) { return do.Invoke[T](s.injector()) }

// Shutdown shuts down derived containers first, then this one. See
// [do.RootScope.Shutdown].
//
// A value registered with ProvideValue is held by every container it cascaded
// into, and each shuts down what it holds, so such a value's shutdown hook runs
// once per container rather than once overall.
func (c *Container) Shutdown() error {
	c.mu.Lock()
	children := append([]*Container(nil), c.children...)
	c.mu.Unlock()

	var errs []error
	for _, child := range children {
		if err := child.Shutdown(); err != nil {
			errs = append(errs, err)
		}
	}
	if report := c.inj.Shutdown(); report != nil && len(report.Errors) > 0 {
		errs = append(errs, report)
	}
	return errors.Join(errs...)
}

// applyProvider and applyValue always use the do.Override family so a
// registration can land on a container that inherited the type from Derive
// without tripping samber's "already declared" panic. Whether a registration is
// legal is Container's decision, not the injector's.
func applyProvider[T any](provider Provider[T]) func(do.Injector) {
	return func(i do.Injector) {
		do.Override(i, func(scope do.Injector) (T, error) {
			return provider(Resolver{i: scope})
		})
	}
}

func applyValue[T any](value T) func(do.Injector) {
	return func(i do.Injector) { do.OverrideValue(i, value) }
}

func (c *Container) register(name string, apply func(do.Injector), rejectDuplicate bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if resident, ok := c.setAt[name]; rejectDuplicate && ok && resident == c.depth {
		panic(fmt.Errorf("container: service `%s` is already registered on this container", name))
	}

	apply(c.inj)
	c.setAt[name] = c.depth
	for _, child := range c.children {
		child.cascade(name, apply, c.depth)
	}
}

func (c *Container) cascade(name string, apply func(do.Injector), setterDepth int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// A resident registration from deeper than the setter is more specific, so it
	// wins. Everything below is at least as deep, so the whole subtree is settled
	// and there is nothing left to visit.
	if resident, ok := c.setAt[name]; ok && resident > setterDepth {
		return
	}

	apply(c.inj)
	c.setAt[name] = setterDepth
	for _, child := range c.children {
		child.cascade(name, apply, setterDepth)
	}
}
