package orgresolver

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	linkingclient "github.com/smartcontractkit/chainlink-protos/linking-service/go/v1"
)

// mockLinkingClient implements the LinkingServiceClient interface for testing
type mockLinkingClient struct{}

func (m *mockLinkingClient) GetOrganizationFromWorkflowOwner(ctx context.Context, req *linkingclient.GetOrganizationFromWorkflowOwnerRequest, opts ...grpc.CallOption) (*linkingclient.GetOrganizationFromWorkflowOwnerResponse, error) {
	orgID := "org-" + req.WorkflowOwner
	return &linkingclient.GetOrganizationFromWorkflowOwnerResponse{
		OrganizationId: orgID,
	}, nil
}

// mockJWTGenerator implements the JWTGenerator interface for testing
type mockJWTGenerator struct {
	token string
	err   error
}

func (m *mockJWTGenerator) CreateJWTForRequest(req any) (string, error) {
	return m.token, m.err
}

// mockLinkingClientWithAuthCheck implements the LinkingServiceClient interface and checks for authorization header
type mockLinkingClientWithAuthCheck struct {
	expectedAuthHeader string
	receivedAuthHeader string
}

func (m *mockLinkingClientWithAuthCheck) GetOrganizationFromWorkflowOwner(ctx context.Context, req *linkingclient.GetOrganizationFromWorkflowOwnerRequest, opts ...grpc.CallOption) (*linkingclient.GetOrganizationFromWorkflowOwnerResponse, error) {
	// Extract authorization header from context
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		if authHeaders := md.Get("authorization"); len(authHeaders) > 0 {
			m.receivedAuthHeader = authHeaders[0]
		}
	}

	orgID := "org-" + req.WorkflowOwner
	return &linkingclient.GetOrganizationFromWorkflowOwnerResponse{
		OrganizationId: orgID,
	}, nil
}

// mockLinkingServer implements the LinkingServiceServer interface for testing
type mockLinkingServer struct {
	linkingclient.UnimplementedLinkingServiceServer
}

func (s *mockLinkingServer) GetOrganizationFromWorkflowOwner(ctx context.Context, req *linkingclient.GetOrganizationFromWorkflowOwnerRequest) (*linkingclient.GetOrganizationFromWorkflowOwnerResponse, error) {
	orgID := "org-" + req.WorkflowOwner
	return &linkingclient.GetOrganizationFromWorkflowOwnerResponse{
		OrganizationId: orgID,
	}, nil
}

func TestOrgResolver_Get(t *testing.T) {
	ctx := context.Background()
	client := &mockLinkingClient{}

	cfg := Config{
		URL:                           "test-url",
		TLSEnabled:                    false,
		WorkflowRegistryAddress:       "0x1234567890abcdef",
		WorkflowRegistryChainSelector: 1,
	}

	resolver, err := NewOrgResolverWithClient(cfg, client, logger.Test(t))
	require.NoError(t, err)

	workflowOwner := "0xabcdef1234567890"

	orgID, err := resolver.Get(ctx, workflowOwner)
	require.NoError(t, err)
	require.Equal(t, "org-"+workflowOwner, orgID)
}

func TestOrgResolver_NewOrgResolver_RequiresClientOrURL(t *testing.T) {
	cfg := Config{
		URL:                           "", // Empty URL should cause error
		TLSEnabled:                    false,
		WorkflowRegistryAddress:       "0x1234567890abcdef",
		WorkflowRegistryChainSelector: 1,
	}

	_, err := NewOrgResolverWithClient(cfg, nil, logger.Test(t))
	require.Error(t, err)
	require.Contains(t, err.Error(), "URL is required when client is not provided")
}

func TestOrgResolver_NewOrgResolver_WithMockServer(t *testing.T) {
	// Use in-memory connection for faster testing
	lis := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	linkingclient.RegisterLinkingServiceServer(server, &mockLinkingServer{})

	go func() {
		_ = server.Serve(lis)
	}()
	t.Cleanup(func() { server.Stop() })

	// Create gRPC client connection using bufconn
	ctx := context.Background()
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()

	client := linkingclient.NewLinkingServiceClient(conn)

	// Create OrgResolver using the client (simulating what NewOrgResolver would do)
	cfg := Config{
		URL:                           "bufnet", // Not used since client is injected
		TLSEnabled:                    false,
		WorkflowRegistryAddress:       "0x1234567890abcdef",
		WorkflowRegistryChainSelector: 1,
	}

	resolver, err := NewOrgResolverWithClient(cfg, client, logger.Test(t))
	require.NoError(t, err)

	workflowOwner := "0xabcdef1234567890"

	orgID, err := resolver.Get(ctx, workflowOwner)
	require.NoError(t, err)
	require.Equal(t, "org-"+workflowOwner, orgID)

	err = resolver.Close()
	require.NoError(t, err)
}

func TestOrgResolver_Get_WithJWTGenerator(t *testing.T) {
	ctx := context.Background()
	client := &mockLinkingClientWithAuthCheck{}

	// Test with JWT generator that returns a valid token
	jwtGenerator := &mockJWTGenerator{
		token: "test-jwt-token-123",
		err:   nil,
	}

	cfg := Config{
		URL:                           "test-url",
		TLSEnabled:                    false,
		WorkflowRegistryAddress:       "0x1234567890abcdef",
		WorkflowRegistryChainSelector: 1,
		JWTGenerator:                  jwtGenerator,
	}

	resolver, err := NewOrgResolverWithClient(cfg, client, logger.Test(t))
	require.NoError(t, err)

	workflowOwner := "0xabcdef1234567890"

	orgID, err := resolver.Get(ctx, workflowOwner)
	require.NoError(t, err)
	require.Equal(t, "org-"+workflowOwner, orgID)

	// Verify that the authorization header was set correctly
	require.Equal(t, "Bearer test-jwt-token-123", client.receivedAuthHeader)
}

func TestOrgResolver_Get_WithJWTGeneratorError(t *testing.T) {
	ctx := context.Background()
	client := &mockLinkingClient{}

	// Test with JWT generator that returns an error
	jwtGenerator := &mockJWTGenerator{
		token: "",
		err:   errors.New("JWT generation failed"),
	}

	cfg := Config{
		URL:                           "test-url",
		TLSEnabled:                    false,
		WorkflowRegistryAddress:       "0x1234567890abcdef",
		WorkflowRegistryChainSelector: 1,
		JWTGenerator:                  jwtGenerator,
	}

	resolver, err := NewOrgResolverWithClient(cfg, client, logger.Test(t))
	require.NoError(t, err)

	workflowOwner := "0xabcdef1234567890"

	// The Get call should fail due to JWT generation error
	_, err = resolver.Get(ctx, workflowOwner)
	require.Error(t, err)
	require.Contains(t, err.Error(), "JWT generation failed")
}

func TestOrgResolver_Get_WithoutJWTGenerator(t *testing.T) {
	ctx := context.Background()
	client := &mockLinkingClientWithAuthCheck{}

	// Test without JWT generator (should not set authorization header)
	cfg := Config{
		URL:                           "test-url",
		TLSEnabled:                    false,
		WorkflowRegistryAddress:       "0x1234567890abcdef",
		WorkflowRegistryChainSelector: 1,
		JWTGenerator:                  nil, // No JWT generator
	}

	resolver, err := NewOrgResolverWithClient(cfg, client, logger.Test(t))
	require.NoError(t, err)

	workflowOwner := "0xabcdef1234567890"

	orgID, err := resolver.Get(ctx, workflowOwner)
	require.NoError(t, err)
	require.Equal(t, "org-"+workflowOwner, orgID)

	// Verify that no authorization header was set
	require.Empty(t, client.receivedAuthHeader)
}
