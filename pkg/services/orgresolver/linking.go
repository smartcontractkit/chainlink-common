package orgresolver

import (
	"context"
	"errors"
	"fmt"

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

// JWTGenerator interface for JWT token creation
type JWTGenerator interface {
	CreateJWTForRequest(req any) (string, error)
}

// OrgResolver interface defines methods for resolving organization IDs from workflow owners
type OrgResolver interface {
	services.Service
	Get(ctx context.Context, owner string) (string, error)
}

type Config struct {
	URL                           string
	TLSEnabled                    bool
	WorkflowRegistryAddress       string
	WorkflowRegistryChainSelector uint64
	JWTGenerator                  JWTGenerator

	Client linkingclient.LinkingServiceClient // optional
	Meter  metric.Meter                       // optional
}

// orgResolver makes direct calls to the linking service to resolve organization IDs from workflow owners.
// This simplified implementation makes a network call for each Get() request.
type orgResolver struct {
	workflowRegistryAddress       string
	workflowRegistryChainSelector uint64

	client       linkingclient.LinkingServiceClient
	conn         *grpc.ClientConn // nil if client was injected
	logger       log.SugaredLogger
	jwtGenerator JWTGenerator

	passCount metric.Int64Counter
	failCount metric.Int64Counter
}

// NewOrgResolver creates a new org resolver with the specified configuration
// Deprecated: Use Config.New
func NewOrgResolver(cfg Config, logger log.Logger) (*orgResolver, error) {
	return cfg.New(logger)
}

// NewOrgResolverWithClient creates a new org resolver with an optional injected client (for testing)
// Deprecated: Use Config.New
func NewOrgResolverWithClient(cfg Config, client linkingclient.LinkingServiceClient, logger log.Logger) (*orgResolver, error) {
	cfg.Client = client
	return cfg.New(logger)
}

func (cfg *Config) New(logger log.Logger) (*orgResolver, error) {
	resolver := &orgResolver{
		workflowRegistryAddress:       cfg.WorkflowRegistryAddress,
		workflowRegistryChainSelector: cfg.WorkflowRegistryChainSelector,
		logger:                        log.Sugared(logger).Named("OrgResolver"),
		jwtGenerator:                  cfg.JWTGenerator,
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
	return resp.OrganizationId, nil
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
