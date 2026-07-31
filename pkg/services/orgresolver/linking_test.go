package orgresolver

import (
	"context"
	"database/sql"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

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

// mockHangingLinkingClient blocks until the call context is done, simulating a blackholed connection
type mockHangingLinkingClient struct{}

func (m *mockHangingLinkingClient) GetOrganizationFromWorkflowOwner(ctx context.Context, req *linkingclient.GetOrganizationFromWorkflowOwnerRequest, opts ...grpc.CallOption) (*linkingclient.GetOrganizationFromWorkflowOwnerResponse, error) {
	<-ctx.Done()
	return nil, ctx.Err()
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

func TestOrgResolver_Get_TimesOutWhenServiceHangs(t *testing.T) {
	cfg := Config{
		WorkflowRegistryAddress:       "0x1234567890abcdef",
		WorkflowRegistryChainSelector: 1,
		RequestTimeout:                50 * time.Millisecond,
		Client:                        &mockHangingLinkingClient{},
	}

	resolver, err := cfg.New(logger.Test(t))
	require.NoError(t, err)

	start := time.Now()
	_, err = resolver.Get(t.Context(), "0xabcdef1234567890")
	require.ErrorIs(t, err, context.DeadlineExceeded)
	require.Less(t, time.Since(start), 5*time.Second)
}

func TestOrgResolver_Get_DefaultTimeoutApplied(t *testing.T) {
	cfg := Config{
		WorkflowRegistryAddress:       "0x1234567890abcdef",
		WorkflowRegistryChainSelector: 1,
		Client:                        &mockLinkingClient{},
	}

	resolver, err := cfg.New(logger.Test(t))
	require.NoError(t, err)
	require.Equal(t, defaultRequestTimeout, resolver.requestTimeout)

	orgID, err := resolver.Get(t.Context(), "0xabcdef1234567890")
	require.NoError(t, err)
	require.Equal(t, "org-0xabcdef1234567890", orgID)
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

// mockCacheStore implements the CacheStore interface for testing.
type mockCacheStore struct {
	data          map[string]string
	getErr        error // error to return from GetOrg (overrides data lookup)
	upsertErr     error // error to return from UpsertOrg
	getCalls      int
	upsertCalls   int
	lastUpsertKey string
	lastUpsertVal string
	mu            sync.Mutex
}

func newMockCacheStore() *mockCacheStore {
	return &mockCacheStore{data: make(map[string]string)}
}

func (m *mockCacheStore) GetOrg(_ context.Context, owner string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getCalls++
	if m.getErr != nil {
		return "", m.getErr
	}
	orgID, ok := m.data[owner]
	if !ok {
		return "", ErrCacheMiss
	}
	return orgID, nil
}

func (m *mockCacheStore) UpsertOrg(_ context.Context, owner, orgID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.upsertCalls++
	m.lastUpsertKey = owner
	m.lastUpsertVal = orgID
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.data[owner] = orgID
	return nil
}

func TestOrgResolver_Cache_HitSkipsLinkingService(t *testing.T) {
	ctx := context.Background()
	client := &mockLinkingClient{}
	cache := newMockCacheStore()

	// Pre-populate the cache with a known mapping.
	cachedOrgID := "cached-org-123"
	workflowOwner := "0xabcdef1234567890"
	cache.data[workflowOwner] = cachedOrgID

	cfg := Config{
		URL:                           "test-url",
		WorkflowRegistryAddress:       "0x1234567890abcdef",
		WorkflowRegistryChainSelector: 1,
		CacheEnabled:                  true,
		CacheStore:                    cache,
		Client:                        client,
	}

	resolver, err := cfg.New(logger.Test(t))
	require.NoError(t, err)

	orgID, err := resolver.Get(ctx, workflowOwner)
	require.NoError(t, err)
	require.Equal(t, cachedOrgID, orgID)

	// The linking service client must not be invoked on a cache hit; verify
	// indirectly by checking that no upsert happened (which only occurs after
	// a successful remote fetch).
	require.Equal(t, 1, cache.getCalls)
	require.Equal(t, 0, cache.upsertCalls)
}

func TestOrgResolver_Cache_MissFallsBackToLinkingServiceAndStores(t *testing.T) {
	ctx := context.Background()
	client := &mockLinkingClient{}
	cache := newMockCacheStore()

	workflowOwner := "0xabcdef1234567890"

	cfg := Config{
		URL:                           "test-url",
		WorkflowRegistryAddress:       "0x1234567890abcdef",
		WorkflowRegistryChainSelector: 1,
		CacheEnabled:                  true,
		CacheStore:                    cache,
		Client:                        client,
	}

	resolver, err := cfg.New(logger.Test(t))
	require.NoError(t, err)

	orgID, err := resolver.Get(ctx, workflowOwner)
	require.NoError(t, err)
	require.Equal(t, "org-"+workflowOwner, orgID)

	// Cache miss should trigger a linking-service fetch, and the result should
	// be persisted to the cache store.
	require.Equal(t, 1, cache.getCalls)
	require.Equal(t, 1, cache.upsertCalls)
	require.Equal(t, workflowOwner, cache.lastUpsertKey)
	require.Equal(t, "org-"+workflowOwner, cache.lastUpsertVal)
	require.Equal(t, "org-"+workflowOwner, cache.data[workflowOwner])
}

func TestOrgResolver_Cache_HitOnSecondCall(t *testing.T) {
	ctx := context.Background()
	client := &mockLinkingClient{}
	cache := newMockCacheStore()

	workflowOwner := "0xabcdef1234567890"

	cfg := Config{
		URL:                           "test-url",
		WorkflowRegistryAddress:       "0x1234567890abcdef",
		WorkflowRegistryChainSelector: 1,
		CacheEnabled:                  true,
		CacheStore:                    cache,
		Client:                        client,
	}

	resolver, err := cfg.New(logger.Test(t))
	require.NoError(t, err)

	// First call: cache miss -> fetches from linking service and stores.
	orgID1, err := resolver.Get(ctx, workflowOwner)
	require.NoError(t, err)
	require.Equal(t, "org-"+workflowOwner, orgID1)
	require.Equal(t, 1, cache.upsertCalls)

	// Second call: should hit the cache that was populated on the first call.
	orgID2, err := resolver.Get(ctx, workflowOwner)
	require.NoError(t, err)
	require.Equal(t, "org-"+workflowOwner, orgID2)

	// Only one upsert should have happened (on the first call).
	require.Equal(t, 1, cache.upsertCalls)
	require.Equal(t, 2, cache.getCalls)
}

func TestOrgResolver_Cache_ErrorFallsBackToLinkingService(t *testing.T) {
	ctx := context.Background()
	client := &mockLinkingClient{}
	cache := newMockCacheStore()
	cache.getErr = errors.New("cache store unavailable")

	workflowOwner := "0xabcdef1234567890"

	cfg := Config{
		URL:                           "test-url",
		WorkflowRegistryAddress:       "0x1234567890abcdef",
		WorkflowRegistryChainSelector: 1,
		CacheEnabled:                  true,
		CacheStore:                    cache,
		Client:                        client,
	}

	resolver, err := cfg.New(logger.Test(t))
	require.NoError(t, err)

	// A cache store error should not prevent resolution; it falls back to the
	// linking service.
	orgID, err := resolver.Get(ctx, workflowOwner)
	require.NoError(t, err)
	require.Equal(t, "org-"+workflowOwner, orgID)

	require.Equal(t, 1, cache.getCalls)
	require.Equal(t, 1, cache.upsertCalls)
}

func TestOrgResolver_Cache_SQLErrNoRowsTreatedAsMiss(t *testing.T) {
	ctx := context.Background()
	client := &mockLinkingClient{}
	cache := newMockCacheStore()
	cache.getErr = sql.ErrNoRows

	workflowOwner := "0xabcdef1234567890"

	cfg := Config{
		URL:                           "test-url",
		WorkflowRegistryAddress:       "0x1234567890abcdef",
		WorkflowRegistryChainSelector: 1,
		CacheEnabled:                  true,
		CacheStore:                    cache,
		Client:                        client,
	}

	resolver, err := cfg.New(logger.Test(t))
	require.NoError(t, err)

	orgID, err := resolver.Get(ctx, workflowOwner)
	require.NoError(t, err)
	require.Equal(t, "org-"+workflowOwner, orgID)

	// sql.ErrNoRows is treated as a miss, so the linking service is consulted
	// and the result is stored.
	require.Equal(t, 1, cache.upsertCalls)
}

func TestOrgResolver_Cache_UpsertErrorDoesNotFailGet(t *testing.T) {
	ctx := context.Background()
	client := &mockLinkingClient{}
	cache := newMockCacheStore()
	cache.upsertErr = errors.New("cache write failed")

	workflowOwner := "0xabcdef1234567890"

	cfg := Config{
		URL:                           "test-url",
		WorkflowRegistryAddress:       "0x1234567890abcdef",
		WorkflowRegistryChainSelector: 1,
		CacheEnabled:                  true,
		CacheStore:                    cache,
		Client:                        client,
	}

	resolver, err := cfg.New(logger.Test(t))
	require.NoError(t, err)

	// A cache write failure should be logged but must not cause Get to fail.
	orgID, err := resolver.Get(ctx, workflowOwner)
	require.NoError(t, err)
	require.Equal(t, "org-"+workflowOwner, orgID)

	require.Equal(t, 1, cache.upsertCalls)
	// The data should not have been written due to the upsert error.
	_, exists := cache.data[workflowOwner]
	require.False(t, exists)
}

func TestOrgResolver_Cache_RequiresCacheStoreWhenEnabled(t *testing.T) {
	cfg := Config{
		URL:                           "test-url",
		WorkflowRegistryAddress:       "0x1234567890abcdef",
		WorkflowRegistryChainSelector: 1,
		CacheEnabled:                  true,
		CacheStore:                    nil, // missing store
		Client:                        &mockLinkingClient{},
	}

	_, err := cfg.New(logger.Test(t))
	require.Error(t, err)
	require.Contains(t, err.Error(), "CacheStore is required when CacheEnabled is true")
}

func TestOrgResolver_Cache_DisabledDoesNotUseCacheStore(t *testing.T) {
	ctx := context.Background()
	client := &mockLinkingClient{}
	cache := newMockCacheStore()

	workflowOwner := "0xabcdef1234567890"

	cfg := Config{
		URL:                           "test-url",
		WorkflowRegistryAddress:       "0x1234567890abcdef",
		WorkflowRegistryChainSelector: 1,
		CacheEnabled:                  false,
		CacheStore:                    cache, // provided but disabled
		Client:                        client,
	}

	resolver, err := cfg.New(logger.Test(t))
	require.NoError(t, err)

	orgID, err := resolver.Get(ctx, workflowOwner)
	require.NoError(t, err)
	require.Equal(t, "org-"+workflowOwner, orgID)

	// With caching disabled, neither GetOrg nor UpsertOrg should be called.
	require.Equal(t, 0, cache.getCalls)
	require.Equal(t, 0, cache.upsertCalls)
}
