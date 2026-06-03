package orgresolver

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/jpillora/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	log "github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/utils/retry"
)

const (
	maxGetRetries       = 3
	initialRetryBackoff = 100 * time.Millisecond
	maxRetryBackoff     = 1 * time.Second
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

// getResult carries the outcome of a single Get attempt. When err is non-nil the
// attempt is terminal; the retry loop must receive a nil error to stop.
type getResult struct {
	orgID string
	err   error
}

// OrgResolverFallback wraps an OrgResolver and maintains an in-memory cache of
// owner->orgID mappings. On successful resolution the cache is updated. When
// the inner resolver returns NotFound or a retriable gRPC error, the cache is
// consulted as a fallback (after bounded retries for retriable errors).
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

func isRetriableGRPCCode(code codes.Code) bool {
	return slices.Contains([]codes.Code{codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted, codes.Unknown},
		code)
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
