package billing

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/mocks"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	pb "github.com/smartcontractkit/chainlink-protos/billing/go"
)

// mockRequest is a simple type that implements fmt.Stringer.
type MockRequest struct {
	Field string
}

func (d MockRequest) String() string {
	return d.Field
}


// ---------- Test Server Implementation ----------

// testWorkflowServer implements pb.WorkflowServiceServer for testing.
type testWorkflowServer struct {
	pb.UnimplementedWorkflowServiceServer
}

func (s *testWorkflowServer) GetAccountCredits(_ context.Context, _ *pb.GetAccountCreditsRequest) (*pb.GetAccountCreditsResponse, error) {
	return &pb.GetAccountCreditsResponse{
		Credits: []*pb.AccountCredits{
			{CreditType: "TEST", Credits: 100},
		},
	}, nil
}


// ---------- Test GRPC Dial with TLS Credentials ----------

func TestIntegration_GRPCWithCerts(t *testing.T) {
	// Paths to self-signed certificate and key fixtures.
	serverCertPath := "./test-fixtures/domain_test.pem"
	serverKeyPath := "./test-fixtures/domain_test.key"

	// Ensure fixture files exist.
	_, err := os.Stat(serverCertPath)
	require.NoError(t, err)
	_, err = os.Stat(serverKeyPath)
	require.NoError(t, err)

	// Create server TLS credentials.
	serverCreds, err := credentials.NewServerTLSFromFile(serverCertPath, serverKeyPath)
	require.NoError(t, err)

	// Start a test gRPC server with TLS.
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	grpcServer := grpc.NewServer(grpc.Creds(serverCreds))
	testSrv := &testWorkflowServer{}
	pb.RegisterWorkflowServiceServer(grpcServer, testSrv)
	go func() {
		_ = grpcServer.Serve(lis)
	}()
	defer grpcServer.Stop()

	// Create client TLS credentials by loading the server certificate.
	clientCreds, err := credentials.NewClientTLSFromFile(serverCertPath, "")
	require.NoError(t, err)

	certBytes, err := os.ReadFile(serverCertPath)
	require.NoError(t, err)
	require.NotEmpty(t, certBytes)

	addr := lis.Addr().String()

	// Create mock JWT manager for testing
	mockJWT := mocks.NewJWTManager(t)
	// Since we're making a real call, expect JWT creation
	mockJWT.EXPECT().CreateJWTForRequest(&pb.GetAccountCreditsRequest{AccountId: "test-account"}).Return("test.jwt.token", nil).Once()

	lggr := logger.Test(t)
	wc, err := NewWorkflowClient(addr,
		WithWorkflowTransportCredentials(clientCreds), // Provided but may be overridden by TLS cert.
		WithWorkflowTLSCert(string(certBytes)),
		WithJWTManager(mockJWT),
		WithWorkflowLogger(lggr),
		WithServerName("localhost"),
	)
	require.NoError(t, err)
	defer func(wc WorkflowClient) {
		err2 := wc.Close()
		if err2 != nil {
			t.Error(err2)
		}
	}(wc)

	// Call a method to verify that the client and server communicate over TLS.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := wc.GetAccountCredits(ctx, &pb.GetAccountCreditsRequest{AccountId: "test-account"})
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "TEST", resp.Credits[0].CreditType)
}

func TestIntegration_GRPC_Insecure(t *testing.T) {
	// Paths to self-signed certificate and key fixtures.
	serverCertPath := "./test-fixtures/domain_test.pem"
	serverKeyPath := "./test-fixtures/domain_test.key"

	_, err := os.Stat(serverCertPath)
	require.NoError(t, err)
	_, err = os.Stat(serverKeyPath)
	require.NoError(t, err)

	serverCreds, err := credentials.NewServerTLSFromFile(serverCertPath, serverKeyPath)
	require.NoError(t, err)

	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	grpcServer := grpc.NewServer(grpc.Creds(serverCreds))
	testSrv := &testWorkflowServer{}
	pb.RegisterWorkflowServiceServer(grpcServer, testSrv)
	go func() {
		_ = grpcServer.Serve(lis)
	}()
	defer grpcServer.Stop()

	addr := lis.Addr().String()
	lggr := logger.Test(t)

	wc, err := NewWorkflowClient(addr,
		WithWorkflowTransportCredentials(insecure.NewCredentials()),
		WithWorkflowLogger(lggr),
		WithServerName("localhost"),
	)

	assert.NoError(t, err)
	assert.NotNil(t, wc)

	_, err = wc.ConsumeCredits(context.Background(), nil)

	require.Error(t, err)
}

// Test that NewWorkflowClient fails when given an invalid address.
func TestNewWorkflowClient_InvalidAddress(t *testing.T) {
	lggr := logger.Test(t)
	wc, err := NewWorkflowClient("invalid-address",
		WithWorkflowTransportCredentials(insecure.NewCredentials()),
		WithWorkflowLogger(lggr),
		WithServerName("localhost"),
	)

	require.NotNil(t, wc)
	require.NoError(t, err)

	_, err = wc.ConsumeCredits(context.Background(), nil)

	require.Error(t, err, "Expected error when dialing an invalid address")
}

// Test that calling Close() twice does not cause a panic.
func TestWorkflowClient_CloseTwice(t *testing.T) {
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	grpcServer := grpc.NewServer()
	go func() {
		_ = grpcServer.Serve(lis)
	}()
	defer grpcServer.Stop()

	addr := lis.Addr().String()
	lggr := logger.Test(t)
	wc, err := NewWorkflowClient(addr,
		WithWorkflowTransportCredentials(insecure.NewCredentials()),
		WithWorkflowLogger(lggr),
		WithServerName("localhost"),
	)
	require.NoError(t, err)
	require.NotNil(t, wc)

	err = wc.Close()
	require.NoError(t, err, "First Close() should not return an error")

	err = wc.Close()
	t.Log("Second Close() call error (if any):", err)
}

// Additional test: Verify that dialGrpc fails if an unreachable address is provided.
func TestWorkflowClient_DialUnreachable(t *testing.T) {
	lggr := logger.Test(t)
	unreachableAddr := "192.0.2.1:12345" // Reserved for documentation.
	wc, err := NewWorkflowClient(unreachableAddr,
		WithWorkflowTransportCredentials(insecure.NewCredentials()),
		WithWorkflowLogger(lggr),
		WithServerName("localhost"),
	)

	require.NotNil(t, wc)
	require.NoError(t, err)

	_, err = wc.ConsumeCredits(context.Background(), nil)

	require.Error(t, err, "Expected dialing an unreachable address to fail")
}

// ---------- Test JWT Token Creation ----------

func TestWorkflowClient_AddJWTAuthToContext(t *testing.T) {

	mockJWT := mocks.NewJWTManager(t)
	req := MockRequest{Field: "test request"}
	expectedToken := "mock.jwt.token"

	// Set expectations - client should call JWT manager
	mockJWT.EXPECT().CreateJWTForRequest(req).Return(expectedToken, nil).Once()

	wc := &workflowClient{
		logger:     logger.Test(t),
		jwtManager: mockJWT,
	}

	// Test that client calls JWT manager and adds token to header
	ctx := context.Background()
	newCtx, err := wc.addJWTAuth(ctx, req)
	require.NoError(t, err)

	// Verify JWT is added to metadata
	md, ok := metadata.FromOutgoingContext(newCtx)
	require.True(t, ok, "Expected outgoing metadata to be present")

	values := md["authorization"]
	require.NotEmpty(t, values, "authorization header should be present")
	authHeader := values[0]
	require.Equal(t, "Bearer "+expectedToken, authHeader, "Authorization header should contain expected token")
}

// Test that client handles the case when no JWT manager is provided.
func TestWorkflowClient_NoSigningKey(t *testing.T) {
	ctx := context.Background()
	req := MockRequest{Field: "test"}
	wc := &workflowClient{
		logger: logger.Test(t),
		// jwtManager is nil
	}
	newCtx, err := wc.addJWTAuth(ctx, req)
	require.NoError(t, err)

	// Should return the same context since no JWT manager is provided
	assert.Equal(t, ctx, newCtx)
}

// Test that client handles JWT manager errors properly
func TestWorkflowClient_VerifySignature_Invalid(t *testing.T) {
	mockJWT := mocks.NewJWTManager(t)
	req := MockRequest{Field: "test"}

	// Set expectation for error
	mockJWT.EXPECT().CreateJWTForRequest(req).Return("", fmt.Errorf("mock JWT creation error")).Once()

	wc := &workflowClient{
		logger:     logger.Test(t),
		jwtManager: mockJWT,
	}

	ctx := context.Background()
	_, err := wc.addJWTAuth(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create JWT")
}

func TestWorkflowClient_RepeatedSign(t *testing.T) {
	mockJWT := mocks.NewJWTManager(t)
	req := MockRequest{Field: "repeatable"}
	expectedToken := "consistent.jwt.token"

	// Expect the same call twice - client should call JWT manager each time
	mockJWT.EXPECT().CreateJWTForRequest(req).Return(expectedToken, nil).Times(2)

	wc := &workflowClient{
		logger:     logger.Test(t),
		jwtManager: mockJWT,
	}

	ctx1 := context.Background()
	newCtx1, err := wc.addJWTAuth(ctx1, req)
	require.NoError(t, err)

	ctx2 := context.Background()
	newCtx2, err := wc.addJWTAuth(ctx2, req)
	require.NoError(t, err)

	// Both should have the same token since we're mocking the same response
	md1, ok := metadata.FromOutgoingContext(newCtx1)
	require.True(t, ok)
	md2, ok := metadata.FromOutgoingContext(newCtx2)
	require.True(t, ok)

	assert.Equal(t, md1["authorization"], md2["authorization"], "Expected same authorization header for same request")
}