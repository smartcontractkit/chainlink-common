package orgresolver

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	log "github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// fakeResolver is a test OrgResolver whose responses can be swapped and whose
// Get calls are counted.
type fakeResolver struct {
	mu       sync.Mutex
	values   map[string]string
	err      error
	getCalls int32
}

func newFakeResolver() *fakeResolver {
	return &fakeResolver{values: map[string]string{}}
}

func (f *fakeResolver) set(owner, orgID string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.values[owner] = orgID
}

func (f *fakeResolver) setErr(err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.err = err
}

func (f *fakeResolver) Get(_ context.Context, owner string) (string, error) {
	atomic.AddInt32(&f.getCalls, 1)
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.err != nil {
		return "", f.err
	}
	v, ok := f.values[owner]
	if !ok {
		return "", errors.New("not found")
	}
	return v, nil
}

func (f *fakeResolver) Start(context.Context) error    { return nil }
func (f *fakeResolver) Close() error                   { return nil }
func (f *fakeResolver) Ready() error                   { return nil }
func (f *fakeResolver) HealthReport() map[string]error { return map[string]error{} }
func (f *fakeResolver) Name() string                   { return "fakeResolver" }

func TestCachingOrgResolver_ColdMissThenCache(t *testing.T) {
	inner := newFakeResolver()
	inner.set("owner-1", "org-1")
	c := NewCaching(inner, 0) // no background refresh

	got, err := c.Get(context.Background(), "owner-1")
	require.NoError(t, err)
	assert.Equal(t, "org-1", got)
	assert.EqualValues(t, 1, atomic.LoadInt32(&inner.getCalls))

	// Second Get is served from cache; inner is not called again.
	got, err = c.Get(context.Background(), "owner-1")
	require.NoError(t, err)
	assert.Equal(t, "org-1", got)
	assert.EqualValues(t, 1, atomic.LoadInt32(&inner.getCalls))
}

func TestCachingOrgResolver_ColdMissError(t *testing.T) {
	inner := newFakeResolver()
	c := NewCaching(inner, 0)
	_, err := c.Get(context.Background(), "unknown")
	require.Error(t, err)
}

func TestCachingOrgResolver_BackgroundRefresh(t *testing.T) {
	inner := newFakeResolver()
	inner.set("owner-1", "org-1")
	c := NewCaching(inner, 20*time.Millisecond)
	require.NoError(t, c.Start(context.Background()))
	t.Cleanup(func() { _ = c.Close() })

	// Prime the cache.
	got, err := c.Get(context.Background(), "owner-1")
	require.NoError(t, err)
	assert.Equal(t, "org-1", got)

	// Change the upstream value; background refresh should pick it up.
	inner.set("owner-1", "org-2")
	require.Eventually(t, func() bool {
		v, err := c.Get(context.Background(), "owner-1")
		return err == nil && v == "org-2"
	}, time.Second, 5*time.Millisecond)
}

func TestCachingOrgResolver_RefreshKeepsStaleOnError(t *testing.T) {
	inner := newFakeResolver()
	inner.set("owner-1", "org-1")
	c := NewCaching(inner, 20*time.Millisecond)
	require.NoError(t, c.Start(context.Background()))
	t.Cleanup(func() { _ = c.Close() })

	_, err := c.Get(context.Background(), "owner-1")
	require.NoError(t, err)

	inner.setErr(errors.New("linking service down"))
	// Give the refresh loop time to run and fail a few times.
	time.Sleep(80 * time.Millisecond)

	// Cached value is still served despite refresh failures.
	got, err := c.Get(context.Background(), "owner-1")
	require.NoError(t, err)
	assert.Equal(t, "org-1", got)
}

func TestResolveOrEmpty(t *testing.T) {
	lggr := log.Nop()

	t.Run("nil resolver", func(t *testing.T) {
		assert.Empty(t, ResolveOrEmpty(context.Background(), nil, "owner-1", lggr))
	})

	t.Run("empty owner", func(t *testing.T) {
		inner := newFakeResolver()
		assert.Empty(t, ResolveOrEmpty(context.Background(), inner, "", lggr))
	})

	t.Run("success", func(t *testing.T) {
		inner := newFakeResolver()
		inner.set("owner-1", "org-1")
		assert.Equal(t, "org-1", ResolveOrEmpty(context.Background(), inner, "owner-1", lggr))
	})

	t.Run("error yields empty", func(t *testing.T) {
		inner := newFakeResolver()
		inner.setErr(errors.New("boom"))
		assert.Empty(t, ResolveOrEmpty(context.Background(), inner, "owner-1", lggr))
	})
}
