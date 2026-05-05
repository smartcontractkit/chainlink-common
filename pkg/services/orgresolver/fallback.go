package orgresolver

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	log "github.com/smartcontractkit/chainlink-common/pkg/logger"
)

const unavailableRetryDelay = 500 * time.Millisecond

// OrgResolverFallback wraps an OrgResolver and maintains an in-memory cache of
// owner->orgID mappings. On successful resolution the cache is updated. When
// the inner resolver returns NotFound or Unavailable, the cache is consulted
// as a fallback (with one retry for Unavailable before falling back).
//
// This addresses a race condition where a workflow owner can be unlinked from
// an org just before a WorkflowDeleted event is processed, causing the
// resolver to return NotFound for an owner whose org was previously known.
type OrgResolverFallback struct {
	inner  OrgResolver
	cache  sync.Map // owner (string) -> orgID (string)
	logger log.SugaredLogger
}

func NewOrgResolverWithFallback(inner OrgResolver, logger log.Logger) *OrgResolverFallback {
	return &OrgResolverFallback{
		inner:  inner,
		logger: log.Sugared(logger).Named("OrgResolverFallback"),
	}
}

func (c *OrgResolverFallback) Get(ctx context.Context, owner string) (string, error) {
	orgID, err := c.inner.Get(ctx, owner)
	if err == nil {
		c.cache.Store(owner, orgID)
		return orgID, nil
	}

	code := grpcStatusCode(err)

	if code == codes.Unavailable {
		c.logger.Warnw("Org resolver unavailable, retrying once", "owner", owner, "err", err)

		select {
		case <-ctx.Done():
			return c.fallbackToCache(owner, err)
		case <-time.After(unavailableRetryDelay):
		}

		orgID, retryErr := c.inner.Get(ctx, owner)
		if retryErr == nil {
			c.cache.Store(owner, orgID)
			return orgID, nil
		}
		c.logger.Warnw("Org resolver retry failed", "owner", owner, "err", retryErr)
		return c.fallbackToCache(owner, err)
	}

	if code == codes.NotFound {
		return c.fallbackToCache(owner, err)
	}

	return "", err
}

func (c *OrgResolverFallback) fallbackToCache(owner string, originalErr error) (string, error) {
	if cached, ok := c.cache.Load(owner); ok {
		orgID := cached.(string)
		c.logger.Infow("Using cached org ID after resolver failure", "owner", owner, "cachedOrgID", orgID)
		return orgID, nil
	}
	return "", originalErr
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

func (c *OrgResolverFallback) Start(ctx context.Context) error { return c.inner.Start(ctx) }
func (c *OrgResolverFallback) Close() error                    { return c.inner.Close() }
func (c *OrgResolverFallback) Ready() error                    { return c.inner.Ready() }
func (c *OrgResolverFallback) HealthReport() map[string]error  { return c.inner.HealthReport() }
func (c *OrgResolverFallback) Name() string                    { return c.inner.Name() }
