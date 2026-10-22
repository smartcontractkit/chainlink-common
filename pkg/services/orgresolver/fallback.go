package orgresolver

import (
	"context"
	"errors"
	"sync"

	"google.golang.org/grpc/codes"

	log "github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// Deprecated: Use CachingResolver
type OrgResolverFallback struct {
	inner  OrgResolver
	cache  sync.Map // owner (string) -> orgID (string)
	logger log.SugaredLogger
}

// Deprecated: Use NewCachingResolver
func NewOrgResolverWithFallback(inner OrgResolver, logger log.Logger) *OrgResolverFallback {
	return &OrgResolverFallback{
		inner:  inner,
		logger: log.Sugared(logger).Named("OrgResolverFallback"),
	}
}

func (c *OrgResolverFallback) Get(ctx context.Context, owner string) (string, error) {
	result, err := getRetryStrategy().Do(ctx, c.logger, func(ctx context.Context) (getResult, error) {
		orgID, err := c.inner.Get(ctx, owner)
		if err == nil {
			c.cache.Store(owner, orgID)
			return getResult{orgID: orgID}, nil
		}

		code := grpcStatusCode(err)
		if code == codes.NotFound {
			orgID, err := c.fallbackToCache(owner, err)
			if err != nil {
				return getResult{err: err}, nil
			}
			return getResult{orgID: orgID}, nil
		}
		if !isRetriableGRPCCode(code) {
			return getResult{err: err}, nil
		}
		return getResult{}, err
	})
	if err == nil {
		if result.err != nil {
			return "", result.err
		}
		return result.orgID, nil
	}

	if ctx.Err() != nil {
		return c.fallbackToCache(owner, context.Cause(ctx))
	}
	return c.fallbackToCache(owner, errors.Unwrap(err))
}

func (c *OrgResolverFallback) fallbackToCache(owner string, originalErr error) (string, error) {
	if cached, ok := c.cache.Load(owner); ok {
		orgID := cached.(string)
		c.logger.Infow("Using cached org ID after resolver failure", "owner", owner, "cachedOrgID", orgID)
		return orgID, nil
	}
	return "", originalErr
}

func (c *OrgResolverFallback) Start(ctx context.Context) error { return c.inner.Start(ctx) }
func (c *OrgResolverFallback) Close() error                    { return c.inner.Close() }
func (c *OrgResolverFallback) Ready() error                    { return c.inner.Ready() }
func (c *OrgResolverFallback) HealthReport() map[string]error  { return c.inner.HealthReport() }
func (c *OrgResolverFallback) Name() string                    { return c.inner.Name() }
