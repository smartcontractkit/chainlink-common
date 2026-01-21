package grpcsource

import (
	"context"
	"crypto/tls"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	nodeauthgrpc "github.com/smartcontractkit/chainlink-common/pkg/nodeauth/grpc"
	auth "github.com/smartcontractkit/chainlink-common/pkg/nodeauth/jwt"
	pb "github.com/smartcontractkit/chainlink-protos/workflows/go/sources"
)

// Client is a GRPC client for the WorkflowMetadataSourceService.
type Client struct {
	conn         *grpc.ClientConn
	client       pb.WorkflowMetadataSourceServiceClient
	name         string
	jwtGenerator auth.JWTGenerator
}

// clientConfig holds configuration for the client
type clientConfig struct {
	tlsEnabled   bool
	jwtGenerator auth.JWTGenerator
}

// ClientOption configures the Client
type ClientOption func(*clientConfig)

func WithTLS(enabled bool) ClientOption {
	return func(c *clientConfig) {
		c.tlsEnabled = enabled
	}
}

func WithJWTGenerator(generator auth.JWTGenerator) ClientOption {
	return func(c *clientConfig) {
		c.jwtGenerator = generator
	}
}

// NewClient creates a new GRPC client for the WorkflowMetadataSourceService.
// addr is the GRPC endpoint address (e.g., "localhost:50051").
// name is a human-readable identifier for logging.
// opts are optional configuration options.
func NewClient(addr string, name string, opts ...ClientOption) (*Client, error) {
	cfg := &clientConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	var dialOpts []grpc.DialOption
	if cfg.tlsEnabled {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})))
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(addr, dialOpts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:         conn,
		client:       pb.NewWorkflowMetadataSourceServiceClient(conn),
		name:         name,
		jwtGenerator: cfg.jwtGenerator,
	}, nil
}

// NewClientWithOptions creates a new GRPC client with custom dial options.
// This is useful for testing or when custom options are needed.
func NewClientWithOptions(addr string, name string, opts ...grpc.DialOption) (*Client, error) {
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: pb.NewWorkflowMetadataSourceServiceClient(conn),
		name:   name,
	}, nil
}

func (c *Client) addJWTAuth(ctx context.Context, req any) (context.Context, error) {
	if c.jwtGenerator == nil {
		return ctx, nil // Skip if no generator configured
	}

	jwtToken, err := c.jwtGenerator.CreateJWTForRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT: %w", err)
	}

	return metadata.AppendToOutgoingContext(ctx, nodeauthgrpc.AuthorizationHeader, nodeauthgrpc.BearerPrefix+jwtToken), nil
}

// ListWorkflowMetadata fetches workflow metadata from the GRPC source.
// families is the list of DON families to filter workflows by.
// start is the pagination offset (0-indexed).
// limit is the maximum number of workflows to return per page.
// Returns workflows, hasMore flag indicating if more pages exist, and error.
func (c *Client) ListWorkflowMetadata(ctx context.Context, families []string, start, limit int64) ([]*pb.WorkflowMetadata, bool, error) {
	req := &pb.ListWorkflowMetadataRequest{
		DonFamilies: families,
		Start:       start,
		Limit:       limit,
	}

	// Inject JWT auth
	ctx, err := c.addJWTAuth(ctx, req)
	if err != nil {
		return nil, false, err
	}

	resp, err := c.client.ListWorkflowMetadata(ctx, req)
	if err != nil {
		return nil, false, err
	}
	return resp.Workflows, resp.HasMore, nil
}

// Close closes the underlying GRPC connection.
func (c *Client) Close() error {
	return c.conn.Close()
}

// Name returns the human-readable name of this client.
func (c *Client) Name() string {
	return c.name
}
