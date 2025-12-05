package streams

import (
	"fmt"
	"regexp"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
)

// Authorizer handles workflow authorization for streams trigger
// Ensures only authorized workflows (e.g., Data Feeds) can use the trigger
type Authorizer struct {
	allowedWorkflowIDs          map[string]bool
	allowedWorkflowPattern      *regexp.Regexp
	allowedWorkflowOwners       map[string]bool
	allowedWorkflowNamePattern  *regexp.Regexp
	enabled                     bool
}

// AuthConfig configures authorization rules for the streams trigger
type AuthConfig struct {
	// Enable authorization checks (set to false to disable authorization)
	Enabled bool
	
	// AllowedWorkflowIDs is an explicit allowlist of workflow IDs
	AllowedWorkflowIDs []string
	
	// AllowedWorkflowPattern is a regex pattern for allowed workflow IDs
	// Example: "^df-.*" allows all workflows starting with "df-"
	AllowedWorkflowPattern string
	
	// AllowedWorkflowOwners is an explicit allowlist of workflow owner addresses
	// Example: ["0xDFOwner1", "0xDFOwner2"]
	AllowedWorkflowOwners []string
	
	// AllowedWorkflowNamePattern is a regex pattern for allowed workflow names
	// Example: "^data-feed-.*" for workflow names starting with "data-feed-"
	AllowedWorkflowNamePattern string
}

// NewAuthorizer creates a new authorizer with the given configuration
func NewAuthorizer(config AuthConfig) (*Authorizer, error) {
	auth := &Authorizer{
		enabled:               config.Enabled,
		allowedWorkflowIDs:    make(map[string]bool),
		allowedWorkflowOwners: make(map[string]bool),
	}

	// If authorization is disabled, return early
	if !config.Enabled {
		return auth, nil
	}

	// Build workflow ID allowlist map for O(1) lookups
	for _, id := range config.AllowedWorkflowIDs {
		auth.allowedWorkflowIDs[id] = true
	}

	// Build workflow owner allowlist map for O(1) lookups
	for _, owner := range config.AllowedWorkflowOwners {
		auth.allowedWorkflowOwners[owner] = true
	}

	// Compile workflow ID pattern if provided
	if config.AllowedWorkflowPattern != "" {
		pattern, err := regexp.Compile(config.AllowedWorkflowPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid workflow ID pattern '%s': %w", config.AllowedWorkflowPattern, err)
		}
		auth.allowedWorkflowPattern = pattern
	}

	// Compile workflow name pattern if provided
	if config.AllowedWorkflowNamePattern != "" {
		pattern, err := regexp.Compile(config.AllowedWorkflowNamePattern)
		if err != nil {
			return nil, fmt.Errorf("invalid workflow name pattern '%s': %w", config.AllowedWorkflowNamePattern, err)
		}
		auth.allowedWorkflowNamePattern = pattern
	}

	return auth, nil
}

// NewDefaultDataFeedsAuthorizer creates an authorizer for Data Feeds workflows
// This is a convenience function for the common case
// Allows workflows with IDs starting with "df-" or names containing "data-feed"
func NewDefaultDataFeedsAuthorizer() (*Authorizer, error) {
	return NewAuthorizer(AuthConfig{
		Enabled:                    true,
		AllowedWorkflowPattern:     "^df-.*", // Allow workflow IDs starting with "df-"
		AllowedWorkflowNamePattern: "data-feed", // Allow workflow names containing "data-feed"
	})
}

// IsAuthorized checks if a workflow is authorized to use the streams trigger
// Returns nil if authorized, error otherwise
// Authorization checks (in order):
// 1. Explicit workflow ID allowlist
// 2. Workflow ID pattern matching
// 3. Workflow owner address allowlist
// 4. Workflow name pattern matching
// If ANY check passes, the workflow is authorized
func (a *Authorizer) IsAuthorized(metadata capabilities.RequestMetadata) error {
	// If authorization is disabled, allow all
	if !a.enabled {
		return nil
	}

	workflowID := metadata.WorkflowID
	workflowName := metadata.WorkflowName
	if workflowName == "" {
		workflowName = metadata.DecodedWorkflowName
	}
	workflowOwner := metadata.WorkflowOwner

	// If no checks configured, deny by default
	if len(a.allowedWorkflowIDs) == 0 && a.allowedWorkflowPattern == nil &&
		len(a.allowedWorkflowOwners) == 0 && a.allowedWorkflowNamePattern == nil {
		return fmt.Errorf("workflow %s: no authorization checks configured, denying by default", workflowID)
	}

	// Check 1: Explicit workflow ID allowlist
	if len(a.allowedWorkflowIDs) > 0 {
		if a.allowedWorkflowIDs[workflowID] {
			return nil // Authorized
		}
	}

	// Check 2: Workflow ID pattern matching
	if a.allowedWorkflowPattern != nil {
		if a.allowedWorkflowPattern.MatchString(workflowID) {
			return nil // Authorized
		}
	}

	// Check 3: Workflow owner allowlist
	if len(a.allowedWorkflowOwners) > 0 && workflowOwner != "" {
		if a.allowedWorkflowOwners[workflowOwner] {
			return nil // Authorized
		}
	}

	// Check 4: Workflow name pattern matching
	if a.allowedWorkflowNamePattern != nil && workflowName != "" {
		if a.allowedWorkflowNamePattern.MatchString(workflowName) {
			return nil // Authorized
		}
	}

	// None of the checks passed
	return fmt.Errorf("workflow %s (name: %s, owner: %s) not authorized", workflowID, workflowName, workflowOwner)
}

// String returns a human-readable description of the authorization rules
func (a *Authorizer) String() string {
	if !a.enabled {
		return "Authorization: Disabled (all workflows allowed)"
	}

	desc := "Authorization: Enabled\n"
	
	if len(a.allowedWorkflowIDs) > 0 {
		desc += fmt.Sprintf("  - Workflow ID allowlist: %d entries\n", len(a.allowedWorkflowIDs))
	}
	
	if a.allowedWorkflowPattern != nil {
		desc += fmt.Sprintf("  - Workflow ID pattern: %s\n", a.allowedWorkflowPattern.String())
	}
	
	if len(a.allowedWorkflowOwners) > 0 {
		desc += fmt.Sprintf("  - Workflow owner allowlist: %d entries\n", len(a.allowedWorkflowOwners))
	}
	
	if a.allowedWorkflowNamePattern != nil {
		desc += fmt.Sprintf("  - Workflow name pattern: %s\n", a.allowedWorkflowNamePattern.String())
	}
	
	return desc
}

