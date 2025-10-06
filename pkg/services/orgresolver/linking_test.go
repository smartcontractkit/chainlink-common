package orgresolver

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	defer conn.Close()

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
