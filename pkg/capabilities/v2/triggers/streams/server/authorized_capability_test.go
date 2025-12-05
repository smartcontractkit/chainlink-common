package server_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/triggers/streams/server"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/types/core"
)

func TestAuthorizedCapabilityBlocksUnauthorizedWorkflows(t *testing.T) {
	lggr, _ := logger.New()
	mockCap := &mockStreamsCapability{}

	// Create authorized capability with DF authorization
	authCap, err := server.NewDefaultDataFeedsCapability(mockCap, lggr)
	require.NoError(t, err)

	// Test 1: Authorized DF workflow (by ID pattern) - should succeed
	ch, err := authCap.RegisterTrigger(
		context.Background(),
		"trigger-1",
		capabilities.RequestMetadata{
			WorkflowID:    "df-btc-usd",
			WorkflowOwner: "0xDF001",
			WorkflowName:  "Bitcoin Data Feed",
		},
		&streams.Config{FeedIds: []string{"0x001"}},
	)
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	// Test 2: Authorized DF workflow (by name pattern) - should succeed
	ch, err = authCap.RegisterTrigger(
		context.Background(),
		"trigger-2",
		capabilities.RequestMetadata{
			WorkflowID:    "workflow-123",
			WorkflowOwner: "0xDF002",
			WorkflowName:  "mainnet-data-feed-eth",
		},
		&streams.Config{FeedIds: []string{"0x002"}},
	)
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	// Test 3: Unauthorized workflow (doesn't match ID or name) - should fail with auth error
	ch, err = authCap.RegisterTrigger(
		context.Background(),
		"trigger-3",
		capabilities.RequestMetadata{
			WorkflowID:    "other-workflow",
			WorkflowOwner: "0xOTHER",
			WorkflowName:  "Other Workflow",
		},
		&streams.Config{FeedIds: []string{"0x003"}},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authorization failed")
	assert.Nil(t, ch)
}

func TestAuthorizedCapabilityWithCustomConfig(t *testing.T) {
	lggr, _ := logger.New()
	mockCap := &mockStreamsCapability{}

	// Custom authorization: specific allowlist
	authConfig := streams.AuthConfig{
		Enabled: true,
		AllowedWorkflowIDs: []string{
			"df-prod-btc-usd",
			"df-prod-eth-usd",
		},
	}

	authCap, err := server.NewAuthorizedStreamsCapability(mockCap, authConfig, lggr)
	require.NoError(t, err)

	// Test allowed workflow
	ch, err := authCap.RegisterTrigger(
		context.Background(),
		"trigger-1",
		capabilities.RequestMetadata{
			WorkflowID: "df-prod-btc-usd",
		},
		&streams.Config{},
	)
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	// Test non-allowed workflow
	ch, err = authCap.RegisterTrigger(
		context.Background(),
		"trigger-2",
		capabilities.RequestMetadata{
			WorkflowID: "df-prod-link-usd", // Not in allowlist
		},
		&streams.Config{},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authorization failed")
	assert.Nil(t, ch)
}

func TestAuthorizedCapabilityWorkflowOwnerAllowlist(t *testing.T) {
	lggr, _ := logger.New()
	mockCap := &mockStreamsCapability{}

	// Authorization by owner address
	authConfig := streams.AuthConfig{
		Enabled: true,
		AllowedWorkflowOwners: []string{
			"0xDFOwner1",
			"0xDFOwner2",
		},
	}

	authCap, err := server.NewAuthorizedStreamsCapability(mockCap, authConfig, lggr)
	require.NoError(t, err)

	// Test allowed owner
	ch, err := authCap.RegisterTrigger(
		context.Background(),
		"trigger-1",
		capabilities.RequestMetadata{
			WorkflowID:    "any-workflow-id",
			WorkflowOwner: "0xDFOwner1",
		},
		&streams.Config{},
	)
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	// Test non-allowed owner
	ch, err = authCap.RegisterTrigger(
		context.Background(),
		"trigger-2",
		capabilities.RequestMetadata{
			WorkflowID:    "any-workflow-id",
			WorkflowOwner: "0xOtherOwner",
		},
		&streams.Config{},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authorization failed")
	assert.Nil(t, ch)
}

func TestAuthorizedCapabilityDisabled(t *testing.T) {
	lggr, _ := logger.New()
	mockCap := &mockStreamsCapability{}

	// Disable authorization
	authConfig := streams.AuthConfig{
		Enabled: false,
	}

	authCap, err := server.NewAuthorizedStreamsCapability(mockCap, authConfig, lggr)
	require.NoError(t, err)

	// Any workflow should be allowed
	ch, err := authCap.RegisterTrigger(
		context.Background(),
		"trigger-1",
		capabilities.RequestMetadata{
			WorkflowID:    "any-workflow",
			WorkflowOwner: "0xAnyone",
			WorkflowName:  "Anything",
		},
		&streams.Config{},
	)
	assert.NoError(t, err)
	assert.NotNil(t, ch)
}

func TestAuthorizedCapabilityUnregisterAlsoChecksAuth(t *testing.T) {
	lggr, _ := logger.New()
	mockCap := &mockStreamsCapability{}

	// Authorization enabled
	authConfig := streams.AuthConfig{
		Enabled:                true,
		AllowedWorkflowPattern: "^df-.*",
	}

	authCap, err := server.NewAuthorizedStreamsCapability(mockCap, authConfig, lggr)
	require.NoError(t, err)

	// Test authorized unregister
	err = authCap.UnregisterTrigger(
		context.Background(),
		"trigger-1",
		capabilities.RequestMetadata{
			WorkflowID: "df-prod-btc",
		},
		&streams.Config{},
	)
	assert.NoError(t, err)

	// Test unauthorized unregister
	err = authCap.UnregisterTrigger(
		context.Background(),
		"trigger-2",
		capabilities.RequestMetadata{
			WorkflowID: "other-workflow",
		},
		&streams.Config{},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authorization failed")
}

// Mock implementations for testing

type mockStreamsCapability struct {
	registerCalled bool
}

func (m *mockStreamsCapability) RegisterTrigger(ctx context.Context, triggerID string, metadata capabilities.RequestMetadata, input *streams.Config) (<-chan capabilities.TriggerAndId[*streams.Feed], error) {
	m.registerCalled = true
	ch := make(chan capabilities.TriggerAndId[*streams.Feed])
	close(ch)
	return ch, nil
}

func (m *mockStreamsCapability) UnregisterTrigger(ctx context.Context, triggerID string, metadata capabilities.RequestMetadata, input *streams.Config) error {
	return nil
}

func (m *mockStreamsCapability) Start(ctx context.Context) error {
	return nil
}

func (m *mockStreamsCapability) Close() error {
	return nil
}

func (m *mockStreamsCapability) HealthReport() map[string]error {
	return map[string]error{}
}

func (m *mockStreamsCapability) Name() string {
	return "MockStreams"
}

func (m *mockStreamsCapability) Description() string {
	return "Mock"
}

func (m *mockStreamsCapability) Ready() error {
	return nil
}

func (m *mockStreamsCapability) Initialise(ctx context.Context, deps core.StandardCapabilitiesDependencies) error {
	return nil
}

