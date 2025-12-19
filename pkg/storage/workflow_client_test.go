package storage

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/nodeauth/jwt/mocks"
	pb "github.com/smartcontractkit/chainlink-protos/storage-service/go"
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
	pb.UnsafeNodeServiceServer
}

func (s *testWorkflowServer) DownloadArtifact(ctx context.Context, req *pb.DownloadArtifactRequest) (*pb.DownloadArtifactResponse, error) {

	resp := &pb.DownloadArtifactResponse{
		Url:    "pre-signed-url",
		Expiry: timestamppb.New(time.Now().Add(time.Hour)),
	}
	return resp, nil
}

// ---------- Test GRPC Dial with TLS Credentials ----------

func TestIntegration_GRPCWithCerts(t *testing.T) {
	t.Parallel()
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
	pb.RegisterNodeServiceServer(grpcServer, testSrv)
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
	mockJWT := mocks.NewJWTGenerator(t)
	// Since we're making a real call, expect JWT creation
	mockJWT.EXPECT().CreateJWTForRequest(&pb.DownloadArtifactRequest{
		Id:   "test",
		Type: pb.ArtifactType_ARTIFACT_TYPE_BINARY,
	}).Return("test.jwt.token", nil).Once()

	lggr := logger.Test(t)
	wc, err := NewWorkflowClient(lggr, addr,
		WithWorkflowTransportCredentials(clientCreds),
		WithWorkflowTLSCert(string(certBytes)),
		WithJWTGenerator(mockJWT),
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
	resp, err := wc.DownloadArtifact(ctx, &pb.DownloadArtifactRequest{
		Id:   "test",
		Type: pb.ArtifactType_ARTIFACT_TYPE_BINARY,
	})
	require.NoError(t, err, "DownloadArtifact should succeed over TLS")
	assert.Equal(t, "pre-signed-url", resp.Url)
}

func TestIntegration_GRPC_Insecure(t *testing.T) {
	t.Parallel()
	// Start a test gRPC server without TLS.
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	grpcServer := grpc.NewServer()
	testSrv := &testWorkflowServer{}
	pb.RegisterNodeServiceServer(grpcServer, testSrv)
	go func() {
		_ = grpcServer.Serve(lis)
	}()
	defer grpcServer.Stop()

	addr := lis.Addr().String()

	// Create mock JWT manager for testing
	mockJWT := mocks.NewJWTGenerator(t)
	// Since we're making a real call, expect JWT creation
	mockJWT.EXPECT().CreateJWTForRequest(&pb.DownloadArtifactRequest{
		Id:   "test",
		Type: pb.ArtifactType_ARTIFACT_TYPE_BINARY,
	}).Return("test.jwt.token", nil).Once()

	lggr := logger.Test(t)
	wc, err := NewWorkflowClient(lggr, addr,
		WithWorkflowTransportCredentials(insecure.NewCredentials()),
		WithJWTGenerator(mockJWT),
		WithServerName("localhost"),
	)

	require.NoError(t, err)
	defer func(wc WorkflowClient) {
		err2 := wc.Close()
		if err2 != nil {
			t.Error(err2)
		}
	}(wc)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = wc.DownloadArtifact(ctx, &pb.DownloadArtifactRequest{
		Id:   "test",
		Type: pb.ArtifactType_ARTIFACT_TYPE_BINARY,
	})
	require.NoError(t, err, "DownloadArtifact should succeed over insecure connection")
}

func TestNewWorkflowClient_InvalidAddress(t *testing.T) {
	t.Parallel()
	lggr := logger.Test(t)
	wc, err := NewWorkflowClient(lggr, "invalid-address",
		WithWorkflowTransportCredentials(insecure.NewCredentials()),
		WithServerName("localhost"),
	)

	require.NotNil(t, wc)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = wc.DownloadArtifact(ctx, &pb.DownloadArtifactRequest{
		Id:   "test",
		Type: pb.ArtifactType_ARTIFACT_TYPE_BINARY,
	})

	require.Error(t, err, "Expected error when dialing an invalid address")
}

func TestWorkflowClient_CloseTwice(t *testing.T) {
	t.Parallel()
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	grpcServer := grpc.NewServer()
	go func() {
		_ = grpcServer.Serve(lis)
	}()
	defer grpcServer.Stop()

	addr := lis.Addr().String()
	lggr := logger.Test(t)
	wc, err := NewWorkflowClient(lggr, addr,
		WithWorkflowTransportCredentials(insecure.NewCredentials()),
		WithServerName("localhost"),
	)
	require.NoError(t, err)
	require.NotNil(t, wc)

	err = wc.Close()
	require.NoError(t, err, "First Close() should not return an error")

	err = wc.Close()
	t.Log("Second Close() call error (if any):", err)
}

func TestWorkflowClient_DialUnreachable(t *testing.T) {
	t.Parallel()
	lggr := logger.Test(t)
	unreachableAddr := "192.0.2.1:12345" // Reserved for documentation.
	wc, err := NewWorkflowClient(lggr, unreachableAddr,
		WithWorkflowTransportCredentials(insecure.NewCredentials()),
		WithServerName("localhost"),
	)

	require.NotNil(t, wc)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = wc.DownloadArtifact(ctx, &pb.DownloadArtifactRequest{
		Id:   "test",
		Type: pb.ArtifactType_ARTIFACT_TYPE_BINARY,
	})

	require.Error(t, err, "Expected dialing an unreachable address to fail")
}

// ---------- Test JWT Token Creation ----------

func TestWorkflowClient_AddJWTAuthToContext(t *testing.T) {
	t.Parallel()

	mockJWT := mocks.NewJWTGenerator(t)
	req := MockRequest{Field: "test request"}
	expectedToken := "mock.jwt.token"

	mockJWT.EXPECT().CreateJWTForRequest(req).Return(expectedToken, nil).Once()

	wc := &workflowClient{
		logger:       logger.Test(t),
		jwtGenerator: mockJWT,
	}

	ctx := context.Background()
	newCtx, err := wc.injectToken(ctx, req)
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
	t.Parallel()
	ctx := context.Background()
	req := MockRequest{Field: "test"}
	wc := &workflowClient{
		logger:       logger.Test(t),
		jwtGenerator: nil,
	}
	newCtx, err := wc.injectToken(ctx, req)
	require.NoError(t, err)

	// Should return the same context
	assert.Equal(t, ctx, newCtx)
}

// Test that client handles JWT manager errors properly
func TestWorkflowClient_VerifySignature_Invalid(t *testing.T) {
	t.Parallel()
	mockJWT := mocks.NewJWTGenerator(t)
	req := MockRequest{Field: "test"}

	mockJWT.EXPECT().CreateJWTForRequest(req).Return("", fmt.Errorf("mock JWT creation error")).Once()

	wc := &workflowClient{
		logger:       logger.Test(t),
		jwtGenerator: mockJWT,
	}

	ctx := context.Background()
	_, err := wc.injectToken(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate JWT token")
}

func TestWorkflowClient_RepeatedSign(t *testing.T) {
	t.Parallel()
	mockJWT := mocks.NewJWTGenerator(t)
	req := MockRequest{Field: "repeatable"}
	expectedToken := "consistent.jwt.token"

	// Expect the same call twice
	mockJWT.EXPECT().CreateJWTForRequest(req).Return(expectedToken, nil).Times(2)

	wc := &workflowClient{
		logger:       logger.Test(t),
		jwtGenerator: mockJWT,
	}

	ctx1 := context.Background()
	newCtx1, err := wc.injectToken(ctx1, req)
	require.NoError(t, err)

	ctx2 := context.Background()
	newCtx2, err := wc.injectToken(ctx2, req)
	require.NoError(t, err)

	// Both should have the same token since we're mocking the same response
	md1, ok := metadata.FromOutgoingContext(newCtx1)
	require.True(t, ok)
	md2, ok := metadata.FromOutgoingContext(newCtx2)
	require.True(t, ok)

	assert.Equal(t, md1["authorization"], md2["authorization"], "Expected same authorization header for same request")
}
