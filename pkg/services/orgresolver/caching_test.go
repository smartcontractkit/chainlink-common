package orgresolver

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// mockInnerResolver is a controllable OrgResolver for testing the caching layer.
type mockInnerResolver struct {
	calls    atomic.Int32
	mu       sync.Mutex
	getFunc  func(ctx context.Context, owner string) (string, error)
	startErr error
	closeErr error
	readyErr error
	name     string
}

func (m *mockInnerResolver) Get(ctx context.Context, owner string) (string, error) {
	m.calls.Add(1)
	m.mu.Lock()
	fn := m.getFunc
	m.mu.Unlock()
	return fn(ctx, owner)
}

func (m *mockInnerResolver) setGetFunc(fn func(ctx context.Context, owner string) (string, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getFunc = fn
}

func (m *mockInnerResolver) Start(_ context.Context) error  { return m.startErr }
func (m *mockInnerResolver) Close() error                   { return m.closeErr }
func (m *mockInnerResolver) Ready() error                   { return m.readyErr }
func (m *mockInnerResolver) Name() string                   { return m.name }
func (m *mockInnerResolver) HealthReport() map[string]error { return map[string]error{m.name: nil} }

func newMockInner(fn func(ctx context.Context, owner string) (string, error)) *mockInnerResolver {
	return &mockInnerResolver{getFunc: fn, name: "mockInner"}
}

// erroringCache wraps a Cache and can be made to fail Get/Set on demand, to
// exercise CachingResolver's handling of a misbehaving durable store.
type erroringCache struct {
	Cache
	getErr error
	setErr error
}

func (c *erroringCache) Get(ctx context.Context, owner string) (CacheEntry, bool, error) {
	if c.getErr != nil {
		return CacheEntry{}, false, c.getErr
	}
	return c.Cache.Get(ctx, owner)
}

func (c *erroringCache) Set(ctx context.Context, owner string, entry CacheEntry) error {
	if c.setErr != nil {
		return c.setErr
	}
	return c.Cache.Set(ctx, owner, entry)
}

func newTestCachingResolver(t *testing.T, inner OrgResolver, cfg CachingResolverConfig) *CachingResolver {
	t.Helper()
	if cfg.Cache == nil {
		cfg.Cache = NewInMemoryCache()
	}
	c, err := NewCachingResolver(inner, cfg, logger.Test(t))
	require.NoError(t, err)
	return c
}

// newTestCachingResolverStarted additionally starts the background refresh
// loop and registers its shutdown, for tests that exercise the ticker end to
// end.
func newTestCachingResolverStarted(t *testing.T, inner OrgResolver, cfg CachingResolverConfig) *CachingResolver {
	t.Helper()
	c := newTestCachingResolver(t, inner, cfg)
	require.NoError(t, c.Start(context.Background()))
	t.Cleanup(func() { require.NoError(t, c.Close()) })
	return c
}

const eventuallyTimeout = 2 * time.Second
const eventuallyTick = 5 * time.Millisecond

// -- Fast-path tests: any cache hit must return immediately without blocking --

func TestCachingResolver_FreshCacheHit_SkipsInnerResolver(t *testing.T) {
	inner := newMockInner(func(_ context.Context, owner string) (string, error) {
		return "org-for-" + owner, nil
	})
	cache := NewInMemoryCache()
	c := newTestCachingResolver(t, inner, CachingResolverConfig{Cache: cache, RefreshInterval: time.Hour})

	// First call: cache miss, blocks on the inner resolver and populates the cache.
	orgID, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-for-owner-a", orgID)
	assert.Equal(t, int32(1), inner.calls.Load())

	// Second call: served entirely from cache, no network call.
	orgID, err = c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-for-owner-a", orgID)
	assert.Equal(t, int32(1), inner.calls.Load(), "cache hit must not call inner resolver")
}

func TestCachingResolver_CacheHit_NeverBlocksOnBackgroundRefresh(t *testing.T) {
	// Even while the background loop is refreshing an owner, a cache hit for
	// that owner must be returned synchronously.
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "org-first", nil
	})
	cache := NewInMemoryCache()
	require.NoError(t, cache.Set(context.Background(), "owner-a", CacheEntry{OrgID: "org-first", RefreshedAt: time.Now()}))
	c := newTestCachingResolverStarted(t, inner, CachingResolverConfig{Cache: cache, RefreshInterval: time.Millisecond})

	release := make(chan struct{})
	inner.setGetFunc(func(_ context.Context, _ string) (string, error) {
		<-release
		return "org-first", nil
	})
	t.Cleanup(func() { close(release) }) // let the background refresh finish so it doesn't leak past the test

	start := time.Now()
	orgID, err := c.Get(context.Background(), "owner-a")
	elapsed := time.Since(start)
	require.NoError(t, err)
	assert.Equal(t, "org-first", orgID)
	assert.Less(t, elapsed, 100*time.Millisecond, "cache hit must not block on the background refresh loop")
}

func TestCachingResolver_CacheHit_IgnoresCancelledContext(t *testing.T) {
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "org-cached", nil
	})
	cache := NewInMemoryCache()
	c := newTestCachingResolver(t, inner, CachingResolverConfig{Cache: cache, RefreshInterval: time.Hour})

	_, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	orgID, err := c.Get(ctx, "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-cached", orgID)
}

func TestCachingResolver_DefaultRefreshInterval(t *testing.T) {
	inner := newMockInner(func(_ context.Context, owner string) (string, error) {
		return "org-for-" + owner, nil
	})
	c := newTestCachingResolver(t, inner, CachingResolverConfig{})
	assert.Equal(t, defaultRefreshInterval, c.refreshInterval)
}

func TestCachingResolver_RequiresCache(t *testing.T) {
	inner := newMockInner(func(_ context.Context, owner string) (string, error) {
		return "org", nil
	})
	_, err := NewCachingResolver(inner, CachingResolverConfig{}, logger.Test(t))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cache is required")
}

// -- Owner-tracking tests --

func TestCachingResolver_Get_RemembersOwnerForBackgroundRefresh(t *testing.T) {
	inner := newMockInner(func(_ context.Context, owner string) (string, error) {
		return "org-for-" + owner, nil
	})
	c := newTestCachingResolver(t, inner, CachingResolverConfig{})

	_, err := c.Get(context.Background(), "owner-a") // cache miss
	require.NoError(t, err)
	_, err = c.Get(context.Background(), "owner-a") // cache hit
	require.NoError(t, err)

	c.ownersMu.Lock()
	_, remembered := c.owners["owner-a"]
	c.ownersMu.Unlock()
	assert.True(t, remembered, "Get must remember the owner regardless of cache hit/miss")
}

// -- refreshOwner tests: exercise the per-owner refresh logic directly and
// deterministically, without depending on the background loop's timing --

func TestCachingResolver_RefreshOwner_UpdatesCache(t *testing.T) {
	inner := newMockInner(func(_ context.Context, owner string) (string, error) {
		return "org-refreshed", nil
	})
	cache := NewInMemoryCache()
	require.NoError(t, cache.Set(context.Background(), "owner-a", CacheEntry{OrgID: "org-old", RefreshedAt: time.Now()}))
	c := newTestCachingResolver(t, inner, CachingResolverConfig{Cache: cache})

	c.refreshOwner(context.Background(), "owner-a")

	entry, ok, err := cache.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "org-refreshed", entry.OrgID)
}

func TestCachingResolver_RefreshOwner_SkipsOwnerWithNoCacheEntry(t *testing.T) {
	inner := newMockInner(func(_ context.Context, owner string) (string, error) {
		return "org-for-" + owner, nil
	})
	c := newTestCachingResolver(t, inner, CachingResolverConfig{})

	c.refreshOwner(context.Background(), "owner-never-cached")

	assert.Equal(t, int32(0), inner.calls.Load(), "an owner with no cache entry must not be refreshed")
}

func TestCachingResolver_RefreshOwner_RetriesRetriableErrors(t *testing.T) {
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "", status.Error(codes.Unavailable, "still down")
	})
	cache := NewInMemoryCache()
	require.NoError(t, cache.Set(context.Background(), "owner-a", CacheEntry{OrgID: "org-first", RefreshedAt: time.Now()}))
	c := newTestCachingResolver(t, inner, CachingResolverConfig{Cache: cache})

	c.refreshOwner(context.Background(), "owner-a")

	assert.Equal(t, int32(4), inner.calls.Load(), "initial attempt + 3 retries")
	entry, ok, err := cache.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "org-first", entry.OrgID, "a failed refresh must leave the existing entry in place")
}

func TestCachingResolver_RefreshOwner_DoesNotRetryNotFound(t *testing.T) {
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "", status.Error(codes.NotFound, "not found")
	})
	cache := NewInMemoryCache()
	require.NoError(t, cache.Set(context.Background(), "owner-a", CacheEntry{OrgID: "org-known", RefreshedAt: time.Now()}))
	c := newTestCachingResolver(t, inner, CachingResolverConfig{Cache: cache})

	c.refreshOwner(context.Background(), "owner-a")

	assert.Equal(t, int32(1), inner.calls.Load(), "no retry for NotFound")
	entry, ok, err := cache.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "org-known", entry.OrgID)
}

func TestCachingResolver_RefreshOwner_NonGRPCError_KeepsCachedValue(t *testing.T) {
	jwtErr := errors.New("JWT generation failed")
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "", jwtErr
	})
	cache := NewInMemoryCache()
	require.NoError(t, cache.Set(context.Background(), "owner-a", CacheEntry{OrgID: "org-good", RefreshedAt: time.Now()}))
	c := newTestCachingResolver(t, inner, CachingResolverConfig{Cache: cache})

	c.refreshOwner(context.Background(), "owner-a")

	assert.Equal(t, int32(1), inner.calls.Load(), "non-retriable error should not be retried")
	entry, ok, err := cache.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, "org-good", entry.OrgID)
}

func TestCachingResolver_RefreshOwner_DetectsWrappedGRPCErrors(t *testing.T) {
	// Simulate the wrapping done in linking.go: fmt.Errorf("failed to fetch ...: %w", grpcErr)
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "", fmt.Errorf("failed to fetch organization from workflow owner: %w",
			status.Error(codes.NotFound, "owner not linked"))
	})
	cache := NewInMemoryCache()
	require.NoError(t, cache.Set(context.Background(), "owner-a", CacheEntry{OrgID: "org-good", RefreshedAt: time.Now()}))
	c := newTestCachingResolver(t, inner, CachingResolverConfig{Cache: cache})

	c.refreshOwner(context.Background(), "owner-a")

	assert.Equal(t, int32(1), inner.calls.Load(), "wrapped NotFound should be detected and not retried")
}

func TestCachingResolver_RefreshOwner_RecordsMappingChangedMetric(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })

	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "org-new", nil
	})
	cache := NewInMemoryCache()
	require.NoError(t, cache.Set(context.Background(), "owner-a", CacheEntry{OrgID: "org-old", RefreshedAt: time.Now()}))
	c := newTestCachingResolver(t, inner, CachingResolverConfig{Cache: cache, Meter: provider.Meter("test")})

	c.refreshOwner(context.Background(), "owner-a")

	assert.Equal(t, map[string]int64{"owner-a": 1}, counterCountsByAttr(t, reader, "org_resolver_mapping_changed", "owner"))
}

func TestCachingResolver_RefreshOwner_NoMeter_MappingChangeDoesNotPanic(t *testing.T) {
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "org-new", nil
	})
	cache := NewInMemoryCache()
	require.NoError(t, cache.Set(context.Background(), "owner-a", CacheEntry{OrgID: "org-old", RefreshedAt: time.Now()}))
	c := newTestCachingResolver(t, inner, CachingResolverConfig{Cache: cache})

	assert.NotPanics(t, func() {
		c.refreshOwner(context.Background(), "owner-a")
	})
}

// -- refreshAllOwners tests --

func TestCachingResolver_RefreshAllOwners_RefreshesEveryRememberedOwner(t *testing.T) {
	inner := newMockInner(func(_ context.Context, owner string) (string, error) {
		return "org-refreshed-" + owner, nil
	})
	cache := NewInMemoryCache()
	require.NoError(t, cache.Set(context.Background(), "owner-a", CacheEntry{OrgID: "org-old-a", RefreshedAt: time.Now()}))
	require.NoError(t, cache.Set(context.Background(), "owner-b", CacheEntry{OrgID: "org-old-b", RefreshedAt: time.Now()}))
	c := newTestCachingResolver(t, inner, CachingResolverConfig{Cache: cache})
	c.rememberOwner("owner-a")
	c.rememberOwner("owner-b")

	c.refreshAllOwners(context.Background())

	for _, owner := range []string{"owner-a", "owner-b"} {
		entry, ok, err := cache.Get(context.Background(), owner)
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, "org-refreshed-"+owner, entry.OrgID)
	}
}

// -- Background loop tests: exercise Start's ticker end to end --

func TestCachingResolver_BackgroundLoop_PeriodicallyRefreshesKnownOwners(t *testing.T) {
	callCount := atomic.Int32{}
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		n := callCount.Add(1)
		if n == 1 {
			return "org-old", nil
		}
		return "org-new", nil
	})
	cache := NewInMemoryCache()
	require.NoError(t, cache.Set(context.Background(), "owner-a", CacheEntry{OrgID: "org-old", RefreshedAt: time.Now()}))
	c := newTestCachingResolverStarted(t, inner, CachingResolverConfig{Cache: cache, RefreshInterval: 5 * time.Millisecond})

	// The stale value is served immediately, from cache; the loop hasn't
	// necessarily ticked yet.
	orgID, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-old", orgID)

	require.Eventually(t, func() bool {
		entry, ok, err := cache.Get(context.Background(), "owner-a")
		return err == nil && ok && entry.OrgID == "org-new"
	}, eventuallyTimeout, eventuallyTick, "the background loop must eventually refresh a known owner")
}

func TestCachingResolver_BackgroundLoop_DoesNotRefreshBeforeIntervalElapses(t *testing.T) {
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "org-refreshed", nil
	})
	cache := NewInMemoryCache()
	require.NoError(t, cache.Set(context.Background(), "owner-a", CacheEntry{OrgID: "org-old", RefreshedAt: time.Now()}))
	c := newTestCachingResolverStarted(t, inner, CachingResolverConfig{Cache: cache, RefreshInterval: time.Hour})

	_, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)

	time.Sleep(20 * time.Millisecond)
	assert.Equal(t, int32(0), inner.calls.Load(), "the loop must not refresh before RefreshInterval elapses")
}

// -- Cache-miss tests: only path where Get blocks on the underlying resolver --

func TestCachingResolver_CacheMiss_PopulatesCache(t *testing.T) {
	inner := newMockInner(func(_ context.Context, owner string) (string, error) {
		return "org-123", nil
	})
	c := newTestCachingResolver(t, inner, CachingResolverConfig{})

	orgID, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-123", orgID)
	assert.Equal(t, int32(1), inner.calls.Load())
}

func TestCachingResolver_CacheMiss_Unavailable_RetrySucceeds(t *testing.T) {
	callCount := atomic.Int32{}
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		if callCount.Add(1) == 1 {
			return "", status.Error(codes.Unavailable, "service down")
		}
		return "org-retry-ok", nil
	})
	c := newTestCachingResolver(t, inner, CachingResolverConfig{})

	orgID, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-retry-ok", orgID)
	assert.Equal(t, int32(2), inner.calls.Load(), "inner should be called exactly twice (initial + retry)")
}

func TestCachingResolver_CacheMiss_Unavailable_RetriesExhausted(t *testing.T) {
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "", status.Error(codes.Unavailable, "down")
	})
	c := newTestCachingResolver(t, inner, CachingResolverConfig{})

	_, err := c.Get(context.Background(), "owner-never-seen")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Unavailable")
	assert.Equal(t, int32(4), inner.calls.Load(), "initial + 3 retries")
}

func TestCachingResolver_CacheMiss_DeadlineExceeded_RetrySucceeds(t *testing.T) {
	callCount := atomic.Int32{}
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		if callCount.Add(1) < 3 {
			return "", status.Error(codes.DeadlineExceeded, "timeout")
		}
		return "org-after-retries", nil
	})
	c := newTestCachingResolver(t, inner, CachingResolverConfig{})

	orgID, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-after-retries", orgID)
	assert.Equal(t, int32(3), inner.calls.Load())
}

func TestCachingResolver_CacheMiss_ContextCancelledDuringBackoff(t *testing.T) {
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "", status.Error(codes.Unavailable, "down")
	})
	c := newTestCachingResolver(t, inner, CachingResolverConfig{})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.Get(ctx, "owner-never-seen")
	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.NotContains(t, err.Error(), "Unavailable")
}

func TestCachingResolver_CacheMiss_NotFound(t *testing.T) {
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "", status.Error(codes.NotFound, "not found")
	})
	c := newTestCachingResolver(t, inner, CachingResolverConfig{})

	_, err := c.Get(context.Background(), "owner-unknown")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NotFound")
	assert.Equal(t, int32(1), inner.calls.Load(), "no retry for NotFound")
}

func TestCachingResolver_CacheMiss_NonGRPCError_Propagates(t *testing.T) {
	jwtErr := errors.New("JWT generation failed")
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "", jwtErr
	})
	c := newTestCachingResolver(t, inner, CachingResolverConfig{})

	_, err := c.Get(context.Background(), "owner-a")
	require.Error(t, err)
	assert.Equal(t, jwtErr, err)
	assert.Equal(t, int32(1), inner.calls.Load(), "non-retriable error should not be retried")
}

func TestCachingResolver_CacheMiss_WrappedUnavailableError_Retries(t *testing.T) {
	callCount := atomic.Int32{}
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		if callCount.Add(1) == 1 {
			return "", fmt.Errorf("failed to fetch organization from workflow owner: %w",
				status.Error(codes.Unavailable, "connection refused"))
		}
		return "org-good", nil
	})
	c := newTestCachingResolver(t, inner, CachingResolverConfig{})

	orgID, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-good", orgID)
	assert.Equal(t, int32(2), inner.calls.Load())
}

// -- Cache store failure tests --

func TestCachingResolver_CacheReadError_TreatedAsMissFallsBackToInner(t *testing.T) {
	inner := newMockInner(func(_ context.Context, owner string) (string, error) {
		return "org-for-" + owner, nil
	})
	cache := &erroringCache{Cache: NewInMemoryCache(), getErr: errors.New("cache unavailable")}
	c := newTestCachingResolver(t, inner, CachingResolverConfig{Cache: cache})

	orgID, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-for-owner-a", orgID)
	assert.Equal(t, int32(1), inner.calls.Load())
}

func TestCachingResolver_CacheWriteError_DoesNotFailGet(t *testing.T) {
	inner := newMockInner(func(_ context.Context, owner string) (string, error) {
		return "org-for-" + owner, nil
	})
	cache := &erroringCache{Cache: NewInMemoryCache(), setErr: errors.New("cache write failed")}
	c := newTestCachingResolver(t, inner, CachingResolverConfig{Cache: cache})

	orgID, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-for-owner-a", orgID)
}

// -- Metrics tests --

func TestCachingResolver_RecordsCacheLookupMetrics(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })

	inner := newMockInner(func(_ context.Context, owner string) (string, error) {
		return "org-for-" + owner, nil
	})
	cache := &erroringCache{Cache: NewInMemoryCache()}
	c := newTestCachingResolver(t, inner, CachingResolverConfig{Cache: cache, Meter: provider.Meter("test")})

	_, err := c.Get(context.Background(), "owner-a") // miss, populates the cache
	require.NoError(t, err)

	_, err = c.Get(context.Background(), "owner-a") // hit
	require.NoError(t, err)

	cache.getErr = errors.New("cache down")
	_, err = c.Get(context.Background(), "owner-a") // error, falls back to inner resolver
	require.NoError(t, err)

	assert.Equal(t, map[string]int64{
		cacheResultMiss:  1,
		cacheResultHit:   1,
		cacheResultError: 1,
	}, cacheLookupCounts(t, reader))
}

func TestCachingResolver_NoMeter_DoesNotPanic(t *testing.T) {
	inner := newMockInner(func(_ context.Context, owner string) (string, error) {
		return "org-for-" + owner, nil
	})
	c := newTestCachingResolver(t, inner, CachingResolverConfig{})

	assert.NotPanics(t, func() {
		_, err := c.Get(context.Background(), "owner-a")
		require.NoError(t, err)
		_, err = c.Get(context.Background(), "owner-a")
		require.NoError(t, err)
	})
}

func cacheLookupCounts(t *testing.T, reader *sdkmetric.ManualReader) map[string]int64 {
	t.Helper()
	return counterCountsByAttr(t, reader, "org_resolver_cache_lookups", "result")
}

// counterCountsByAttr sums the named Int64Counter's data points, grouped by
// the value of attrKey (or ungrouped, under the "" key, if attrKey is empty).
func counterCountsByAttr(t *testing.T, reader *sdkmetric.ManualReader, name, attrKey string) map[string]int64 {
	t.Helper()
	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))

	counts := map[string]int64{}
	for _, sm := range rm.ScopeMetrics {
		for _, m := range sm.Metrics {
			if m.Name != name {
				continue
			}
			sum, ok := m.Data.(metricdata.Sum[int64])
			require.True(t, ok)
			for _, dp := range sum.DataPoints {
				key := ""
				if attrKey != "" {
					val, ok := dp.Attributes.Value(attribute.Key(attrKey))
					require.True(t, ok, "missing attribute %q", attrKey)
					key = val.AsString()
				}
				counts[key] += dp.Value
			}
		}
	}
	return counts
}

// -- Service interface delegation tests --

func TestCachingResolver_DelegatesServiceMethods(t *testing.T) {
	inner := &mockInnerResolver{
		getFunc:  func(_ context.Context, _ string) (string, error) { return "", nil },
		startErr: errors.New("start-err"),
		closeErr: errors.New("close-err"),
		readyErr: errors.New("ready-err"),
		name:     "test-resolver",
	}
	c := newTestCachingResolver(t, inner, CachingResolverConfig{})

	assert.Equal(t, inner.startErr, c.Start(context.Background()))
	assert.Equal(t, inner.closeErr, c.Close())
	assert.Equal(t, inner.readyErr, c.Ready())
	assert.Equal(t, "test-resolver", c.Name())
	hr := c.HealthReport()
	assert.Contains(t, hr, "test-resolver")
}

// -- Concurrency test --

func TestCachingResolver_ConcurrentAccess(t *testing.T) {
	callCount := atomic.Int32{}
	inner := newMockInner(func(_ context.Context, owner string) (string, error) {
		n := callCount.Add(1)
		if n%3 == 0 {
			return "", status.Error(codes.Unavailable, "intermittent")
		}
		if n%5 == 0 {
			return "", status.Error(codes.NotFound, "not found")
		}
		return "org-for-" + owner, nil
	})
	c := newTestCachingResolverStarted(t, inner, CachingResolverConfig{RefreshInterval: time.Millisecond})

	var wg sync.WaitGroup
	owners := []string{"owner-1", "owner-2", "owner-3", "owner-4", "owner-5"}
	for i := range 50 {
		wg.Go(func() {
			owner := owners[i%len(owners)]
			orgID, err := c.Get(context.Background(), owner)
			// Either succeeds or returns an error; must not panic.
			if err == nil {
				assert.NotEmpty(t, orgID)
			}
		})
	}
	wg.Wait()
}
