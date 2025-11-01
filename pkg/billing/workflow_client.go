package billing

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	auth "github.com/smartcontractkit/chainlink-common/pkg/nodeauth/jwt"
	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/types"
	pb "github.com/smartcontractkit/chainlink-protos/billing/go"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
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
// Returns the updated context and the JWT token string (for logging purposes)
func (wc *workflowClient) addJWTAuth(ctx context.Context, req any) (context.Context, string, error) {
	// Skip authentication if no JWT manager provided
	if wc.jwtGenerator == nil {
		return ctx, "", nil
	}

	// Create JWT token using the JWT manager
	jwtToken, err := wc.jwtGenerator.CreateJWTForRequest(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create JWT: %w", err)
	}

	// Add JWT to Authorization header
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+jwtToken), jwtToken, nil
}

// parseJWTForLogging parses the JWT token without verification to extract claims for logging purposes
func parseJWTForLogging(tokenString string) *types.NodeJWTClaims {
	if tokenString == "" {
		return nil
	}
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenString, &types.NodeJWTClaims{})
	if err != nil {
		return nil
	}

	claims, ok := token.Claims.(*types.NodeJWTClaims)
	if !ok {
		return nil
	}

	return claims
}

// DigestDebugInfo contains debugging information about digest calculation
type DigestDebugInfo struct {
	RequestType      string
	IsProtoMessage   bool
	SerializedLength int
	RequestJSON      string
	MarshalSuccess   bool
}

// calculateDigestWithDebugging calculates the request digest with detailed debugging information
func calculateDigestWithDebugging(req any) (string, DigestDebugInfo) {
	debugInfo := DigestDebugInfo{
		RequestType: fmt.Sprintf("%T", req),
	}

	var data []byte
	if m, ok := req.(proto.Message); ok {
		debugInfo.IsProtoMessage = true
		// Use protobuf canonical serialization
		serialized, err := proto.Marshal(m)
		if err == nil {
			debugInfo.MarshalSuccess = true
			data = serialized
		} else {
			debugInfo.MarshalSuccess = false
			// fallback to string representation if marshal fails
			data = fmt.Appendf(nil, "%v", req)
		}
	} else if s, ok := req.(fmt.Stringer); ok {
		debugInfo.IsProtoMessage = false
		debugInfo.MarshalSuccess = true
		data = []byte(s.String())
	} else {
		debugInfo.IsProtoMessage = false
		debugInfo.MarshalSuccess = true
		data = fmt.Appendf(nil, "%v", req)
	}

	debugInfo.SerializedLength = len(data)

	// Create JSON representation of the request for human-readable debugging
	if jsonBytes, err := json.Marshal(req); err == nil {
		debugInfo.RequestJSON = string(jsonBytes)
	} else {
		// Fallback to string representation
		debugInfo.RequestJSON = fmt.Sprintf("%+v", req)
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), debugInfo
}

func (wc *workflowClient) GetOrganizationCreditsByWorkflow(ctx context.Context, req *pb.GetOrganizationCreditsByWorkflowRequest) (*pb.GetOrganizationCreditsByWorkflowResponse, error) {
	ctx, _, err := wc.addJWTAuth(ctx, req)
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
	ctx, _, err := wc.addJWTAuth(ctx, req)
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
	ctx, _, err := wc.addJWTAuth(ctx, req)
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
	// Calculate digest and get debug info before adding JWT
	clientDigest, debugInfo := calculateDigestWithDebugging(req)

	// Add JWT authentication
	ctx, jwtToken, err := wc.addJWTAuth(ctx, req)
	if err != nil {
		wc.logger.Errorw("Failed to add JWT auth to SubmitWorkflowReceipt request", "error", err)
		return nil, err
	}

	// Parse JWT claims for logging
	parsedClaims := parseJWTForLogging(jwtToken)

	// Log detailed request information (matching billing service format)
	logFields := []any{
		"method", "SubmitWorkflowReceipt",
		"jwt_token", jwtToken,
		"client_calculated_digest", clientDigest,
		"request_type", debugInfo.RequestType,
		"is_proto_message", debugInfo.IsProtoMessage,
		"serialized_length", debugInfo.SerializedLength,
		"request_json", debugInfo.RequestJSON,
		"marshal_success", debugInfo.MarshalSuccess,
	}

	if parsedClaims != nil {
		logFields = append(logFields,
			"parsed_public_key", parsedClaims.PublicKey,
			"parsed_digest_from_jwt", parsedClaims.Digest,
			"digest_match", parsedClaims.Digest == clientDigest,
			"parsed_issuer", parsedClaims.Issuer,
			"parsed_subject", parsedClaims.Subject,
			"parsed_expires_at", parsedClaims.ExpiresAt,
			"parsed_issued_at", parsedClaims.IssuedAt,
			"parsed_audience", parsedClaims.Audience,
		)
	}

	wc.logger.Infow("Sending SubmitWorkflowReceipt request", logFields...)

	// Make the actual RPC call
	resp, err := wc.client.SubmitWorkflowReceipt(ctx, req)
	if err != nil {
		// Log error with the same detailed information for debugging
		wc.logger.Errorw("SubmitWorkflowReceipt failed", append(logFields, "error", err)...)
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
