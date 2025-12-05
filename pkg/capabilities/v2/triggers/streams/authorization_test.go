package streams_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/triggers/streams"
)

func TestAuthorizerDisabled(t *testing.T) {
	config := streams.AuthConfig{
		Enabled: false,
	}

	auth, err := streams.NewAuthorizer(config)
	require.NoError(t, err)

	// Any workflow should be allowed when disabled
	metadata := capabilities.RequestMetadata{
		WorkflowID:    "any-workflow",
		WorkflowName:  "anything",
		WorkflowOwner: "0xAnyOwner",
	}

	err = auth.IsAuthorized(metadata)
	assert.NoError(t, err, "Should allow all workflows when authorization is disabled")
}

func TestAuthorizerWorkflowIDAllowlist(t *testing.T) {
	config := streams.AuthConfig{
		Enabled: true,
		AllowedWorkflowIDs: []string{
			"df-prod-1",
			"df-prod-2",
			"df-staging-1",
		},
	}

	auth, err := streams.NewAuthorizer(config)
	require.NoError(t, err)

	tests := []struct {
		name        string
		workflowID  string
		expectError bool
	}{
		{"workflow in allowlist 1", "df-prod-1", false},
		{"workflow in allowlist 2", "df-prod-2", false},
		{"workflow in allowlist 3", "df-staging-1", false},
		{"workflow not in allowlist", "other-workflow", true},
		{"workflow similar but not exact", "df-prod-10", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := capabilities.RequestMetadata{
				WorkflowID: tt.workflowID,
			}
			err := auth.IsAuthorized(metadata)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "not authorized")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthorizerWorkflowIDPattern(t *testing.T) {
	config := streams.AuthConfig{
		Enabled:                true,
		AllowedWorkflowPattern: "^df-.*-mainnet$",
	}

	auth, err := streams.NewAuthorizer(config)
	require.NoError(t, err)

	tests := []struct {
		name        string
		workflowID  string
		expectError bool
	}{
		{"matches pattern 1", "df-btc-mainnet", false},
		{"matches pattern 2", "df-eth-mainnet", false},
		{"matches pattern 3", "df-link-usd-mainnet", false},
		{"doesn't match - no prefix", "other-mainnet", true},
		{"doesn't match - no suffix", "df-btc", true},
		{"doesn't match - wrong suffix", "df-btc-testnet", true},
		{"doesn't match - completely different", "malicious-workflow", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := capabilities.RequestMetadata{
				WorkflowID: tt.workflowID,
			}
			err := auth.IsAuthorized(metadata)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthorizerWorkflowOwner(t *testing.T) {
	config := streams.AuthConfig{
		Enabled: true,
		AllowedWorkflowOwners: []string{
			"0xDFOwner1",
			"0xDFOwner2",
		},
	}

	auth, err := streams.NewAuthorizer(config)
	require.NoError(t, err)

	tests := []struct {
		name          string
		workflowOwner string
		expectError   bool
	}{
		{"owner in allowlist 1", "0xDFOwner1", false},
		{"owner in allowlist 2", "0xDFOwner2", false},
		{"owner not in allowlist", "0xOtherOwner", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := capabilities.RequestMetadata{
				WorkflowID:    "some-workflow",
				WorkflowOwner: tt.workflowOwner,
			}
			err := auth.IsAuthorized(metadata)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthorizerWorkflowNamePattern(t *testing.T) {
	config := streams.AuthConfig{
		Enabled:                    true,
		AllowedWorkflowNamePattern: "data-feed",
	}

	auth, err := streams.NewAuthorizer(config)
	require.NoError(t, err)

	tests := []struct {
		name         string
		workflowName string
		expectError  bool
	}{
		{"matches pattern 1", "data-feed-btc-usd", false},
		{"matches pattern 2", "mainnet-data-feed", false},
		{"matches pattern 3", "data-feed", false},
		{"doesn't match", "other-workflow", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := capabilities.RequestMetadata{
				WorkflowID:   "some-id",
				WorkflowName: tt.workflowName,
			}
			err := auth.IsAuthorized(metadata)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthorizerInvalidPattern(t *testing.T) {
	config := streams.AuthConfig{
		Enabled:                true,
		AllowedWorkflowPattern: "[invalid(regex",
	}

	_, err := streams.NewAuthorizer(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid workflow ID pattern")
}

func TestAuthorizerInvalidNamePattern(t *testing.T) {
	config := streams.AuthConfig{
		Enabled:                    true,
		AllowedWorkflowNamePattern: "[invalid(regex",
	}

	_, err := streams.NewAuthorizer(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid workflow name pattern")
}

func TestAuthorizerCombinedChecksAnyMatch(t *testing.T) {
	// If ANY check passes, workflow is authorized
	config := streams.AuthConfig{
		Enabled:                true,
		AllowedWorkflowPattern: "^df-.*",
		AllowedWorkflowOwners:  []string{"0xDFOwner"},
	}

	auth, err := streams.NewAuthorizer(config)
	require.NoError(t, err)

	tests := []struct {
		name          string
		workflowID    string
		workflowOwner string
		expectError   bool
	}{
		{
			name:          "matches ID pattern",
			workflowID:    "df-prod-1",
			workflowOwner: "0xOther",
			expectError:   false, // Passes ID pattern check
		},
		{
			name:          "matches owner",
			workflowID:    "other-workflow",
			workflowOwner: "0xDFOwner",
			expectError:   false, // Passes owner check
		},
		{
			name:          "matches both",
			workflowID:    "df-prod-1",
			workflowOwner: "0xDFOwner",
			expectError:   false, // Passes both checks
		},
		{
			name:          "matches neither",
			workflowID:    "other-workflow",
			workflowOwner: "0xOther",
			expectError:   true, // Fails both checks
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := capabilities.RequestMetadata{
				WorkflowID:    tt.workflowID,
				WorkflowOwner: tt.workflowOwner,
			}
			err := auth.IsAuthorized(metadata)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthorizerNoChecksConfigured(t *testing.T) {
	// If no checks configured, deny by default
	config := streams.AuthConfig{
		Enabled: true,
		// No checks configured
	}

	auth, err := streams.NewAuthorizer(config)
	require.NoError(t, err)

	metadata := capabilities.RequestMetadata{
		WorkflowID: "any-workflow",
	}

	err = auth.IsAuthorized(metadata)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no authorization checks configured")
}

func TestDefaultDataFeedsAuthorizer(t *testing.T) {
	auth, err := streams.NewDefaultDataFeedsAuthorizer()
	require.NoError(t, err)

	tests := []struct {
		name         string
		workflowID   string
		workflowName string
		expectError  bool
	}{
		{
			name:         "DF workflow ID",
			workflowID:   "df-btc-usd",
			workflowName: "",
			expectError:  false,
		},
		{
			name:         "DF workflow name",
			workflowID:   "other-id",
			workflowName: "data-feed-eth-usd",
			expectError:  false,
		},
		{
			name:         "Neither matches",
			workflowID:   "other-id",
			workflowName: "other-name",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := capabilities.RequestMetadata{
				WorkflowID:   tt.workflowID,
				WorkflowName: tt.workflowName,
			}
			err := auth.IsAuthorized(metadata)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthorizerString(t *testing.T) {
	// Test disabled
	auth1, _ := streams.NewAuthorizer(streams.AuthConfig{Enabled: false})
	str := auth1.String()
	assert.Contains(t, str, "Disabled")

	// Test with ID allowlist
	auth2, _ := streams.NewAuthorizer(streams.AuthConfig{
		Enabled:            true,
		AllowedWorkflowIDs: []string{"id1", "id2"},
	})
	str = auth2.String()
	assert.Contains(t, str, "Enabled")
	assert.Contains(t, str, "Workflow ID allowlist")

	// Test with pattern
	auth3, _ := streams.NewAuthorizer(streams.AuthConfig{
		Enabled:                true,
		AllowedWorkflowPattern: "^df-.*",
	})
	str = auth3.String()
	assert.Contains(t, str, "pattern")

	// Test with owner allowlist
	auth4, _ := streams.NewAuthorizer(streams.AuthConfig{
		Enabled:               true,
		AllowedWorkflowOwners: []string{"0xOwner1"},
	})
	str = auth4.String()
	assert.Contains(t, str, "owner")
}

// BenchmarkAuthorizerCheck benchmarks the authorization check
func BenchmarkAuthorizerCheck(b *testing.B) {
	config := streams.AuthConfig{
		Enabled:                true,
		AllowedWorkflowPattern: "^df-.*",
	}

	auth, _ := streams.NewAuthorizer(config)

	metadata := capabilities.RequestMetadata{
		WorkflowID: "df-prod-1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = auth.IsAuthorized(metadata)
	}
}

func BenchmarkAuthorizerCheckAllowlist(b *testing.B) {
	// Create large allowlist
	allowlist := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		allowlist[i] = fmt.Sprintf("df-workflow-%d", i)
	}

	config := streams.AuthConfig{
		Enabled:            true,
		AllowedWorkflowIDs: allowlist,
	}

	auth, _ := streams.NewAuthorizer(config)

	metadata := capabilities.RequestMetadata{
		WorkflowID: "df-workflow-500", // Middle of allowlist
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = auth.IsAuthorized(metadata)
	}
}
