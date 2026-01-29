package orgresolver

import (
	"context"
	"sync"
	"time"
)

const (
	cacheDurationOrg   = 24 * time.Hour
	cacheDurationError = time.Minute
)

type entry struct {
	orgID      string
	err        error
	expiration time.Time
}

var _ OrgResolver = (*cached)(nil)

type cached struct {
	OrgResolver
	mu         sync.RWMutex
	m          map[string]entry
	stop, done chan struct{}
}

func NewCache(orgResolver OrgResolver) OrgResolver {
	return &cached{OrgResolver: orgResolver, m: make(map[string]entry)}
}

func (c *cached) Start(ctx context.Context) error {
	err := c.OrgResolver.Start(ctx)
	if err == nil {
		go c.reaper()
	}
	return err
}

func (c *cached) Close() error {
	close(c.stop)
	return c.OrgResolver.Close()
}

func (c *cached) reaper() {
	defer close(c.done)
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		select {
		case <-c.stop:
			return
		case <-t.C:
			c.reap()
		}
	}
}
func (c *cached) reap() {
	for owner, e := range c.m {
		if e.expiration.Before(time.Now()) {
			delete(c.m, owner)
		}
	}
}

func (c *cached) Get(ctx context.Context, owner string) (string, error) {
	// fast path
	c.mu.RLock()
	e, ok := c.m[owner]
	if ok && e.expiration.After(time.Now()) {
		c.mu.RUnlock()
		return e.orgID, e.err
	}
	c.mu.RUnlock()

	// maybe update
	c.mu.Lock()
	defer c.mu.Unlock()
	if ok && e.expiration.After(time.Now()) {
		return e.orgID, e.err // already updated
	}

	// update
	orgID, err := c.OrgResolver.Get(ctx, owner)
	if err != nil {
		c.m[owner] = entry{
			err:        err,
			expiration: time.Now().Add(cacheDurationError),
		}
	} else {
		c.m[owner] = entry{
			orgID:      orgID,
			expiration: time.Now().Add(cacheDurationOrg),
		}
	}

	return orgID, err
}
