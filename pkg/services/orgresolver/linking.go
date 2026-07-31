package orgresolver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	log "github.com/smartcontractkit/chainlink-common/pkg/logger"
	nodeauthgrpc "github.com/smartcontractkit/chainlink-common/pkg/nodeauth/grpc"
	"github.com/smartcontractkit/chainlink-common/pkg/services"
	linkingclient "github.com/smartcontractkit/chainlink-protos/linking-service/go/v1"
)

// defaultRequestTimeout bounds Get(); callers sit in capability hot paths, so a hung connection must not propagate.
const defaultRequestTimeout = 2 * time.Second

// JWTGenerator interface for JWT token creation
type JWTGenerator interface {
	CreateJWTForRequest(req any) (string, error)
}

// OrgResolver interface defines methods for resolving organization IDs from workflow owners
type OrgResolver interface {
	services.Service
	Get(ctx context.Context, owner string) (string, error)
}

// CacheStore persists owner->orgID mappings to durable storage (e.g. Postgres).
// Implementations are provided by core and must be safe for concurrent use.
type CacheStore interface {
	// GetOrg returns the cached orgID for owner, or sql.ErrNoRows if absent.
	GetOrg(ctx context.Context, owner string) (string, error)
	// UpsertOrg stores the owner->orgID mapping.
	UpsertOrg(ctx context.Context, owner, orgID string) error
}

type Config struct {
	URL                           string
	TLSEnabled                    bool
	WorkflowRegistryAddress       string
	WorkflowRegistryChainSelector uint64
	JWTGenerator                  JWTGenerator

	RequestTimeout time.Duration // bounds each Get() call; zero means defaultRequestTimeout

	Client linkingclient.LinkingServiceClient // optional
	Meter  metric.Meter                       // optional

	// CacheEnabled turns on durable caching of owner->orgID mappings via CacheStore.
	CacheEnabled bool
	// CacheStore is required when CacheEnabled is true.
	CacheStore CacheStore
}

// orgResolver makes direct calls to the linking service to resolve organization IDs from workflow owners.
// This simplified implementation makes a network call for each Get() request, bounded by requestTimeout.
type orgResolver struct {
	workflowRegistryAddress       string
	workflowRegistryChainSelector uint64

	client         linkingclient.LinkingServiceClient
	conn           *grpc.ClientConn // nil if client was injected
	logger         log.SugaredLogger
	jwtGenerator   JWTGenerator
	requestTimeout time.Duration

	cacheEnabled bool
	cacheStore   CacheStore

	passCount    metric.Int64Counter
	failCount    metric.Int64Counter
	cacheLookups metric.Int64Counter // tagged with result=hit|miss|error
}

// cacheLookupResult is the attribute key for cache lookup outcomes.
var cacheResultAttr = "result"

const (
	cacheResultHit  = "hit"
	cacheResultMiss = "miss"
)

// ErrCacheMiss is returned by CacheStore.GetOrg when no mapping exists for owner.
// Stores backed by SQL may return sql.ErrNoRows; both are treated as a miss.
var ErrCacheMiss = errors.New("org not found in cache")

// NewOrgResolver creates a new org resolver with the specified configuration
// Deprecated: Use Config.New
//
//go:fix inline
func NewOrgResolver(cfg Config, logger log.Logger) (*orgResolver, error) {
	return cfg.New(logger)
}

// NewOrgResolverWithClient creates a new org resolver with an optional injected client (for testing)
// Deprecated: Use Config.New
//
//go:fix inlin
func NewOrgResolverWithClient(cfg Config, client linkingclient.LinkingServiceClient, logger log.Logger) (*orgResolver, error) {
	cfg.Client = client
	return cfg.New(logger)
}

func (cfg *Config) New(logger log.Logger) (*orgResolver, error) {
	requestTimeout := cfg.RequestTimeout
	if requestTimeout <= 0 {
		requestTimeout = defaultRequestTimeout
	}

	if cfg.CacheEnabled && cfg.CacheStore == nil {
		return nil, errors.New("CacheStore is required when CacheEnabled is true")
	}

	resolver := &orgResolver{
		workflowRegistryAddress:       cfg.WorkflowRegistryAddress,
		workflowRegistryChainSelector: cfg.WorkflowRegistryChainSelector,
		logger:                        log.Sugared(logger).Named("OrgResolver"),
		jwtGenerator:                  cfg.JWTGenerator,
		requestTimeout:                requestTimeout,
		cacheEnabled:                  cfg.CacheEnabled,
		cacheStore:                    cfg.CacheStore,
	}

	if cfg.Client != nil {
		resolver.client = cfg.Client
	} else {
		if cfg.URL == "" {
			return nil, errors.New("URL is required when client is not provided")
		}

		var opts []grpc.DialOption
		if cfg.TLSEnabled {
			opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
		} else {
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}

		conn, err := grpc.NewClient(cfg.URL, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create linking service client at %s: %w", cfg.URL, err)
		}

		resolver.conn = conn
		resolver.client = linkingclient.NewLinkingServiceClient(conn)
	}

	if cfg.Meter != nil {
		var err error
		resolver.passCount, err = cfg.Meter.Int64Counter("org_resolver_success")
		if err != nil {
			return nil, fmt.Errorf("failed to create success count metric: %w", err)
		}
		resolver.failCount, err = cfg.Meter.Int64Counter("org_resolver_fail")
		if err != nil {
			return nil, fmt.Errorf("failed to create failure count metric: %w", err)
		}
		if resolver.cacheEnabled {
			resolver.cacheLookups, err = cfg.Meter.Int64Counter("org_resolver_cache_lookups")
			if err != nil {
				return nil, fmt.Errorf("failed to create cache lookups metric: %w", err)
			}
		}
	}

	return resolver, nil
}

// addJWTAuth creates and signs a JWT token, then adds it to the context
func (o *orgResolver) addJWTAuth(ctx context.Context, req any) (context.Context, error) {
	// Skip authentication if no JWT generator provided
	if o.jwtGenerator == nil {
		return ctx, nil
	}

	// Create JWT token using the JWT generator
	jwtToken, err := o.jwtGenerator.CreateJWTForRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT: %w", err)
	}

	// Add JWT to Authorization header
	return metadata.AppendToOutgoingContext(ctx, nodeauthgrpc.AuthorizationHeader, nodeauthgrpc.BearerPrefix+jwtToken), nil
}

func (o *orgResolver) Get(ctx context.Context, owner string) (string, error) {
	if o.cacheEnabled {
		if orgID, ok := o.checkCache(ctx, owner); ok {
			return orgID, nil
		}
	}

	ctx, cancel := context.WithTimeout(ctx, o.requestTimeout)
	defer cancel()

	req := &linkingclient.GetOrganizationFromWorkflowOwnerRequest{
		WorkflowOwner:           owner,
		WorkflowRegistryAddress: o.workflowRegistryAddress,
		ChainSelector:           o.workflowRegistryChainSelector,
	}

	ctx, err := o.addJWTAuth(ctx, req)
	if err != nil {
		o.logger.Errorw("Failed to add JWT auth to GetOrganizationFromWorkflowOwner request", "error", err)
		return "", err
	}

	resp, err := o.client.GetOrganizationFromWorkflowOwner(ctx, req)
	if err != nil {
		if o.failCount != nil {
			o.failCount.Add(ctx, 1)
		}
		return "", fmt.Errorf("failed to fetch organization from workflow owner: %w", err)
	}

	if o.passCount != nil {
		o.passCount.Add(ctx, 1)
	}

	if o.cacheEnabled {
		o.storeInCache(ctx, owner, resp.OrganizationId)
	}
	return resp.OrganizationId, nil
}

// checkCache looks up owner in the durable cache. Returns (orgID, true) on hit.
// A cache store error is logged and treated as a miss so lookups remain resilient.
func (o *orgResolver) checkCache(ctx context.Context, owner string) (string, bool) {
	orgID, err := o.cacheStore.GetOrg(ctx, owner)
	if err != nil {
		if errors.Is(err, ErrCacheMiss) || errors.Is(err, sql.ErrNoRows) {
			o.recordCacheLookup(ctx, cacheResultMiss)
		} else {
			o.logger.Warnw("Failed to read org from cache store, falling back to linking service", "owner", owner, "error", err)
			o.recordCacheLookup(ctx, "error")
		}
		return "", false
	}
	o.recordCacheLookup(ctx, cacheResultHit)
	return orgID, true
}

// storeInCache persists the owner->orgID mapping. Failures are logged but not
// propagated so a store hiccup does not break resolution.
func (o *orgResolver) storeInCache(ctx context.Context, owner, orgID string) {
	if err := o.cacheStore.UpsertOrg(ctx, owner, orgID); err != nil {
		o.logger.Warnw("Failed to persist org to cache store", "owner", owner, "error", err)
	}
}

func (o *orgResolver) recordCacheLookup(ctx context.Context, result string) {
	if o.cacheLookups != nil {
		o.cacheLookups.Add(ctx, 1, metric.WithAttributes(attribute.String(cacheResultAttr, result)))
	}
}

func (o *orgResolver) Start(_ context.Context) error {
	return nil
}

func (o *orgResolver) HealthReport() map[string]error {
	return map[string]error{o.Name(): nil}
}

func (o *orgResolver) Close() error {
	if o.conn != nil {
		return o.conn.Close()
	}
	return nil
}

func (o *orgResolver) Name() string {
	return o.logger.Name()
}

func (o *orgResolver) Ready() error {
	return nil
}
