package billing

import (
	"context"
	"crypto/x509"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/smartcontractkit/chainlink-protos/billing/go"
	auth "github.com/smartcontractkit/chainlink-common/pkg/nodeauth"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// WorkflowClient is a specialized interface for the Workflow node use-case.
type WorkflowClient interface {
	// Reads
	GetAccountCredits(ctx context.Context, req *pb.GetAccountCreditsRequest) (*pb.GetAccountCreditsResponse, error)
	BatchGetCreditsForAccounts(ctx context.Context, req *pb.BatchGetCreditsForAccountsRequest) (*pb.BatchGetCreditsForAccountsResponse, error)

	// Workflow-based credit ops
	ReserveCredits(ctx context.Context, req *pb.ReserveCreditsRequest) (*pb.ReserveCreditsResponse, error)
	ReleaseReservation(ctx context.Context, req *pb.ReleaseReservationRequest) (*pb.ReleaseReservationResponse, error)
	ConsumeCredits(ctx context.Context, req *pb.ConsumeCreditsRequest) (*pb.ConsumeCreditsResponse, error)
	ConsumeReservation(ctx context.Context, req *pb.ConsumeReservationRequest) (*pb.ConsumeReservationResponse, error)

	// Workflow receipt
	SubmitWorkflowReceipt(ctx context.Context, req *pb.SubmitWorkflowReceiptRequest) (*pb.SubmitWorkflowReceiptResponse, error)

	// Closer
	Close() error
}

type workflowClient struct {
	address string
	conn    *grpc.ClientConn
	client  pb.WorkflowServiceClient
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
	loggerInst, _ := logger.New()
	// By default, no JWT manager is set and we fallback to insecure creds.
	return workflowConfig{
		transportCredentials: insecure.NewCredentials(),
		log:                  loggerInst,
		tlsCert:              "",
		// Default to "localhost" if not overridden.
		serverName: "localhost",
		jwtGenerator: nil,
	}
}

func WithWorkflowTransportCredentials(creds credentials.TransportCredentials) WorkflowClientOpt {
	return func(cfg *workflowConfig) {
		cfg.transportCredentials = creds
	}
}

func WithWorkflowLogger(l logger.Logger) WorkflowClientOpt {
	return func(cfg *workflowConfig) {
		cfg.log = l
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
func NewWorkflowClient(address string, opts ...WorkflowClientOpt) (WorkflowClient, error) {
	cfg := defaultWorkflowConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	wc := &workflowClient{
		address:    address,
		logger:     cfg.log,
		tlsCert:    cfg.tlsCert,
		creds:      cfg.transportCredentials,
		serverName: cfg.serverName,
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
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+jwtToken), nil
}

func (wc *workflowClient) GetAccountCredits(ctx context.Context, req *pb.GetAccountCreditsRequest) (*pb.GetAccountCreditsResponse, error) {
	ctx, err := wc.addJWTAuth(ctx, req)
	if err != nil {
		wc.logger.Errorw("Failed to add JWT auth to GetAccountCredits request", "error", err)
		return nil, err
	}
	resp, err := wc.client.GetAccountCredits(ctx, req)
	if err != nil {
		wc.logger.Errorw("GetAccountCredits failed", "error", err)
		return nil, err
	}
	return resp, nil
}

func (wc *workflowClient) BatchGetCreditsForAccounts(ctx context.Context, req *pb.BatchGetCreditsForAccountsRequest) (*pb.BatchGetCreditsForAccountsResponse, error) {
	ctx, err := wc.addJWTAuth(ctx, req)
	if err != nil {
		wc.logger.Errorw("Failed to add JWT auth to BatchGetCreditsForAccounts request", "error", err)
		return nil, err
	}
	resp, err := wc.client.BatchGetCreditsForAccounts(ctx, req)
	if err != nil {
		wc.logger.Errorw("BatchGetCreditsForAccounts failed", "error", err)
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

func (wc *workflowClient) ReleaseReservation(ctx context.Context, req *pb.ReleaseReservationRequest) (*pb.ReleaseReservationResponse, error) {
	ctx, err := wc.addJWTAuth(ctx, req)
	if err != nil {
		wc.logger.Errorw("Failed to add JWT auth to ReleaseReservation request", "error", err)
		return nil, err
	}
	resp, err := wc.client.ReleaseReservation(ctx, req)
	if err != nil {
		wc.logger.Errorw("ReleaseReservation failed", "error", err)
		return nil, err
	}
	return resp, nil
}

func (wc *workflowClient) ConsumeCredits(ctx context.Context, req *pb.ConsumeCreditsRequest) (*pb.ConsumeCreditsResponse, error) {
	ctx, err := wc.addJWTAuth(ctx, req)
	if err != nil {
		wc.logger.Errorw("Failed to add JWT auth to ConsumeCredits request", "error", err)
		return nil, err
	}
	resp, err := wc.client.ConsumeCredits(ctx, req)
	if err != nil {
		wc.logger.Errorw("ConsumeCredits failed", "error", err)
		return nil, err
	}
	return resp, nil
}

func (wc *workflowClient) ConsumeReservation(ctx context.Context, req *pb.ConsumeReservationRequest) (*pb.ConsumeReservationResponse, error) {
	ctx, err := wc.addJWTAuth(ctx, req)
	if err != nil {
		wc.logger.Errorw("Failed to add JWT auth to ConsumeReservation request", "error", err)
		return nil, err
	}
	resp, err := wc.client.ConsumeReservation(ctx, req)
	if err != nil {
		wc.logger.Errorw("ConsumeReservation failed", "error", err)
		return nil, err
	}
	return resp, nil
}

func (wc *workflowClient) SubmitWorkflowReceipt(ctx context.Context, req *pb.SubmitWorkflowReceiptRequest) (*pb.SubmitWorkflowReceiptResponse, error) {
	ctx, err := wc.addJWTAuth(ctx, req)
	if err != nil {
		wc.logger.Errorw("Failed to add JWT auth to SubmitWorkflowReceipt request", "error", err)
		return nil, err
	}
	resp, err := wc.client.WorkflowReceipt(ctx, req)
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
	wc.client = pb.NewWorkflowServiceClient(conn)

	return nil
}
