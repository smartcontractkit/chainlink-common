package storage

import (
	"context"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	nodeauth "github.com/smartcontractkit/chainlink-common/pkg/nodeauth/jwt"
	pb "github.com/smartcontractkit/chainlink-protos/storage-service/go"
)

// WorkflowClient is a gRPC client for the storage service (node) to be used by workflow node.
type WorkflowClient interface {
	// DownloadArtifact downloads an artifact from the storage service
	DownloadArtifact(ctx context.Context, req *pb.DownloadArtifactRequest) (*pb.DownloadArtifactResponse, error)

	// Close closes the gRPC connection
	Close() error
}

// workflowClient is a concrete implementation of WorkflowClient
type workflowClient struct {
	address string
	client  pb.NodeServiceClient
	conn    *grpc.ClientConn
	logger  logger.Logger
	tlsCert string
	creds   credentials.TransportCredentials
	// serverName is the expected server name in the TLS certificate.
	serverName string

	// JWT-based authentication
	jwtGenerator nodeauth.JWTGenerator
}

func (wc *workflowClient) downloadArtifact(ctx context.Context, req *pb.DownloadArtifactRequest) (*pb.DownloadArtifactResponse, error) {
	if wc.jwtGenerator != nil {
		var err error
		ctx, err = wc.injectToken(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	return wc.client.DownloadArtifact(ctx, req)
}

// DownloadArtifact returns pre-signed URL for downloading artifacts
func (wc *workflowClient) DownloadArtifact(ctx context.Context, req *pb.DownloadArtifactRequest) (*pb.DownloadArtifactResponse, error) {
	wc.logger.Infow("DownloadArtifact RPC called", "id", req.Id, "type", req.Type.String(), "environment", req.Environment.String())
	return wc.downloadArtifact(ctx, req)
}

func (wc *workflowClient) Close() error {
	err := wc.conn.Close()
	if err != nil {
		wc.logger.Errorw("Failed to close WorkflowClient connection", "error", err)
		return err
	}
	wc.logger.Infow("Closed WorkflowClient connection")
	return nil
}

type workflowConfig struct {
	log                  logger.Logger
	transportCredentials credentials.TransportCredentials
	tlsCert              string
	serverName           string
	jwtGenerator         nodeauth.JWTGenerator // Optional JWT manager for authentication
}

// WorkflowClientOpt is a functional option type for configuring the WorkflowClient
type WorkflowClientOpt func(*workflowConfig)

// defaultNodeConfig returns a default configuration for the WorkflowClient
func defaultWorkflowConfig() workflowConfig {
	// By default, no JWT manager is set and we fallback to insecure creds.
	return workflowConfig{
		transportCredentials: insecure.NewCredentials(),
		tlsCert:              "",
		// Default to "localhost" if not overridden.
		serverName:   "localhost",
		jwtGenerator: nil,
	}
}

func WithWorkflowTransportCredentials(creds credentials.TransportCredentials) WorkflowClientOpt {
	return func(cfg *workflowConfig) {
		cfg.transportCredentials = creds
	}
}

// WithJWTGenerator sets the JWT generator for authentication.
func WithJWTGenerator(jwtGenerator nodeauth.JWTGenerator) WorkflowClientOpt {
	return func(cfg *workflowConfig) {
		cfg.jwtGenerator = jwtGenerator
	}
}

func WithWorkflowTLSCert(cert string) WorkflowClientOpt {
	return func(cfg *workflowConfig) {
		cfg.tlsCert = cert
	}
}

// WithServerName allows overriding the expected server name in the TLS certificate.
func WithServerName(name string) WorkflowClientOpt {
	return func(cfg *workflowConfig) {
		cfg.serverName = name
	}
}

// NewWorkflowClient creates a new WorkflowClient with the specified address and options
// It returns a WorkflowClient which can DownloadArtifacts
func NewWorkflowClient(lggr logger.Logger, address string, opts ...WorkflowClientOpt) (WorkflowClient, error) {
	cfg := defaultWorkflowConfig()

	for _, opt := range opts {
		opt(&cfg)
	}

	wc := &workflowClient{
		address:      address,
		logger:       logger.Named(lggr, "WorkflowClient"),
		tlsCert:      cfg.tlsCert,
		creds:        cfg.transportCredentials,
		serverName:   cfg.serverName,
		jwtGenerator: cfg.jwtGenerator,
	}

	err := wc.initGrpcConn()
	if err != nil {
		return nil, fmt.Errorf("failed to dial storage service at %s: %w", address, err)
	}

	wc.logger.Infow("Connected to Storage service (WorkflowClient)", "address", address)
	return wc, nil
}

func (wc *workflowClient) injectToken(ctx context.Context, req any) (context.Context, error) {
	if wc.jwtGenerator == nil {
		wc.logger.Warnw("authentication: no JWT generator")
		return ctx, nil
	}
	token, err := wc.jwtGenerator.CreateJWTForRequest(req)
	if err != nil {
		return ctx, fmt.Errorf("failed to generate JWT token: %w", err)
	}

	// Inject the token into the context
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token), nil
}

func (wc *workflowClient) initGrpcConn(opts ...grpc.DialOption) error {
	if wc.tlsCert != "" {
		cp := x509.NewCertPool()
		if !cp.AppendCertsFromPEM([]byte(wc.tlsCert)) {
			return fmt.Errorf("credentials: failed to append certificates")
		}
		wc.logger.Infow("Dialing with TLS (using provided certificate)", "address", wc.address)
		// Use the provided serverName variable.
		wc.creds = credentials.NewClientTLSFromCert(cp, wc.serverName)
	} else {
		wc.logger.Infow("Dialing with provided credentials", "address", wc.address)
	}

	conn, err := grpc.NewClient(
		wc.address,
		append(opts,
			grpc.WithTransportCredentials(wc.creds),
		)...,
	)
	if err != nil {
		wc.logger.Errorw("Failed to create grpc client", "error", err, "address", wc.address)

		return fmt.Errorf("failed to dial grpc client: %w", err)
	}

	wc.conn = conn
	wc.client = pb.NewNodeServiceClient(conn)

	return nil
}
