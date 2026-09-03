package orgresolver

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/jpillora/backoff"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	log "github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/retry"
)

const (
	maxGetRetries       = 3
	initialRetryBackoff = 100 * time.Millisecond
	maxRetryBackoff     = 1 * time.Second

	// defaultRefreshInterval is how often the background loop revalidates
	// every known owner against the underlying resolver
	defaultRefreshInterval = 10 * time.Minute
	// refreshOwnerDelay is the pause between successive owner revalidations
	// within a single refreshAllOwners sweep
	refreshOwnerDelay = 100 * time.Millisecond

	cacheResultHit   = "hit"
	cacheResultMiss  = "miss"
	cacheResultError = "error"
)

// CacheEntry is a cached owner->orgID mapping, along with the time it was last
// confirmed against the underlying resolver.
type CacheEntry struct {
	OrgID       string
	RefreshedAt time.Time
}

// Cache abstracts the storage backing CachingResolver, so the application
// wiring it up can choose an in-memory cache (see NewInMemoryCache) or a
// durable store, such as a DB table, depending on its deployment needs.
type Cache interface {
	// Get returns the cached entry for owner. ok is false if no entry exists.
	Get(ctx context.Context, owner string) (entry CacheEntry, ok bool, err error)
	// Set stores or updates the mapping for owner.
	Set(ctx context.Context, owner string, entry CacheEntry) error
}

type InMemoryCache struct {
	mu      sync.Mutex
	entries map[string]CacheEntry // owner -> CacheEntry
}

func NewInMemoryCache() *InMemoryCache {
	return &InMemoryCache{entries: make(map[string]CacheEntry)}
}

func (m *InMemoryCache) Get(_ context.Context, owner string) (CacheEntry, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry, ok := m.entries[owner]
	return entry, ok, nil
}

func (m *InMemoryCache) Set(_ context.Context, owner string, entry CacheEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.entries[owner] = entry
	return nil
}

type CachingResolverConfig struct {
	// Cache backs the owner->orgID mapping. Required; the caller picks the
	// implementation (e.g. NewInMemoryCache, or a durable DB-backed store).
	Cache Cache
	// RefreshInterval controls how often the background loop revalidates
	// every known owner against the underlying resolver. It does not affect
	// whether a cache hit is trusted - a present entry is always returned
	// immediately, refreshed or not. Defaults to DefaultRefreshInterval when
	// zero.
	RefreshInterval time.Duration
	// Meter, if provided, is used to record metrics.
	Meter metric.Meter // optional
}

// CachingResolver wraps an OrgResolver with a Cache of owner->orgID mappings.
//
// An owner's org mapping is not expected to change, so once an entry is
// cached it is trusted indefinitely and returned immediately, without ever
// blocking on a call to the underlying resolver. To guard against a bad
// entry becoming permanent (e.g. from a resolver bug or a transient
// corruption), a single background loop - started by Start and stopped by
// Close - wakes up every RefreshInterval and revalidates every owner ever
// passed to Get against the underlying resolver; this never delays the Get
// call that triggered it. If a revalidation fails, the existing cache entry
// is left in place and continues to be served.
type CachingResolver struct {
	inner           OrgResolver
	cache           Cache
	refreshInterval time.Duration
	logger          log.SugaredLogger

	ownersMu sync.Mutex
	owners   map[string]struct{} // every owner ever passed to Get

	wg     sync.WaitGroup
	cancel context.CancelFunc

	cacheLookups   metric.Int64Counter // tagged with result=hit|miss|error; nil if no Meter was configured
	mappingChanges metric.Int64Counter // incremented on a detected owner->orgID mapping change; nil if no Meter was configured
}

func NewCachingResolver(inner OrgResolver, cfg CachingResolverConfig, logger log.Logger) (*CachingResolver, error) {
	if cfg.Cache == nil {
		return nil, errors.New("Cache is required")
	}
	refreshInterval := cfg.RefreshInterval
	if refreshInterval <= 0 {
		refreshInterval = defaultRefreshInterval
	}
	resolver := &CachingResolver{
		inner:           inner,
		cache:           cfg.Cache,
		refreshInterval: refreshInterval,
		owners:          make(map[string]struct{}),
		logger:          log.Sugared(logger).Named("CachingResolver"),
	}

	if cfg.Meter != nil {
		var err error
		resolver.cacheLookups, err = cfg.Meter.Int64Counter("org_resolver_cache_lookups")
		if err != nil {
			return nil, fmt.Errorf("failed to create cache lookups metric: %w", err)
		}
		resolver.mappingChanges, err = cfg.Meter.Int64Counter("org_resolver_mapping_changed")
		if err != nil {
			return nil, fmt.Errorf("failed to create mapping changed metric: %w", err)
		}
	}

	return resolver, nil
}

func (c *CachingResolver) Get(ctx context.Context, owner string) (string, error) {
	c.rememberOwner(owner)

	entry, cached := c.lookupCache(ctx, owner)
	if cached {
		return entry.OrgID, nil
	}

	// No cached entry: there is nothing to return without waiting on the
	// underlying resolver.
	orgID, err := c.resolveWithRetry(ctx, owner)
	if err != nil {
		return "", err
	}
	c.storeInCache(ctx, owner, orgID)
	return orgID, nil
}

// rememberOwner records owner so the background refresh loop picks it up.
func (c *CachingResolver) rememberOwner(owner string) {
	c.ownersMu.Lock()
	c.owners[owner] = struct{}{}
	c.ownersMu.Unlock()
}

// refreshLoop wakes up every RefreshInterval and revalidates every known
// owner against the underlying resolver.
func (c *CachingResolver) refreshLoop(ctx context.Context) {
	defer c.wg.Done()
	ticker := time.NewTicker(c.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.refreshAllOwners(ctx)
		}
	}
}

// refreshAllOwners revalidates every owner ever passed to Get against the
// underlying resolver, one at a time, pausing refreshOwnerDelay between calls
// to smooth the load on the underlying resolver.
func (c *CachingResolver) refreshAllOwners(ctx context.Context) {
	c.ownersMu.Lock()
	owners := make([]string, 0, len(c.owners))
	for owner := range c.owners {
		owners = append(owners, owner)
	}
	c.ownersMu.Unlock()

	for i, owner := range owners {
		if i > 0 {
			select {
			case <-ctx.Done():
				return
			case <-time.After(refreshOwnerDelay):
			}
		}
		if ctx.Err() != nil {
			return
		}
		c.refreshOwner(ctx, owner)
	}
}

// refreshOwner revalidates owner against the underlying resolver. A failed
// revalidation is logged and otherwise ignored - the existing entry remains
// cached and trusted.
func (c *CachingResolver) refreshOwner(ctx context.Context, owner string) {
	previous, ok, err := c.cache.Get(ctx, owner)
	if err != nil {
		c.logger.Warnw("Failed to read from cache during refresh; skipping", "owner", owner, "error", err)
		return
	}
	if !ok {
		return
	}

	orgID, err := c.resolveWithRetry(ctx, owner)
	if err != nil {
		c.logger.Warnw("Failed to refresh cached org mapping; keeping existing entry", "owner", owner, "error", err)
		return
	}
	if orgID != previous.OrgID {
		c.logger.Errorw("Org mapping changed for owner", "owner", owner, "previousOrgID", previous.OrgID, "newOrgID", orgID)
		if c.mappingChanges != nil {
			c.mappingChanges.Add(ctx, 1, metric.WithAttributes(attribute.String("owner", owner)))
		}
	}
	c.storeInCache(ctx, owner, orgID)
}

var getRetryBackoff = backoff.Backoff{
	Min:    initialRetryBackoff,
	Max:    maxRetryBackoff,
	Factor: 2,
}

type getResult struct {
	orgID string
	err   error
}

func getRetryStrategy() *retry.Strategy[getResult] {
	return &retry.Strategy[getResult]{
		MaxRetries: maxGetRetries,
		Backoff:    getRetryBackoff.Copy(),
	}
}

// resolveWithRetry calls the underlying resolver, retrying retriable gRPC
// errors with backoff. NotFound and other non-retriable errors are returned
// immediately without retrying.
func (c *CachingResolver) resolveWithRetry(ctx context.Context, owner string) (string, error) {
	result, doErr := getRetryStrategy().Do(ctx, c.logger, func(ctx context.Context) (getResult, error) {
		orgID, err := c.inner.Get(ctx, owner)
		if err == nil {
			return getResult{orgID: orgID}, nil
		}
		if !isRetriableGRPCCode(grpcStatusCode(err)) {
			return getResult{err: err}, nil
		}
		return getResult{}, err
	})
	if doErr != nil {
		if ctx.Err() != nil {
			return "", context.Cause(ctx)
		}
		return "", errors.Unwrap(doErr)
	}
	if result.err != nil {
		return "", result.err
	}
	return result.orgID, nil
}

func isRetriableGRPCCode(code codes.Code) bool {
	return slices.Contains([]codes.Code{codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted, codes.Unknown},
		code)
}

func (c *CachingResolver) lookupCache(ctx context.Context, owner string) (CacheEntry, bool) {
	entry, ok, err := c.cache.Get(ctx, owner)
	if err != nil {
		c.logger.Warnw("Failed to read from cache, treating as cache miss", "owner", owner, "error", err)
		c.recordCacheLookup(ctx, cacheResultError)
		return CacheEntry{}, false
	}
	if !ok {
		c.recordCacheLookup(ctx, cacheResultMiss)
		return CacheEntry{}, false
	}
	c.recordCacheLookup(ctx, cacheResultHit)
	return entry, true
}

func (c *CachingResolver) recordCacheLookup(ctx context.Context, result string) {
	if c.cacheLookups != nil {
		c.cacheLookups.Add(ctx, 1, metric.WithAttributes(attribute.String("result", result)))
	}
}

func (c *CachingResolver) storeInCache(ctx context.Context, owner, orgID string) {
	if err := c.cache.Set(ctx, owner, CacheEntry{OrgID: orgID, RefreshedAt: time.Now()}); err != nil {
		c.logger.Warnw("Failed to write to cache", "owner", owner, "error", err)
	}
}

// grpcStatusCode extracts the gRPC status code from an error, handling
// wrapped errors from fmt.Errorf("%w", ...) chains.
func grpcStatusCode(err error) codes.Code {
	type grpcStatus interface {
		GRPCStatus() *status.Status
	}
	var se grpcStatus
	if ok := errorAs(err, &se); ok {
		return se.GRPCStatus().Code()
	}
	return codes.OK
}

// errorAs is a typed wrapper for the standard errors.As, allowing interface targets.
// Go's errors.As requires a pointer to a concrete or interface type; this helper
// keeps the call site at grpcStatusCode clean.
func errorAs[T any](err error, target *T) bool {
	for err != nil {
		if t, ok := err.(T); ok {
			*target = t
			return true
		}
		err = unwrapErr(err)
	}
	return false
}

func unwrapErr(err error) error {
	type wrapper interface{ Unwrap() error }
	if w, ok := err.(wrapper); ok {
		return w.Unwrap()
	}
	return nil
}

// Start starts the inner resolver and the background refresh loop.
func (c *CachingResolver) Start(ctx context.Context) error {
	if err := c.inner.Start(ctx); err != nil {
		return err
	}

	// The loop must outlive the Start call, so it gets its own context,
	// cancelled explicitly by Close rather than inherited from ctx.
	loopCtx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	c.wg.Add(1)
	go c.refreshLoop(loopCtx)
	return nil
}

// Close stops the refresh loop, waits for an in-flight refresh sweep to
// finish, and closes the inner resolver.
func (c *CachingResolver) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
	return c.inner.Close()
}

func (c *CachingResolver) Ready() error                   { return c.inner.Ready() }
func (c *CachingResolver) HealthReport() map[string]error { return c.inner.HealthReport() }
func (c *CachingResolver) Name() string                   { return c.inner.Name() }
