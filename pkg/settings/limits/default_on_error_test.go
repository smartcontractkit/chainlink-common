package limits

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

var errGetterUnavailable = errors.New("settings getter unavailable")

// failingGetter always fails GetScoped, standing in for a settings-service outage.
type failingGetter struct{}

func (failingGetter) GetScoped(context.Context, settings.Scope, string) (string, error) {
	return "", errGetterUnavailable
}

// TestLimiter_Limit_ReturnsDefaultOnReadFailure is the parity fix: Limit() must return
// the value get() already resolved (the compiled default, on a read failure) alongside
// the error, instead of discarding it.
func TestLimiter_Limit_ReturnsDefaultOnReadFailure(t *testing.T) {
	t.Parallel()

	t.Run("bound", func(t *testing.T) {
		t.Parallel()
		setting := settings.Size(1 * config.GByte)
		setting.Key, setting.Scope = "test.bound", settings.ScopeGlobal
		bl, err := MakeUpperBoundLimiter(Factory{Settings: failingGetter{}}, setting)
		require.NoError(t, err)
		t.Cleanup(func() { assert.NoError(t, bl.Close()) })

		v, err := bl.Limit(t.Context())
		require.ErrorIs(t, err, errGetterUnavailable)
		assert.Equal(t, 1*config.GByte, v)
	})

	t.Run("time", func(t *testing.T) {
		t.Parallel()
		setting := settings.Duration(1 * time.Minute)
		setting.Key, setting.Scope = "test.time", settings.ScopeGlobal
		tl, err := Factory{Settings: failingGetter{}}.MakeTimeLimiter(setting)
		require.NoError(t, err)
		t.Cleanup(func() { assert.NoError(t, tl.Close()) })

		d, err := tl.Limit(t.Context())
		require.ErrorIs(t, err, errGetterUnavailable)
		assert.Equal(t, 1*time.Minute, d)
	})

	t.Run("gate", func(t *testing.T) {
		t.Parallel()
		setting := settings.Bool(true)
		setting.Key, setting.Scope = "test.gate", settings.ScopeGlobal
		gl, err := MakeGateLimiter(Factory{Settings: failingGetter{}}, setting)
		require.NoError(t, err)
		t.Cleanup(func() { assert.NoError(t, gl.Close()) })

		open, err := gl.Limit(t.Context())
		require.ErrorIs(t, err, errGetterUnavailable)
		assert.True(t, open)
	})

	t.Run("range", func(t *testing.T) {
		t.Parallel()
		setting := settings.NewSetting(settings.Range[int]{Lower: 1, Upper: 5}, settings.ParseRangeFn(func(s string) (int, error) {
			return 0, nil
		}))
		setting.Key, setting.Scope = "test.range", settings.ScopeGlobal
		rl, err := MakeRangeLimiter[int](Factory{Settings: failingGetter{}}, setting)
		require.NoError(t, err)
		t.Cleanup(func() { assert.NoError(t, rl.Close()) })

		got, err := rl.Limit(t.Context())
		require.ErrorIs(t, err, errGetterUnavailable)
		assert.Equal(t, settings.Range[int]{Lower: 1, Upper: 5}, got)
	})
}

// TestTimeLimiter_WithTimeout_UsableOnReadFailure is the fix for the actual production
// incident: WithTimeout used to return (nil, nil, err) on a read failure, causing callers
// to drop the unit of work instead of running it with the compiled default timeout.
func TestTimeLimiter_WithTimeout_UsableOnReadFailure(t *testing.T) {
	t.Parallel()

	setting := settings.Duration(10 * time.Second)
	setting.Key, setting.Scope = "test.time.with-timeout", settings.ScopeGlobal
	tl, err := Factory{Settings: failingGetter{}}.MakeTimeLimiter(setting)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, tl.Close()) })

	before := time.Now()
	ctx, done, withTimeoutErr := tl.WithTimeout(t.Context())
	require.Error(t, withTimeoutErr, "err is advisory, not fatal - the read did fail")
	require.NotNil(t, ctx, "ctx must be usable despite the read failure")
	require.NotNil(t, done)
	defer done()

	deadline, ok := ctx.Deadline()
	require.True(t, ok, "ctx must carry a real deadline, not be unbounded")
	assert.InDelta(t, 10*time.Second, deadline.Sub(before), float64(2*time.Second),
		"deadline should be based on the compiled default (10s), not hang open or fire instantly")
}

// TestTimeLimiter_WithTimeout_UsableWithoutTenant covers the missing-tenant path. The timeout
// doesn't depend on the tenant, so the context must still be bounded by the compiled default:
// a zero timeout would expire immediately, and a nil context would make callers drop the work.
// A tenant that was required but missing still surfaces an error, so the bug stays visible.
func TestTimeLimiter_WithTimeout_UsableWithoutTenant(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		scope     settings.Scope
		expectErr bool
	}{
		{settings.ScopeOrg, false},     // tenant not required
		{settings.ScopeWorkflow, true}, // tenant required
	} {
		t.Run(tt.scope.String(), func(t *testing.T) {
			t.Parallel()
			setting := settings.Duration(10 * time.Second)
			setting.Key, setting.Scope = "test.time.no-tenant", tt.scope
			tl, err := Factory{}.MakeTimeLimiter(setting)
			require.NoError(t, err)
			t.Cleanup(func() { assert.NoError(t, tl.Close()) })

			before := time.Now()
			ctx, done, err := tl.WithTimeout(t.Context()) // no contexts.WithCRE, so no tenant
			if tt.expectErr {
				require.Error(t, err, "a required but missing tenant must stay visible")
			} else {
				require.NoError(t, err)
			}
			require.NotNil(t, ctx, "ctx must be usable without a tenant")
			require.NotNil(t, done)
			defer done()

			deadline, ok := ctx.Deadline()
			require.True(t, ok, "ctx must carry a real deadline, not be unbounded")
			assert.InDelta(t, 10*time.Second, deadline.Sub(before), float64(2*time.Second),
				"deadline should be based on the compiled default (10s), not fire instantly")
		})
	}
}
