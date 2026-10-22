package limits

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

// TestErrorMissingTenant asserts that every scoped limiter reports a missing required tenant
// as ErrMissingTenant, carrying the scope that was missing. Callers need to tell this apart from a settings read failure: a read
// failure still resolves a usable value, whereas here no lookup happens at all, so the
// returned value is meaningless (zero for most limiters) and must not be enforced as a limit.
func TestErrorMissingTenant(t *testing.T) {
	t.Parallel()

	// ScopeWorkflow requires a tenant; no contexts.WithCRE below, so none is present.
	const scope = settings.ScopeWorkflow
	newFactory := func(t *testing.T) Factory { return Factory{Logger: logger.Test(t)} }

	limitOf := map[string]func(t *testing.T) (func(context.Context) error, func() error){
		"bound": func(t *testing.T) (func(context.Context) error, func() error) {
			s := settings.Size(1 * config.GByte)
			s.Key, s.Scope = "test.missing.bound", scope
			l, err := MakeUpperBoundLimiter(newFactory(t), s)
			require.NoError(t, err)
			return func(ctx context.Context) error { _, e := l.Limit(ctx); return e }, l.Close
		},
		"range": func(t *testing.T) (func(context.Context) error, func() error) {
			s := settings.NewSetting(settings.Range[int]{Lower: 1, Upper: 5},
				settings.ParseRangeFn(func(string) (int, error) { return 0, nil }))
			s.Key, s.Scope = "test.missing.range", scope
			l, err := MakeRangeLimiter[int](newFactory(t), s)
			require.NoError(t, err)
			return func(ctx context.Context) error { _, e := l.Limit(ctx); return e }, l.Close
		},
		"gate": func(t *testing.T) (func(context.Context) error, func() error) {
			s := settings.Bool(true)
			s.Key, s.Scope = "test.missing.gate", scope
			l, err := MakeGateLimiter(newFactory(t), s)
			require.NoError(t, err)
			return func(ctx context.Context) error { _, e := l.Limit(ctx); return e }, l.Close
		},
		"time": func(t *testing.T) (func(context.Context) error, func() error) {
			s := settings.Duration(time.Minute)
			s.Key, s.Scope = "test.missing.time", scope
			l, err := newFactory(t).MakeTimeLimiter(s)
			require.NoError(t, err)
			return func(ctx context.Context) error { _, e := l.Limit(ctx); return e }, l.Close
		},
		"queue": func(t *testing.T) (func(context.Context) error, func() error) {
			s := settings.Int(10)
			s.Key, s.Scope = "test.missing.queue", scope
			l, err := MakeQueueLimiter[int](newFactory(t), s)
			require.NoError(t, err)
			return func(ctx context.Context) error { _, e := l.Limit(ctx); return e }, l.Close
		},
		"rate": func(t *testing.T) (func(context.Context) error, func() error) {
			s := settings.Rate(rate.Every(time.Second), 5)
			s.Key, s.Scope = "test.missing.rate", scope
			l, err := newFactory(t).MakeRateLimiter(s)
			require.NoError(t, err)
			return func(ctx context.Context) error { _, e := l.Limit(ctx); return e }, l.Close
		},
		"resource": func(t *testing.T) (func(context.Context) error, func() error) {
			s := settings.Int(10)
			s.Key, s.Scope = "test.missing.resource", scope
			l, err := MakeResourcePoolLimiter(newFactory(t), s)
			require.NoError(t, err)
			return func(ctx context.Context) error { _, e := l.Limit(ctx); return e }, l.Close
		},
	}

	for name, build := range limitOf {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			limit, closeFn := build(t)
			t.Cleanup(func() { assert.NoError(t, closeFn()) })

			err := limit(t.Context())
			require.Error(t, err, "a required but missing tenant must be an error")

			var missing ErrMissingTenant
			require.ErrorAs(t, err, &missing,
				"callers must be able to tell a missing tenant from a settings read failure")
			assert.Equal(t, scope, missing.Scope, "the error must carry the scope that was missing")
			assert.False(t, IsErrRecoverable(err), "no value was resolved, so there is nothing to fall back to")
		})
	}
}

// TestErrMissingTenant_OnlyForRequiredScopes: ScopeOrg does not require a tenant, so a
// missing one there is not a programming error and must not carry the sentinel. Queues are
// the only scoped limiter with no unlimited instance to fail open to, so they still error —
// just not as ErrMissingTenant.
func TestErrorMissingTenant_OnlyForRequiredScopes(t *testing.T) {
	t.Parallel()
	require.False(t, settings.ScopeOrg.IsTenantRequired(), "precondition: org scope is optional")

	s := settings.Int(10)
	s.Key, s.Scope = "test.optional.queue", settings.ScopeOrg
	q, err := MakeQueueLimiter[int](Factory{Logger: logger.Test(t)}, s)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, q.Close()) })

	_, err = q.Limit(t.Context()) // no contexts.WithCRE, so no org tenant
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrMissingTenant{},
		"an optional scope missing its tenant is not a programming error")
}

// TestErrMissingTenant_NotALimitError guards the categorisation: a missing tenant is a
// programming error, not a breached limit, so it must not satisfy LimitError (which carries
// gRPC ResourceExhausted/PermissionDenied semantics).
func TestErrorMissingTenant_NotALimitError(t *testing.T) {
	t.Parallel()

	var limitErr LimitError
	assert.NotErrorAs(t, ErrMissingTenant{Scope: settings.ScopeWorkflow}, &limitErr)
}

// TestIsRecoverable pins the predicate callers are told to use instead of matching concrete
// error types, so that adding a non-recoverable case later does not silently change meaning.
func TestIsRecoverable(t *testing.T) {
	t.Parallel()

	assert.True(t, IsErrRecoverable(nil), "no error at all is trivially recoverable")
	assert.True(t, IsErrRecoverable(errGetterUnavailable),
		"a settings read failure still leaves the compiled default")
	assert.True(t, IsErrRecoverable(fmt.Errorf("wrapped: %w", errGetterUnavailable)))

	assert.False(t, IsErrRecoverable(ErrMissingTenant{Scope: settings.ScopeWorkflow}))
	assert.False(t, IsErrRecoverable(fmt.Errorf("wrapped: %w", ErrMissingTenant{Scope: settings.ScopeOwner})),
		"must match through wrapping, since limiters add context")
}
