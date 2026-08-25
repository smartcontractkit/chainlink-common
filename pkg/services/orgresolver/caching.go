package orgresolver

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
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

	// defaultRefreshInterval is how often a cached entry triggers a background
	// revalidation call to the underlying resolver, used when
	// CachingResolverConfig.RefreshInterval is unset.
	defaultRefreshInterval = 10 * time.Minute
	defaultRefreshWorkers  = 4
	refreshQueueSize       = 1024
	// refreshJitterMaxFraction extends the effective refresh interval by up
	// to an extra 50% so cache entries populated around the same time
	// don't all become due for a refresh call at the same instant.
	refreshJitterMaxFraction = 0.5

	cacheResultHit   = "hit"
	cacheResultMiss  = "miss"
	cacheResultError = "error"
)

var getRetryBackoff = backoff.Backoff{
	Min:    initialRetryBackoff,
	Max:    maxRetryBackoff,
	Factor: 2,
}

func getRetryStrategy() *retry.Strategy[getResult] {
	return &retry.Strategy[getResult]{
		MaxRetries: maxGetRetries,
		Backoff:    getRetryBackoff.Copy(),
	}
}

type getResult struct {
	orgID string
	err   error
}

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
	// RefreshInterval controls how often a cached entry triggers a background
	// revalidation call to the underlying resolver. It does not affect
	// whether a cache hit is trusted - a present entry is always returned
	// immediately, refreshed or not. Defaults to DefaultRefreshInterval when
	// zero.
	RefreshInterval time.Duration
	// Workers is the size of the background goroutine pool that processes
	// async cache refreshes. Defaults to DefaultRefreshWorkers when zero.
	Workers int
	// Meter, if provided, is used to record metrics.
	Meter metric.Meter // optional
}

// refreshJob is a request to revalidate owner against the underlying
// resolver, carrying the cache entry that was current when the request was
// queued (used only for the mapping-changed log comparison).
type refreshJob struct {
	owner    string
	previous CacheEntry
}

// CachingResolver wraps an OrgResolver with a Cache of owner->orgID mappings.
//
// An owner's org mapping is not expected to change, so once an entry is
// cached it is trusted indefinitely and returned immediately, without ever
// blocking on a call to the underlying resolver. To guard against a bad
// entry becoming permanent (e.g. from a resolver bug or a transient
// corruption), an entry older than RefreshInterval also queues a
// revalidation against the underlying resolver, processed by a background
// worker pool started by Start and stopped by Close; this never delays the
// Get call that triggered it. If the revalidation fails, the existing cache
// entry is left in place and continues to be served.
type CachingResolver struct {
	inner           OrgResolver
	cache           Cache
	refreshInterval time.Duration
	numWorkers      int
	logger          log.SugaredLogger

	refreshCh    chan refreshJob
	refreshingMu sync.Mutex
	refreshing   map[string]struct{} // owner -> struct{}{} while queued or being refreshed

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
	numWorkers := cfg.Workers
	if numWorkers <= 0 {
		numWorkers = defaultRefreshWorkers
	}
	resolver := &CachingResolver{
		inner:           inner,
		cache:           cfg.Cache,
		refreshInterval: refreshInterval,
		numWorkers:      numWorkers,
		refreshCh:       make(chan refreshJob, refreshQueueSize),
		refreshing:      make(map[string]struct{}),
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
	entry, cached := c.lookupCache(ctx, owner)
	if cached {
		jitteredInterval := c.refreshInterval + time.Duration(rand.Float64()*refreshJitterMaxFraction*float64(c.refreshInterval))
		if time.Since(entry.RefreshedAt) >= jitteredInterval {
			c.queueRefresh(owner, entry)
		}
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

// queueRefresh hands owner off to the background worker pool to revalidate
// against the underlying resolver, so that Get never blocks on a cache hit.
// At most one refresh per owner is queued/running at a time. If the queue is
// full, the request is dropped; it will be retried on a later stale cache
// hit.
func (c *CachingResolver) queueRefresh(owner string, previous CacheEntry) {
	c.refreshingMu.Lock()
	_, inFlight := c.refreshing[owner]
	if !inFlight {
		c.refreshing[owner] = struct{}{}
	}
	c.refreshingMu.Unlock()
	if inFlight {
		return
	}
	select {
	case c.refreshCh <- refreshJob{owner: owner, previous: previous}:
	default:
		c.refreshingMu.Lock()
		delete(c.refreshing, owner)
		c.refreshingMu.Unlock()
		c.logger.Warnw("Refresh queue full; will retry on a later stale cache hit", "owner", owner)
	}
}

func (c *CachingResolver) refreshWorker(ctx context.Context) {
	defer c.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-c.refreshCh:
			c.runRefresh(ctx, job)
		}
	}
}

// runRefresh revalidates job.owner against the underlying resolver. A failed
// revalidation is logged and otherwise ignored - the existing entry remains
// cached and trusted.
func (c *CachingResolver) runRefresh(ctx context.Context, job refreshJob) {
	defer func() {
		c.refreshingMu.Lock()
		delete(c.refreshing, job.owner)
		c.refreshingMu.Unlock()
	}()

	orgID, err := c.resolveWithRetry(ctx, job.owner)
	if err != nil {
		c.logger.Warnw("Failed to refresh cached org mapping; keeping existing entry", "owner", job.owner, "error", err)
		return
	}
	if orgID != job.previous.OrgID {
		c.logger.Errorw("Org mapping changed for owner", "owner", job.owner, "previousOrgID", job.previous.OrgID, "newOrgID", orgID)
		if c.mappingChanges != nil {
			c.mappingChanges.Add(ctx, 1, metric.WithAttributes(attribute.String("owner", job.owner)))
		}
	}
	c.storeInCache(ctx, job.owner, orgID)
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

// Start starts the inner resolver and the background refresh worker pool.
func (c *CachingResolver) Start(ctx context.Context) error {
	if err := c.inner.Start(ctx); err != nil {
		return err
	}

	// Workers must outlive the Start call, so they get their own context,
	// cancelled explicitly by Close rather than inherited from ctx.
	workerCtx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel
	for range c.numWorkers {
		c.wg.Add(1)
		go c.refreshWorker(workerCtx)
	}
	return nil
}

// Close stops the refresh worker pool, waits for in-flight refreshes to
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
