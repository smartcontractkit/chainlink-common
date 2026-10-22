// Package standalonetest builds the dependency injection containers a standalone binary's
// bootstrapper builds, for testing a dependency's registration function and the providers it
// registers.
//
// A dependency is written against a container and the few services the framework puts in it, so a
// test of one needs a container with those services. That is all this is: Root and Embedded stand
// in for the two the bootstrapper builds, with the settings the test wants decoded rather than the
// ones a command line would have supplied.
package standalonetest

import (
	"context"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/standalone"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/standalone/di"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// Root is a container like the one `run` resolves from: no EmbeddedConfig, so a dependency
// registered into it resolves its configured form.
//
// settings are the configurations the test wants those dependencies decoded into, as pointers,
// matched to a registration by type - so a test says what the command line would have said and
// then resolves the dependency that reads it. A dependency whose configuration is not listed gets
// the defaults it registered.
func Root(tb testing.TB, pkgs []func(*di.Container), settings ...any) *di.Container {
	c := base(tb, settings)
	for _, pkg := range pkgs {
		pkg(c)
	}
	return c
}

// Embedded is a container like the one instance cfg.InstanceId of `embed` resolves from: root
// holds rootPkgs, an embed container derived from it holds embedPkgs, and the instance derived from
// that holds cfg.
//
// It is the bootstrapper's shape, because that is what decides what an override reaches: embedPkgs
// override what rootPkgs registered, and anything built over what they replaced follows.
func Embedded(
	tb testing.TB,
	cfg standalone.EmbeddedConfig,
	rootPkgs, embedPkgs []func(*di.Container),
	settings ...any,
) *di.Container {
	root := base(tb, settings)
	for _, pkg := range rootPkgs {
		pkg(root)
	}

	embed := root.Derive()
	for _, pkg := range embedPkgs {
		pkg(embed)
	}

	instance := embed.Derive()
	instance.ProvideValue(cfg)
	return instance
}

// LinkedRoot is Root built from the root registry rather than from a list: every dependency the
// test binary links has registered itself by the time a test runs, which is what a real binary's
// root holds.
func LinkedRoot(tb testing.TB, settings ...any) *di.Container {
	return Root(tb, standalone.RootDependencies(), settings...)
}

// LinkedEmbedded is Embedded built from both registries. See LinkedRoot.
func LinkedEmbedded(tb testing.TB, cfg standalone.EmbeddedConfig, settings ...any) *di.Container {
	return Embedded(tb, cfg, standalone.RootDependencies(), standalone.EmbedDependencies(), settings...)
}

// base is a root container with the services the framework always provides.
func base(tb testing.TB, settings []any) *di.Container {
	c := di.New()
	c.ProvideValue[standalone.Configs](&configs{settings: settings})
	c.ProvideValue[standalone.Closers](&closers{tb: tb})
	c.ProvideValue(logger.TestSugared(tb))
	// A registry of the test's own: the default one is process-wide, and two tests registering the
	// same collector on it is an error rather than two independent tests.
	c.ProvideValue[prometheus.Registerer](prometheus.NewRegistry())
	c.ProvideValue(standalone.CommonConfig{})
	c.ProvideValue(http.NewServeMux())
	c.Provide(func(di.Resolver) (context.Context, error) { return context.Background(), nil })

	tb.Cleanup(func() { _ = c.Shutdown() })
	return c
}

// configs is standalone.Configs over what the test supplied: a registration is handed the listed
// configuration of its own type, and its own defaults when there is none.
type configs struct{ settings []any }

func (c *configs) Register(_ string, cfg any) any {
	want := reflect.TypeOf(cfg)
	for _, s := range c.settings {
		if reflect.TypeOf(s) == want {
			return s
		}
	}
	return cfg
}

// closers closes what the resolved dependencies registered when the test ends, so a test that
// resolves a database or a listener does not have to remember to.
type closers struct{ tb testing.TB }

func (c *closers) Closing(v io.Closer) {
	c.tb.Cleanup(func() {
		if err := v.Close(); err != nil {
			c.tb.Errorf("failed to close a resolved dependency: %s", err)
		}
	})
}

// Isolate empties the global dependency registries for the duration of the test, so what a test
// registers is all a container built from them holds - and so a test's registrations do not leak
// into the next one.
//
// Only a test of the framework itself needs this. A test of one dependency builds a container from
// that dependency's registration function directly, with Root or Embedded.
func Isolate(tb testing.TB) {
	tb.Cleanup(standalone.SwapDependencies())
}
