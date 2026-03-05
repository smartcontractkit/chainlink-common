package billing

import (
	"context"
	"crypto/x509"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	nodeauthgrpc "github.com/smartcontractkit/chainlink-common/pkg/nodeauth/grpc"
	auth "github.com/smartcontractkit/chainlink-common/pkg/nodeauth/jwt"
	pb "github.com/smartcontractkit/chainlink-protos/billing/go"
)

// WorkflowClient is a specialized interface for the Workflow node use-case.
type WorkflowClient interface {
	GetOrganizationCreditsByWorkflow(ctx context.Context, req *pb.GetOrganizationCreditsByWorkflowRequest) (*pb.GetOrganizationCreditsByWorkflowResponse, error)
	GetWorkflowExecutionRates(ctx context.Context, req *pb.GetWorkflowExecutionRatesRequest) (*pb.GetWorkflowExecutionRatesResponse, error)
	ReserveCredits(ctx context.Context, req *pb.ReserveCreditsRequest) (*pb.ReserveCreditsResponse, error)
	SubmitWorkflowReceipt(ctx context.Context, req *pb.SubmitWorkflowReceiptRequest) (*emptypb.Empty, error)

	// Closer
	Close() error
}

type workflowClient struct {
	address string
	conn    *grpc.ClientConn
	client  pb.CreditReservationServiceClient
	logger  logger.Logger
	tlsCert string
	creds   credentials.TransportCredentials
	// serverName is the expected server name in the TLS certificate.
	serverName string

	// JWT-based authentication
	jwtGenerator auth.JWTGenerator
}

type workflowConfig struct {
	log                  logger.Logger
	transportCredentials credentials.TransportCredentials
	tlsCert              string
	serverName           string
	jwtGenerator         auth.JWTGenerator
}

type WorkflowClientOpt func(*workflowConfig)

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
func WithJWTGenerator(jwtGenerator auth.JWTGenerator) WorkflowClientOpt {
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

// NewWorkflowClient creates a new workflow client with JWT signing enabled.
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
		return nil, fmt.Errorf("failed to dial billing service at %s: %w", address, err)
	}

	wc.logger.Infow("Connected to Billing service (WorkflowClient)", "address", address)
	return wc, nil
}

// Close closes the gRPC connection used by the workflow client.
func (wc *workflowClient) Close() error {
	err := wc.conn.Close()
	if err != nil {
		wc.logger.Errorw("Failed to close WorkflowClient connection", "error", err)
		return err
	}
	wc.logger.Infow("Closed WorkflowClient connection")
	return nil
}

// addJWTAuth creates and signs a JWT token, then adds it to the context
func (wc *workflowClient) addJWTAuth(ctx context.Context, req any) (context.Context, error) {
	// Skip authentication if no JWT manager provided
	if wc.jwtGenerator == nil {
		return ctx, nil
	}

	// Create JWT token using the JWT manager
	jwtToken, err := wc.jwtGenerator.CreateJWTForRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT: %w", err)
	}

	// Add JWT to Authorization header
	return metadata.AppendToOutgoingContext(ctx, nodeauthgrpc.AuthorizationHeader, nodeauthgrpc.BearerPrefix+jwtToken), nil
}

func (wc *workflowClient) GetOrganizationCreditsByWorkflow(ctx context.Context, req *pb.GetOrganizationCreditsByWorkflowRequest) (*pb.GetOrganizationCreditsByWorkflowResponse, error) {
	ctx, err := wc.addJWTAuth(ctx, req)
	if err != nil {
		wc.logger.Errorw("Failed to add custom auth header to GetOrganizationCreditsByWorkflow request", "error", err)
		return nil, err
	}
	resp, err := wc.client.GetOrganizationCreditsByWorkflow(ctx, req)
	if err != nil {
		wc.logger.Errorw("GetOrganizationCreditsByWorkflow failed", "error", err)
		return nil, err
	}
	return resp, nil
}

func (wc *workflowClient) GetWorkflowExecutionRates(ctx context.Context, req *pb.GetWorkflowExecutionRatesRequest) (*pb.GetWorkflowExecutionRatesResponse, error) {
	ctx, err := wc.addJWTAuth(ctx, req)
	if err != nil {
		wc.logger.Errorw("Failed to add custom auth header to GetWorkflowExecutionRates request", "error", err)
		return nil, err
	}
	resp, err := wc.client.GetWorkflowExecutionRates(ctx, req)
	if err != nil {
		wc.logger.Errorw("GetWorkflowExecutionRates failed", "error", err)
		return nil, err
	}
	return resp, nil
}

func (wc *workflowClient) ReserveCredits(ctx context.Context, req *pb.ReserveCreditsRequest) (*pb.ReserveCreditsResponse, error) {
	ctx, err := wc.addJWTAuth(ctx, req)
	if err != nil {
		wc.logger.Errorw("Failed to add JWT auth to ReserveCredits request", "error", err)
		return nil, err
	}
	resp, err := wc.client.ReserveCredits(ctx, req)
	if err != nil {
		wc.logger.Errorw("ReserveCredits failed", "error", err)
		return nil, err
	}
	return resp, nil
}

func (wc *workflowClient) SubmitWorkflowReceipt(ctx context.Context, req *pb.SubmitWorkflowReceiptRequest) (*emptypb.Empty, error) {
	ctx, err := wc.addJWTAuth(ctx, req)
	if err != nil {
		wc.logger.Errorw("Failed to add JWT auth to SubmitWorkflowReceipt request", "error", err)
		return nil, err
	}
	resp, err := wc.client.SubmitWorkflowReceipt(ctx, req)
	if err != nil {
		wc.logger.Errorw("SubmitWorkflowReceipt failed", "error", err)
		return nil, err
	}
	return resp, nil
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
	wc.client = pb.NewCreditReservationServiceClient(conn)

	return nil
}
