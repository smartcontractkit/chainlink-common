package limits

import (
	"context"
	"errors"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/config"
	"github.com/smartcontractkit/chainlink-common/pkg/settings"
)

// toggleGetter is a settings.Getter that can be switched between a fixed value and
// failing, to drive the last-known-good fallback path deterministically.
type toggleGetter struct {
	mu      sync.Mutex
	value   string
	failing bool
}

func (g *toggleGetter) succeedWith(value string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = value
	g.failing = false
}

func (g *toggleGetter) fail() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.failing = true
}

var errGetterUnavailable = errors.New("settings getter unavailable")

func (g *toggleGetter) GetScoped(context.Context, settings.Scope, string) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.failing {
		return "", errGetterUnavailable
	}
	return g.value, nil
}

func TestBoundLimiter_LastKnownGoodOnReadFailure(t *testing.T) {
	t.Parallel()

	setting := settings.Size(1 * config.GByte) // compiled default
	setting.Key = "test.bound"
	setting.Scope = settings.ScopeGlobal

	getter := &toggleGetter{}
	bl, err := MakeUpperBoundLimiter(Factory{Settings: getter}, setting)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, bl.Close()) })

	ctx := t.Context()

	getter.succeedWith("20gb")
	v, err := bl.Limit(ctx)
	require.NoError(t, err)
	assert.Equal(t, 20*config.GByte, v)

	getter.fail()
	v, err = bl.Limit(ctx)
	require.ErrorIs(t, err, errGetterUnavailable)
	assert.Equal(t, 20*config.GByte, v, "should fall back to the last resolved value, not the compiled default")
}

func TestBoundLimiter_CompiledDefaultWhenNeverResolved(t *testing.T) {
	t.Parallel()

	setting := settings.Size(1 * config.GByte)
	setting.Key = "test.bound.never-resolved"
	setting.Scope = settings.ScopeGlobal

	getter := &toggleGetter{}
	getter.fail()
	bl, err := MakeUpperBoundLimiter(Factory{Settings: getter}, setting)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, bl.Close()) })

	v, err := bl.Limit(t.Context())
	require.ErrorIs(t, err, errGetterUnavailable)
	assert.Equal(t, 1*config.GByte, v, "with no last-known-good value yet, must fall back to the compiled default")
}

func TestTimeLimiter_LastKnownGoodOnReadFailure(t *testing.T) {
	t.Parallel()

	setting := settings.Duration(1 * time.Minute) // compiled default
	setting.Key = "test.time"
	setting.Scope = settings.ScopeGlobal

	getter := &toggleGetter{}
	tl, err := Factory{Settings: getter}.MakeTimeLimiter(setting)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, tl.Close()) })

	ctx := t.Context()

	getter.succeedWith("10s")
	d, err := tl.Limit(ctx)
	require.NoError(t, err)
	assert.Equal(t, 10*time.Second, d)

	getter.fail()
	d, err = tl.Limit(ctx)
	require.ErrorIs(t, err, errGetterUnavailable)
	assert.Equal(t, 10*time.Second, d, "should fall back to the last resolved value, not the compiled default")
}

// TestTimeLimiter_WithTimeout_UsableOnReadFailure: WithTimeout used to return
// (nil, nil, err) on a read failure, causing callers to drop the unit of work instead
// of running it with a real timeout.
func TestTimeLimiter_WithTimeout_UsableOnReadFailure(t *testing.T) {
	t.Parallel()

	setting := settings.Duration(1 * time.Minute)
	setting.Key = "test.time.with-timeout"
	setting.Scope = settings.ScopeGlobal

	getter := &toggleGetter{}
	getter.succeedWith("10s")
	tl, err := Factory{Settings: getter}.MakeTimeLimiter(setting)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, tl.Close()) })

	// Warm the last-known-good value via a successful call, then fail the getter.
	_, err = tl.Limit(t.Context())
	require.NoError(t, err)
	getter.fail()

	before := time.Now()
	ctx, done, withTimeoutErr := tl.WithTimeout(t.Context())
	require.Error(t, withTimeoutErr, "err is advisory, not fatal - the read did fail")
	require.NotNil(t, ctx, "ctx must be usable despite the read failure")
	require.NotNil(t, done)
	defer done()

	deadline, ok := ctx.Deadline()
	require.True(t, ok, "ctx must carry a real deadline, not be unbounded")
	assert.InDelta(t, 10*time.Second, deadline.Sub(before), float64(2*time.Second),
		"deadline should be based on the last known good timeout (10s), not hang open or fire instantly")
}

func TestGateLimiter_LastKnownGoodOnReadFailure(t *testing.T) {
	t.Parallel()

	setting := settings.Bool(false) // compiled default
	setting.Key = "test.gate"
	setting.Scope = settings.ScopeGlobal

	getter := &toggleGetter{}
	gl, err := MakeGateLimiter(Factory{Settings: getter}, setting)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, gl.Close()) })

	ctx := t.Context()

	getter.succeedWith("true")
	open, err := gl.Limit(ctx)
	require.NoError(t, err)
	assert.True(t, open)

	getter.fail()
	open, err = gl.Limit(ctx)
	require.ErrorIs(t, err, errGetterUnavailable)
	assert.True(t, open, "should fall back to the last resolved value, not the compiled default")
}

func TestRangeLimiter_LastKnownGoodOnReadFailure(t *testing.T) {
	t.Parallel()

	setting := settings.NewSetting(settings.Range[int]{Lower: 0, Upper: 5}, settings.ParseRangeFn(strconv.Atoi))
	setting.Key = "test.range"
	setting.Scope = settings.ScopeGlobal

	getter := &toggleGetter{}
	rl, err := MakeRangeLimiter[int](Factory{Settings: getter}, setting)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, rl.Close()) })

	ctx := t.Context()

	getter.succeedWith("[1,50]")
	got, err := rl.Limit(ctx)
	require.NoError(t, err)
	assert.Equal(t, settings.Range[int]{Lower: 1, Upper: 50}, got)

	getter.fail()
	got, err = rl.Limit(ctx)
	require.ErrorIs(t, err, errGetterUnavailable)
	assert.Equal(t, settings.Range[int]{Lower: 1, Upper: 50}, got, "should fall back to the last resolved value, not the compiled default")
}
