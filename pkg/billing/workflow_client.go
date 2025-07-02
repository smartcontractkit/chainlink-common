package billing

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/smartcontractkit/chainlink-protos/billing/go"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// WorkflowClient is a specialized interface for the Workflow node use-case.
type WorkflowClient interface {
	GetOrganizationCreditsByWorkflow(ctx context.Context, req *pb.GetOrganizationCreditsByWorkflowRequest) (*pb.GetOrganizationCreditsByWorkflowResponse, error)
	GetRateCard(ctx context.Context, req *pb.GetRateCardRequest) (*pb.GetRateCardResponse, error)
	ReserveCredits(ctx context.Context, req *pb.ReserveCreditsRequest) (*pb.ReserveCreditsResponse, error)
	SubmitWorkflowReceipt(ctx context.Context, req *pb.SubmitWorkflowReceiptRequest) (*emptypb.Empty, error)

	// Closer
	Close() error
}

type workflowClient struct {
	address    string
	conn       *grpc.ClientConn
	client     pb.CreditReservationServiceClient
	logger     logger.Logger
	signingKey ed25519.PrivateKey
	// serverPubKey is used for verifying signatures from the server.
	serverPubKey ed25519.PublicKey
	tlsCert      string
	creds        credentials.TransportCredentials
	// serverName is the expected server name in the TLS certificate.
	serverName string
}

type workflowConfig struct {
	log                  logger.Logger
	transportCredentials credentials.TransportCredentials
	signingKey           ed25519.PrivateKey
	serverPubKey         ed25519.PublicKey
	tlsCert              string
	serverName           string
}

type WorkflowClientOpt func(*workflowConfig)

func defaultWorkflowConfig() workflowConfig {
	loggerInst, _ := logger.New()
	// By default, no signing key is set and we fallback to insecure creds.
	return workflowConfig{
		transportCredentials: insecure.NewCredentials(),
		log:                  loggerInst,
		tlsCert:              "",
		// Default to "localhost" if not overridden.
		serverName: "localhost",
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

func WithSigningKey(signingKey ed25519.PrivateKey) WorkflowClientOpt {
	return func(cfg *workflowConfig) {
		cfg.signingKey = signingKey
	}
}

func WithServerPublicKey(pub ed25519.PublicKey) WorkflowClientOpt {
	return func(cfg *workflowConfig) {
		cfg.serverPubKey = pub
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

func NewWorkflowClient(address string, opts ...WorkflowClientOpt) (WorkflowClient, error) {
	cfg := defaultWorkflowConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	wc := &workflowClient{
		address:      address,
		logger:       cfg.log,
		signingKey:   cfg.signingKey,
		serverPubKey: cfg.serverPubKey,
		tlsCert:      cfg.tlsCert,
		creds:        cfg.transportCredentials,
		serverName:   cfg.serverName,
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

func (wc *workflowClient) addSignature(ctx context.Context, req interface{}) (context.Context, error) {
	if wc.signingKey == nil {
		return ctx, nil
	}
	canonical := wc.CanonicalStringFromRequest(req)
	ts := time.Now().UTC().Format(time.RFC3339)
	hasher := sha256.New()
	hasher.Write([]byte(canonical + ts))
	digest := hex.EncodeToString(hasher.Sum(nil))

	// Use "|" as the delimiter.
	pubKeyHex := hex.EncodeToString(wc.signingKey.Public().(ed25519.PublicKey))
	headerValue := fmt.Sprintf("%s|%s|%s", pubKeyHex, ts, digest)
	return metadata.AppendToOutgoingContext(ctx, "x-custom-auth", headerValue), nil
}

func (wc *workflowClient) Sign(request interface{}) (string, error) {
	// Retained for compatibility; not used in custom auth header.
	canonical := wc.CanonicalStringFromRequest(request)
	signature := ed25519.Sign(wc.signingKey, []byte(canonical))
	return hex.EncodeToString(signature), nil
}

func (wc *workflowClient) CanonicalStringFromRequest(request interface{}) string {
	if s, ok := request.(fmt.Stringer); ok {
		return s.String()
	}
	return fmt.Sprintf("%v", request)
}

// Used on server, provided for testing.
func VerifySignature(publicKey ed25519.PublicKey, request interface{}, signature string) error {
	var canonical string
	if s, ok := request.(fmt.Stringer); ok {
		canonical = s.String()
	} else {
		canonical = fmt.Sprintf("%v", request)
	}
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return err
	}
	if !ed25519.Verify(publicKey, []byte(canonical), sigBytes) {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}

func (wc *workflowClient) GetOrganizationCreditsByWorkflow(ctx context.Context, req *pb.GetOrganizationCreditsByWorkflowRequest) (*pb.GetOrganizationCreditsByWorkflowResponse, error) {
	ctx, err := wc.addSignature(ctx, req)
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

func (wc *workflowClient) GetRateCard(ctx context.Context, req *pb.GetRateCardRequest) (*pb.GetRateCardResponse, error) {
	ctx, err := wc.addSignature(ctx, req)
	if err != nil {
		wc.logger.Errorw("Failed to add custom auth header to GetRateCard request", "error", err)
		return nil, err
	}
	resp, err := wc.client.GetRateCard(ctx, req)
	if err != nil {
		wc.logger.Errorw("GetRateCard failed", "error", err)
		return nil, err
	}
	return resp, nil
}

func (wc *workflowClient) ReserveCredits(ctx context.Context, req *pb.ReserveCreditsRequest) (*pb.ReserveCreditsResponse, error) {
	ctx, err := wc.addSignature(ctx, req)
	if err != nil {
		wc.logger.Errorw("Failed to add custom auth header to ReserveCredits request", "error", err)
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
	ctx, err := wc.addSignature(ctx, req)
	if err != nil {
		wc.logger.Errorw("Failed to add custom auth header to SubmitWorkflowReceipt request", "error", err)
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
