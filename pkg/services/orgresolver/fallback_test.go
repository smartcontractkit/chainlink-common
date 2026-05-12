package orgresolver

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// mockInnerResolver is a controllable OrgResolver for testing the fallback layer.
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

// -- Happy path tests --

func TestOrgResolverFallback_Success_PopulatesCache(t *testing.T) {
	inner := newMockInner(func(_ context.Context, owner string) (string, error) {
		return "org-123", nil
	})
	c := NewOrgResolverWithFallback(inner, logger.Test(t))

	orgID, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-123", orgID)
	assert.Equal(t, int32(1), inner.calls.Load())

	// Simulate inner now failing; cache should serve the previously stored value.
	inner.setGetFunc(func(_ context.Context, _ string) (string, error) {
		return "", status.Error(codes.NotFound, "gone")
	})

	orgID, err = c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-123", orgID)
	assert.Equal(t, int32(2), inner.calls.Load())
}

func TestOrgResolverFallback_Success_UpdatesCache(t *testing.T) {
	callCount := atomic.Int32{}
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		n := callCount.Add(1)
		if n == 1 {
			return "org-old", nil
		}
		return "org-new", nil
	})
	c := NewOrgResolverWithFallback(inner, logger.Test(t))

	orgID, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-old", orgID)

	orgID, err = c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-new", orgID)
}

// -- Unavailable error tests --

func TestOrgResolverFallback_Unavailable_RetrySucceeds(t *testing.T) {
	callCount := atomic.Int32{}
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		if callCount.Add(1) == 1 {
			return "", status.Error(codes.Unavailable, "service down")
		}
		return "org-retry-ok", nil
	})
	c := NewOrgResolverWithFallback(inner, logger.Test(t))

	orgID, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-retry-ok", orgID)
	assert.Equal(t, int32(2), inner.calls.Load(), "inner should be called exactly twice (initial + retry)")
}

func TestOrgResolverFallback_Unavailable_RetryFails_CacheHit(t *testing.T) {
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "org-first", nil
	})
	c := NewOrgResolverWithFallback(inner, logger.Test(t))

	// Populate cache.
	_, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, int32(1), inner.calls.Load())

	// Now both initial and retry return Unavailable.
	inner.setGetFunc(func(_ context.Context, _ string) (string, error) {
		return "", status.Error(codes.Unavailable, "still down")
	})

	orgID, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-first", orgID, "should return cached value")
	assert.Equal(t, int32(3), inner.calls.Load(), "initial + retry = 2 more calls")
}

func TestOrgResolverFallback_Unavailable_RetryFails_CacheMiss(t *testing.T) {
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "", status.Error(codes.Unavailable, "down")
	})
	c := NewOrgResolverWithFallback(inner, logger.Test(t))

	_, err := c.Get(context.Background(), "owner-never-seen")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Unavailable")
	assert.Equal(t, int32(2), inner.calls.Load(), "initial + retry")
}

// -- NotFound error tests --

func TestOrgResolverFallback_NotFound_CacheHit(t *testing.T) {
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "org-known", nil
	})
	c := NewOrgResolverWithFallback(inner, logger.Test(t))

	// Populate cache.
	_, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)

	// Now inner returns NotFound (race condition: owner unlinked).
	inner.setGetFunc(func(_ context.Context, _ string) (string, error) {
		return "", status.Error(codes.NotFound, "not found")
	})

	orgID, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-known", orgID)
	assert.Equal(t, int32(2), inner.calls.Load(), "no retry for NotFound")
}

func TestOrgResolverFallback_NotFound_CacheMiss(t *testing.T) {
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "", status.Error(codes.NotFound, "not found")
	})
	c := NewOrgResolverWithFallback(inner, logger.Test(t))

	_, err := c.Get(context.Background(), "owner-unknown")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "NotFound")
	assert.Equal(t, int32(1), inner.calls.Load(), "no retry for NotFound")
}

// -- Error passthrough tests --

func TestOrgResolverFallback_OtherError_NoFallback(t *testing.T) {
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "org-good", nil
	})
	c := NewOrgResolverWithFallback(inner, logger.Test(t))

	// Populate cache.
	_, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)

	// Non-gRPC error (e.g., JWT failure).
	jwtErr := errors.New("JWT generation failed")
	inner.setGetFunc(func(_ context.Context, _ string) (string, error) {
		return "", jwtErr
	})

	_, err = c.Get(context.Background(), "owner-a")
	require.Error(t, err)
	assert.Equal(t, jwtErr, err, "non-gRPC errors should pass through without cache fallback")
	assert.Equal(t, int32(2), inner.calls.Load())
}

func TestOrgResolverFallback_WrappedGRPCError(t *testing.T) {
	// Simulate the wrapping done in linking.go: fmt.Errorf("failed to fetch ...: %w", grpcErr)
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "org-good", nil
	})
	c := NewOrgResolverWithFallback(inner, logger.Test(t))

	// Populate cache.
	_, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)

	inner.setGetFunc(func(_ context.Context, _ string) (string, error) {
		return "", fmt.Errorf("failed to fetch organization from workflow owner: %w",
			status.Error(codes.NotFound, "owner not linked"))
	})

	orgID, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-good", orgID, "should detect NotFound through fmt.Errorf wrapping")
}

func TestOrgResolverFallback_WrappedUnavailableError(t *testing.T) {
	inner := newMockInner(func(_ context.Context, _ string) (string, error) {
		return "org-cached", nil
	})
	c := NewOrgResolverWithFallback(inner, logger.Test(t))

	// Populate cache.
	_, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)

	inner.setGetFunc(func(_ context.Context, _ string) (string, error) {
		return "", fmt.Errorf("failed to fetch organization from workflow owner: %w",
			status.Error(codes.Unavailable, "connection refused"))
	})

	orgID, err := c.Get(context.Background(), "owner-a")
	require.NoError(t, err)
	assert.Equal(t, "org-cached", orgID, "should detect Unavailable through wrapping and fall back to cache after retry")
}

// -- Service interface delegation tests --

func TestOrgResolverFallback_DelegatesServiceMethods(t *testing.T) {
	inner := &mockInnerResolver{
		getFunc:  func(_ context.Context, _ string) (string, error) { return "", nil },
		startErr: errors.New("start-err"),
		closeErr: errors.New("close-err"),
		readyErr: errors.New("ready-err"),
		name:     "test-resolver",
	}
	c := NewOrgResolverWithFallback(inner, logger.Test(t))

	assert.Equal(t, inner.startErr, c.Start(context.Background()))
	assert.Equal(t, inner.closeErr, c.Close())
	assert.Equal(t, inner.readyErr, c.Ready())
	assert.Equal(t, "test-resolver", c.Name())
	hr := c.HealthReport()
	assert.Contains(t, hr, "test-resolver")
}

// -- Concurrency test --

func TestOrgResolverFallback_ConcurrentAccess(t *testing.T) {
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
	c := NewOrgResolverWithFallback(inner, logger.Test(t))

	var wg sync.WaitGroup
	owners := []string{"owner-1", "owner-2", "owner-3", "owner-4", "owner-5"}
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			owner := owners[idx%len(owners)]
			orgID, err := c.Get(context.Background(), owner)
			// Either succeeds or returns an error; must not panic.
			if err == nil {
				assert.NotEmpty(t, orgID)
			}
		}(i)
	}
	wg.Wait()
}
