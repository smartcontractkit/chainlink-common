// Package standalone is the contract a bootstrap dependency implements: something a standalone CRE
// binary can be handed, configure from flags, and resolve into a value its services use.
//
// Only the contract lives here, so that anything implementing it - a database in the capabilities
// repo, an EVM client in chainlink-evm - can do so without depending on the framework that drives
// it. The framework itself (the Bootstrapper, its commands and the Run helpers that resolve these)
// lives in the capabilities repo.
package standalone

import (
	"context"
	"sync"
)

// CommonConfig is the process-wide configuration every dependency is handed when it is resolved.
// It says nothing about which instance is resolving it: by then a dependency has already been
// replaced by the form that serves that instance, so there is no decision left for it to make.
type CommonConfig struct{}

type BootstrapCommand interface {
	// Config returns the settings to bind, as a pointer to the struct they are decoded into, or nil
	// when there are none - which is the usual answer from an embedded dependency, having derived or
	// replaced everything it would otherwise be told.
	Config() any

	Dependencies() []BootstrapCommand

	// Namespace roots this dependency's configuration, so its settings group together and
	// same-named settings from different dependencies don't collide - "database" gives
	// --database.url, the key database.url and the env var CRE_DATABASE_URL. Return "" to
	// keep the settings at the top level.
	Namespace() string
}

type BootstrapDependency[T any] interface {
	Get(ctx context.Context, c CommonConfig) (T, error)

	// ForEmbedding returns the dependency instance i of an embedded run resolves instead of this
	// one. It is called once per instance, after the configuration has been decoded and before
	// anything is resolved, and only for an embedded run - a single instance resolves the
	// dependency exactly as configured.
	//
	// What it returns is already specific to instance i: everything embedding changes - a derived
	// identity instead of a stored one, an in-process transport instead of a socket, a schema of
	// its own instead of the shared one - is settled here, so Get has no instance to ask about and
	// no mode to branch on. A dependency whose embedded form needs none of its settings is free to
	// return something else entirely (see ocr.Host), which is also how a setting that only a real
	// deployment needs stops being required.
	//
	// Return the receiver to be shared by every instance, which is what a dependency backed by
	// one process-wide resource wants: sharing the dependency shares the single value it resolves
	// to. Anything else returns a copy, deep-copying the configuration it adapts so that instances
	// never write through to each other's settings.
	ForEmbedding(i int) BootstrapDependency[T]

	BootstrapCommand
}

// OnceBootstrapper wraps a BootstrapDependency so that Get is evaluated at most
// onceGet: the first call resolves the dependency and caches its (value, error),
// and every subsequent call returns that same result without re-running Get
// (the ctx and CommonConfig of later calls are ignored). Other commands are all delegated
//
// BootstrapDependency implementations are shared and may have Get called more
// than onceGet — e.g. one dependency resolving another, or the same dependency
// feeding several services — so a New function should wrap its dependency with
// OnceBootstrapper before returning it, making repeated Get calls safe and
// side-effect-free.
func OnceBootstrapper[T any](bd BootstrapDependency[T]) BootstrapDependency[T] {
	return &onceBootstrapper[T]{BootstrapDependency: bd}
}

type onceBootstrapper[T any] struct {
	BootstrapDependency[T]
	once sync.Once
	val  T
	err  error
}

func (o *onceBootstrapper[T]) Get(ctx context.Context, c CommonConfig) (T, error) {
	o.once.Do(func() { o.val, o.err = o.BootstrapDependency.Get(ctx, c) })
	return o.val, o.err
}

// ForEmbedding gives each instance its own cache, since each resolves its own value - unless the
// wrapped dependency returns itself, meaning it is shared by every instance, in which case this
// wrapper (and the one value it caches) is shared too.
func (o *onceBootstrapper[T]) ForEmbedding(i int) BootstrapDependency[T] {
	next := o.BootstrapDependency.ForEmbedding(i)
	if next == o.BootstrapDependency {
		return o
	}
	return OnceBootstrapper[T](next)
}
