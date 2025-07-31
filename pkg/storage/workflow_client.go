package storage

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	nodeauth "github.com/smartcontractkit/chainlink-common/pkg/nodeauth/jwt"
	pb "github.com/smartcontractkit/chainlink-protos/storage-service/go"
)

// WorkflowClient is a gRPC client for the node service to be used by workflow node.
type WorkflowClient interface {
	// DownloadArtifact downloads an artifact from the storage service
	DownloadArtifact(ctx context.Context, req *pb.DownloadArtifactRequest) (pb.NodeService_DownloadArtifactClient, error)

	// DownloadArtifactStream streams artifact chunks to a callback as they are received
	DownloadArtifactStream(ctx context.Context, req *pb.DownloadArtifactRequest, onChunk func(*pb.DownloadArtifactChunk) error) error

	// Close closes the gRPC connection
	Close() error
}

// workflowClient is a concrete implementation of WorkflowClient
type workflowClient struct {
	client       pb.NodeServiceClient
	conn         *grpc.ClientConn
	log          logger.Logger
	jwtGenerator *nodeauth.NodeJWTGenerator
}

// DownloadArtifact downloads an artifact from the storage service and returns the raw stream
func (n workflowClient) DownloadArtifact(ctx context.Context, req *pb.DownloadArtifactRequest) (pb.NodeService_DownloadArtifactClient, error) {
	if n.jwtGenerator != nil {
		var err error
		ctx, err = n.injectToken(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	n.log.Infow("DownloadArtifact RPC called", "id", req.Id, "type", req.Type.String(), "environment", req.Environment.String())
	return n.client.DownloadArtifact(ctx, req)
}

// DownloadArtifactStream streams artifact chunks to a callback as they are received
func (n workflowClient) DownloadArtifactStream(ctx context.Context, req *pb.DownloadArtifactRequest, onChunk func(*pb.DownloadArtifactChunk) error) error {
	if n.jwtGenerator != nil {
		var err error
		ctx, err = n.injectToken(ctx, req)
		if err != nil {
			return err
		}
	}
	n.log.Infow("DownloadArtifactStream RPC called", "id", req.Id, "type", req.Type.String(), "environment", req.Environment.String())

	stream, err := n.client.DownloadArtifact(ctx, req)
	if err != nil {
		n.log.Errorw("Failed to initiate artifact download stream", "id", req.Id, "error", err)
		return err
	}

	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			n.log.Errorw("Error while receiving artifact chunk", "id", req.Id, "error", err)
			return err
		}
		if err := onChunk(chunk); err != nil {
			n.log.Errorw("Error in chunk callback", "id", req.Id, "error", err)
			return err
		}
	}

	n.log.Infow("Successfully streamed artifact", "id", req.Id)
	return nil
}
func (n workflowClient) Close() error {
	err := n.conn.Close()
	if err != nil {
		n.log.Errorw("Failed to close WorkflowClient connection", "error", err)
		return err
	}
	n.log.Infow("Closed WorkflowClient connection")
	return nil
}

// NodeClientOpt is a functional option type for configuring the WorkflowClient
type NodeClientOpt func(*nodeConfig)

type nodeConfig struct {
	log                  logger.Logger
	transportCredentials credentials.TransportCredentials
	jwtGenerator         *nodeauth.NodeJWTGenerator // Optional JWT manager for authentication
}

// defaultNodeConfig returns a default configuration for the WorkflowClient
func defaultNodeConfig() nodeConfig {
	loggerInst, _ := logger.New()
	return nodeConfig{
		log:                  loggerInst,
		transportCredentials: credentials.NewTLS(nil), // Use default TLS credentials
		jwtGenerator:         nil,                     // No JWT generator by default
	}
}

// WithLogger sets the logger for the WorkflowClient
func WithLogger(log logger.Logger) NodeClientOpt {
	return func(cfg *nodeConfig) {
		cfg.log = log
	}
}

// WithTransportCredentials sets the transport credentials for the WorkflowClient
func WithTransportCredentials(creds credentials.TransportCredentials) NodeClientOpt {
	return func(cfg *nodeConfig) {
		cfg.transportCredentials = creds
	}
}

func WithJWTGenerator(jwtGenerator *nodeauth.NodeJWTGenerator) NodeClientOpt {
	return func(cfg *nodeConfig) {
		cfg.jwtGenerator = jwtGenerator
	}
}

// NewNodeClient creates a new WorkflowClient with the specified address and options
// It returns a WorkflowClient which can DownloadArtifacts
func NewNodeClient(ctx context.Context, address string, opts ...NodeClientOpt) (WorkflowClient, error) {
	cfg := defaultNodeConfig()

	for _, opt := range opts {
		opt(&cfg)
	}

	grpcOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(cfg.transportCredentials),
	}

	conn, err := grpc.NewClient(address, grpcOpts...)
	if err != nil {
		return nil, err
	}

	cfg.log.Infow("connected to storage service (WorkflowClient)", "address", address)

	client := pb.NewNodeServiceClient(conn)

	cfg.log.Infow("connected to storage service (NodeClient)", "address", address)

	return &workflowClient{
		client: client,
		conn:   conn,
		log:    cfg.log,
	}, nil
}

func (c workflowClient) injectToken(ctx context.Context, req any) (context.Context, error) {
	if c.jwtGenerator == nil {
		c.log.Warnw("authentication: no JWT generator")
		return ctx, fmt.Errorf("authentication: no JWT generator available")
	}
	token, err := c.jwtGenerator.CreateJWTForRequest(req)
	if err != nil {
		return ctx, fmt.Errorf("failed to generate JWT token: %w", err)
	}

	// Inject the token into the context
	ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))

	// Log the token injection
	c.log.Infow("JWT token injected into context")

	return ctx, nil
}
