// Package standalone is the contract a bootstrap dependency registers against: a container of
// pkg/capabilities/v2/standalone/di, what a dependency can ask that container about itself, and the
// few services the framework driving it always puts there.
//
// Only the contract lives here, so that anything registering into a container - a database in the
// capabilities repo, an EVM client in chainlink-evm - can do so without depending on the framework
// that drives it. The framework itself (the Bootstrapper, its commands, and the containers it
// builds) lives in the capabilities repo.
//
// # One registration, and overrides on top of it
//
// A binary's dependencies live in one root container, and an embedded run derives from it:
//
//	root                 what `run` resolves: every dependency as it is configured
//	└── embed            what an embedded run replaces, registered once
//	    ├── 0            instance 0: nothing but its EmbeddedConfig
//	    └── 1            instance 1, ...
//
// Root and Embed are the two global registries those are built from, and a dependency registers
// itself into them from an init function. So a binary is the sum of what it links rather than a
// list something has to keep in step with it - see Root.
//
// A dependency registers its provider once, into Root. The embed registry holds only what an
// embedded run genuinely replaces, and a derived container's override reaches everything built over
// it however deep: override the database and the chain built on it follows, without the chain being
// registered twice. That is what di.Container is for, and what neither samber/do nor uber-go/dig
// gives you on its own.
//
// An instance registers nothing at all. It is a container derived from the embed one with its
// EmbeddedConfig in it, and a provider that needs to know which instance it is serving asks for
// that when it resolves - see Embedded.
package standalone

import (
	"context"
	"io"
	"sync"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/standalone/di"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// CommonConfig is the process-wide configuration, provided in the root container.
type CommonConfig struct{}

// EmbeddedConfig is the instance a container serves, and the run it is part of. The framework puts
// it in every instance container and in no other, so Embedded is a reliable answer to "am I serving
// an embedded instance".
type EmbeddedConfig struct {
	// NumInstances is how many instances the run has.
	//
	// It is here because the instances of an embedded run are peers of each other rather than
	// copies: naming the set they belong to - the members of a DON, say - is a question no
	// instance can answer from its own index, since it has to name members that have not started
	// yet. Everything else follows from InstanceId alone.
	NumInstances int

	// InstanceId is which instance this container serves, from 0 to NumInstances-1.
	InstanceId int
}

// Embedded reports whether s is serving an embedded instance, and which instance of which run.
//
// Ask it when resolving rather than when registering: a provider registered once on the root is
// resolved from whichever container is asking, so this is how the one registration tells an
// embedded instance apart from a configured run.
func Embedded(s di.Scope) (EmbeddedConfig, bool) {
	cfg, err := di.Resolve[EmbeddedConfig](s)
	return cfg, err == nil
}

// registries holds what each container is built from. Global because a dependency registers itself
// from an init function, and an init function has nothing to be handed.
var registries struct {
	mu    sync.Mutex
	root  []func(*di.Container)
	embed []func(*di.Container)
}

// Root registers fns against the root container - what `run` resolves, and what an embedded run
// derives from.
//
// Call it from an init function, so that linking a package is what puts its dependency in the
// binary:
//
//	func init() { standalone.Root(provide) }
//
// Which means a package must be safe for any binary that links it, including one that links it
// transitively and wants nothing to do with it. A dependency that is not - one of two mutually
// exclusive forms, or one whose settings would be required of a binary with no use for them -
// exports a Register function for the binary's own init to call instead.
//
// Each fn is called once, with the root, before the command line is parsed: registering a
// dependency is also what registers its settings, and a flag has to exist before it can be given a
// value.
func Root(fns ...func(*di.Container)) {
	registries.mu.Lock()
	defer registries.mu.Unlock()
	registries.root = append(registries.root, fns...)
}

// Embed registers fns against the embed container - what an embedded run replaces.
//
// Register only what embedding genuinely changes. Everything else is already there, registered once
// into Root, and an override here reaches whatever is built over it: replacing the database is
// enough to give an instance its own chain, log poller and transaction manager, because those are
// resolved from the container doing the asking.
//
// A dependency that reads the same settings either way registers those here too, with the same
// namespace and type - that is how `embed` comes to offer them. Registering nothing means `embed`
// offers nothing, which is the right answer for a setting an embedded instance derives.
//
// Each fn is called once, with the embed container, before the command line is parsed. The
// instances are derived from that container when the command runs, and register nothing of their
// own.
func Embed(fns ...func(*di.Container)) {
	registries.mu.Lock()
	defer registries.mu.Unlock()
	registries.embed = append(registries.embed, fns...)
}

// RootDependencies and EmbedDependencies are what the framework builds each container from. Nothing
// else should need them.
func RootDependencies() []func(*di.Container) {
	registries.mu.Lock()
	defer registries.mu.Unlock()
	return append([]func(*di.Container){}, registries.root...)
}

func EmbedDependencies() []func(*di.Container) {
	registries.mu.Lock()
	defer registries.mu.Unlock()
	return append([]func(*di.Container){}, registries.embed...)
}

// SwapDependencies empties both registries and returns a function putting back what was there.
//
// It is for a test that needs the registries to hold only what it registers, which a global cannot
// otherwise offer: see standalonetest.Isolate, and call that rather than this.
func SwapDependencies() (restore func()) {
	registries.mu.Lock()
	defer registries.mu.Unlock()

	root, embed := registries.root, registries.embed
	registries.root, registries.embed = nil, nil

	return func() {
		registries.mu.Lock()
		defer registries.mu.Unlock()
		registries.root, registries.embed = root, embed
	}
}

// Shared registers a provider whose value is built once and shared by every container derived from
// c, instead of once per container.
//
// Deriving a container copies the graph, so an ordinary provider registered on the root builds a
// value per instance - which is right for a database and wrong for a run of ports, an in-process
// gateway, or a simulated chain the instances are all nodes on. The provider itself is shared by
// every copy, so building at most once inside it is what makes the value shared.
//
// provider must not depend on which container called it: capture EmbeddedConfig.NumInstances if the
// shared value needs to know how many instances there are, and nothing else.
func Shared[T any](c *di.Container, provider di.Provider[T]) {
	var (
		once  sync.Once
		value T
		err   error
	)
	c.Provide(func(r di.Resolver) (T, error) {
		once.Do(func() { value, err = provider(r) })
		return value, err
	})
}

// Configs is the settings the command resolving from a container accepts. There is one in the root
// and one in the embed container, and a dependency puts its own there as it registers itself - so
// which command offers a setting is decided by which registry the dependency registered into, and
// the CLI binds what it finds rather than being told.
//
// Settings is how a dependency reaches it.
type Configs interface {
	// Register records cfg's settings under namespace and returns the instance the values are
	// decoded into.
	//
	// namespace roots them, so a dependency's settings group together and same-named settings from
	// different dependencies don't collide - "database" gives --database.url, the key database.url
	// and the env var CRE_DATABASE_URL. Pass "" to keep them at the top level.
	//
	// The instance returned is not always the one passed in. A configuration is decoded once
	// however many containers read it, so a registration the root has already made is handed the
	// root's instance - which is what makes the embed container, registered separately and with
	// defaults of its own, read the values an operator actually typed.
	//
	// Errors are reported by the framework before anything runs rather than returned here, since a
	// registration function has nowhere to put one. cfg comes back unchanged when there is one.
	Register(namespace string, cfg any) any
}

// Settings registers defaults under namespace in c's Configs and returns the instance the
// configuration is decoded into - which may be one the root already registered, since a
// configuration is decoded once however many containers read it.
//
// Call it while registering, not while resolving: the settings have to be bound before the command
// line is parsed, and a provider does not run until well after that.
//
// A container with no Configs in it - a test's, typically - gets defaults back, so a dependency can
// be registered into a bare di.New() and resolved with its compiled-in configuration.
func Settings[T any](c *di.Container, namespace string, defaults *T) *T {
	configs, err := c.Invoke[Configs]()
	if err != nil {
		return defaults
	}
	if cfg, ok := configs.Register(namespace, defaults).(*T); ok {
		return cfg
	}
	return defaults
}

// Closers is where a dependency whose resolved value needs releasing registers it. The framework
// provides one in the root container; Closing is how a dependency reaches it.
type Closers interface {
	// Closing registers c to be closed on shutdown, after the services built from it.
	Closing(c io.Closer)
}

// Closing registers c to be closed when the process shuts down, after the services built from it,
// and returns it so a provider can end with `return Closing(r, db), nil`.
//
// do shuts down a service implementing one of its own Shutdowner interfaces, and almost nothing in
// this stack does: what they implement is io.Closer. This is that, registered explicitly.
//
// A container with no Closers in it - a test's - registers nothing and the value is the caller's to
// close, which is what a test wants anyway.
func Closing[T io.Closer](s di.Scope, c T) T {
	if closers, err := di.Resolve[Closers](s); err == nil {
		closers.Closing(c)
	}
	return c
}

// Logger is the logger of the instance resolving a dependency, named after that dependency.
//
// It comes from the container rather than from an argument to every registration function because
// it is the instance's: under `embed` each instance logs under a name of its own, and threading
// that through every constructor meant every binary had to know which instance it was building for.
//
// A container with no logger in it gets a nop one, so a dependency resolved in a test logs nowhere
// rather than failing to resolve.
func Logger(s di.Scope, name string) logger.SugaredLogger {
	lggr, err := di.Resolve[logger.SugaredLogger](s)
	if err != nil {
		lggr = logger.Sugared(logger.Nop())
	}
	if name == "" {
		return lggr
	}
	return logger.Sugared(logger.Named(lggr, name))
}

// Context is the context a dependency is resolved under: the process's own, cancelled when the
// binary is shutting down.
//
// It is here because resolving a dependency is startup work that talks to things - opening a
// database and running its migrations, dialling a keystore to ask which accounts it holds, binding
// a listener - and all of it belongs to the one context the binary was started with, so an
// interrupt part-way through starting up abandons it rather than blocking on it.
//
// A container with no context in it - a test's - gets context.Background, since there is no run to
// belong to.
func Context(s di.Scope) context.Context {
	if ctx, err := di.Resolve[context.Context](s); err == nil {
		return ctx
	}
	return context.Background()
}
