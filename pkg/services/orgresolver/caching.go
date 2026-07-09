package orgresolver

import (
	"context"
	"sync"
	"time"

	log "github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// CachingOrgResolver wraps an OrgResolver with a fully in-memory cache that is
// refreshed on a background ticker. It is the resolver the snapshot path uses:
// GetUtilization is contractually no-network, so snapshot org attribution must
// be servable from memory without a synchronous linking-service call.
//
// It tracks every owner it has ever resolved and, on each refresh tick,
// re-resolves all of them through the inner resolver, replacing cached values
// (a failed refresh keeps the last known value). This bounds org staleness to
// refreshInterval rather than caching indefinitely. Get is served from the
// cache; a cold miss (an owner never seen before) falls through to the inner
// resolver once and populates the cache.
type CachingOrgResolver struct {
	inner           OrgResolver
	refreshInterval time.Duration
	logger          log.SugaredLogger

	mu    sync.RWMutex
	cache map[string]string

	stopOnce sync.Once
	stop     chan struct{}
	wg       sync.WaitGroup
}

// NewCaching returns a CachingOrgResolver wrapping inner. When refreshInterval
// is > 0, a background goroutine (started on Start, stopped on Close) refreshes
// every known owner every refreshInterval, bounding staleness to that interval.
func NewCaching(inner OrgResolver, refreshInterval time.Duration) *CachingOrgResolver {
	return &CachingOrgResolver{
		inner:           inner,
		refreshInterval: refreshInterval,
		logger:          log.Sugared(log.Nop()).Named("CachingOrgResolver"),
		cache:           make(map[string]string),
		stop:            make(chan struct{}),
	}
}

// Get serves owner from the cache. On a cold miss it falls through to the inner
// resolver once and, on success, caches the result so subsequent Gets (and the
// background refresh) are served from memory.
func (c *CachingOrgResolver) Get(ctx context.Context, owner string) (string, error) {
	c.mu.RLock()
	orgID, ok := c.cache[owner]
	c.mu.RUnlock()
	if ok {
		return orgID, nil
	}

	orgID, err := c.inner.Get(ctx, owner)
	if err != nil {
		return "", err
	}
	c.mu.Lock()
	c.cache[owner] = orgID
	c.mu.Unlock()
	return orgID, nil
}

func (c *CachingOrgResolver) Start(ctx context.Context) error {
	if err := c.inner.Start(ctx); err != nil {
		return err
	}
	if c.refreshInterval > 0 {
		c.wg.Add(1)
		go c.refreshLoop()
	}
	return nil
}

func (c *CachingOrgResolver) refreshLoop() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			c.refreshAll()
		}
	}
}

// refreshAll re-resolves every known owner through the inner resolver. Each
// refresh is bounded by refreshInterval; a failed refresh keeps the last known
// value so a transient linking-service outage never blanks org attribution.
func (c *CachingOrgResolver) refreshAll() {
	c.mu.RLock()
	owners := make([]string, 0, len(c.cache))
	for owner := range c.cache {
		owners = append(owners, owner)
	}
	c.mu.RUnlock()

	for _, owner := range owners {
		select {
		case <-c.stop:
			return
		default:
		}
		ctx, cancel := context.WithTimeout(context.Background(), c.refreshInterval)
		orgID, err := c.inner.Get(ctx, owner)
		cancel()
		if err != nil {
			c.logger.Debugw("org refresh failed; keeping cached value", "owner", owner, "error", err)
			continue
		}
		c.mu.Lock()
		c.cache[owner] = orgID
		c.mu.Unlock()
	}
}

func (c *CachingOrgResolver) Close() error {
	c.stopOnce.Do(func() { close(c.stop) })
	c.wg.Wait()
	return c.inner.Close()
}

func (c *CachingOrgResolver) Ready() error                   { return c.inner.Ready() }
func (c *CachingOrgResolver) HealthReport() map[string]error { return c.inner.HealthReport() }
func (c *CachingOrgResolver) Name() string                   { return c.logger.Name() }
