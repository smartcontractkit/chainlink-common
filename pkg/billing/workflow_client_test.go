package billing

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/smartcontractkit/chainlink-common/pkg/billing/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// ---------- Test Server Implementation ----------

// testWorkflowServer implements pb.WorkflowServiceServer for testing.
type testWorkflowServer struct {
	pb.UnimplementedWorkflowServiceServer
}

func (s *testWorkflowServer) GetAccountCredits(ctx context.Context, req *pb.GetAccountCreditsRequest) (*pb.GetAccountCreditsResponse, error) {
	return &pb.GetAccountCreditsResponse{
		Credits: []*pb.AccountCredits{
			{CreditType: "TEST", Credits: 100},
		},
	}, nil
}

// ---------- Test for Sign and VerifySignature ----------

// mockRequest is a simple type that implements fmt.Stringer.
type mockRequest struct {
	Field string
}

func (d mockRequest) String() string {
	return d.Field
}

func TestWorkflowClient_SignAndVerify(t *testing.T) {
	lggr, _ := logger.New()
	// Generate an ed25519 key pair for testing.
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	req := mockRequest{Field: "test request"}

	// Create a workflowClient instance with the test signing key.
	wc := &workflowClient{
		logger:     lggr,
		signingKey: priv,
	}

	// Sign the request.
	signature, err := wc.Sign(req)
	require.NoError(t, err)
	require.NotEmpty(t, signature)

	// Verify the signature using our VerifySignature function.
	err = VerifySignature(pub, req, signature)
	require.NoError(t, err)
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

	// Generate a signing key so that the custom auth header is added.
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)
	lggr, _ := logger.New()
	wc, err := NewWorkflowClient(addr,
		WithWorkflowTransportCredentials(clientCreds), // Provided but may be overridden by TLS cert.
		WithWorkflowTLSCert(string(certBytes)),
		WithSigningKey(priv),
		WithServerPublicKey(pub),
		WithWorkflowLogger(lggr),
		WithServerName("localhost"),
	)
	require.NoError(t, err)
	defer wc.Close()

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

	lggr, _ := logger.New()
	wc, err := NewWorkflowClient(addr,
		WithWorkflowTransportCredentials(insecure.NewCredentials()),
		WithWorkflowLogger(lggr),
		WithServerName("localhost"),
	)
	require.Error(t, err)
	assert.Nil(t, wc)
}

// Test that CanonicalStringFromRequest returns the correct string.
func TestWorkflowClient_CanonicalString(t *testing.T) {
	lggr, _ := logger.New()
	wc := &workflowClient{
		logger: lggr,
	}

	dr := mockRequest{Field: "hello"}
	str := wc.CanonicalStringFromRequest(dr)
	require.Equal(t, "hello", str)

	x := 42
	str2 := wc.CanonicalStringFromRequest(x)
	require.Equal(t, "42", str2)

	type sample struct {
		A int
		B string
	}
	s := sample{A: 10, B: "foo"}
	expected := fmt.Sprintf("%v", s)
	require.Equal(t, expected, wc.CanonicalStringFromRequest(s))
}

// Test that VerifySignature fails when the request is altered.
func TestWorkflowClient_VerifySignature_Invalid(t *testing.T) {
	lggr, _ := logger.New()
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	req := mockRequest{Field: "original"}
	wc := &workflowClient{
		logger:     lggr,
		signingKey: priv,
	}
	sig, err := wc.Sign(req)
	require.NoError(t, err)

	reqAltered := mockRequest{Field: "tampered"}
	err = VerifySignature(pub, reqAltered, sig)
	require.Error(t, err, "Expected signature verification to fail for altered request")
}

// Updated test: Verify that addSignature does nothing when no signing key is provided.
func TestWorkflowClient_NoSigningKey(t *testing.T) {
	ctx := context.Background()
	req := mockRequest{Field: "test"}
	wc := &workflowClient{
		// signingKey is nil
	}
	newCtx, err := wc.addSignature(ctx, req)
	require.NoError(t, err)

	md, ok := metadata.FromOutgoingContext(newCtx)
	if ok {
		_, exists := md["x-custom-auth"]
		require.False(t, exists, "Expected no 'x-custom-auth' metadata when no signing key is set")
	} else {
		assert.True(t, true, "No outgoing metadata, as expected")
	}
}

func TestWorkflowClient_CustomAuthHeader(t *testing.T) {
	lggr, _ := logger.New()
	_, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	wc := &workflowClient{
		logger:     lggr,
		signingKey: priv,
	}

	req := mockRequest{Field: "custom auth test"}
	ctx := context.Background()
	newCtx, err := wc.addSignature(ctx, req)
	require.NoError(t, err)

	md, ok := metadata.FromOutgoingContext(newCtx)
	require.True(t, ok, "Expected outgoing metadata to be present")

	values := md["x-custom-auth"]
	require.NotEmpty(t, values, "x-custom-auth header should be present")
	authHeader := values[0]
	parts := strings.Split(authHeader, "|")
	require.Len(t, parts, 3, "x-custom-auth header should contain three parts separated by '|'")

	// First part: public key.
	expectedPubHex := hex.EncodeToString(priv.Public().(ed25519.PublicKey))
	require.Equal(t, expectedPubHex, parts[0], "Public key in header should match")

	// Second part: timestamp.
	ts, err := time.Parse(time.RFC3339, parts[1])
	require.NoError(t, err, "Timestamp should be in RFC3339 format")
	require.Less(t, time.Since(ts).Seconds(), 5.0, "Timestamp should be recent")

	// Third part: SHA256 hash (64 hex characters).
	require.Len(t, parts[2], 64, "Hash in x-custom-auth header should be 64 hex characters")
}

// Test that NewWorkflowClient fails when given an invalid address.
func TestNewWorkflowClient_InvalidAddress(t *testing.T) {
	lggr, _ := logger.New()
	_, err := NewWorkflowClient("invalid-address",
		WithWorkflowTransportCredentials(insecure.NewCredentials()),
		WithWorkflowLogger(lggr),
		WithServerName("localhost"),
	)
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
	lggr, _ := logger.New()
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

// Additional test: Verify that signing produces a valid signature and repeated signing yields the same result.
func TestWorkflowClient_RepeatedSign(t *testing.T) {
	lggr, _ := logger.New()
	pub, priv, err := ed25519.GenerateKey(nil)
	require.NoError(t, err)

	wc := &workflowClient{
		logger:     lggr,
		signingKey: priv,
	}
	req := mockRequest{Field: "repeatable"}
	sig1, err := wc.Sign(req)
	require.NoError(t, err)
	require.NotEmpty(t, sig1)

	sig2, err := wc.Sign(req)
	require.NoError(t, err)
	require.Equal(t, sig1, sig2, "Expected repeated signatures for the same request to match")

	err = VerifySignature(pub, req, sig1)
	require.NoError(t, err)
}

// Additional test: Verify that dialGrpc fails if an unreachable address is provided.
func TestWorkflowClient_DialUnreachable(t *testing.T) {
	lggr, _ := logger.New()
	unreachableAddr := "192.0.2.1:12345" // Reserved for documentation.
	_, err := NewWorkflowClient(unreachableAddr,
		WithWorkflowTransportCredentials(insecure.NewCredentials()),
		WithWorkflowLogger(lggr),
		WithServerName("localhost"),
	)
	require.Error(t, err, "Expected dialing an unreachable address to fail")
}
