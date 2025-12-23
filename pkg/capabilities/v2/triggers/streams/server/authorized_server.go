package server

import (
	"context"
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/triggers/streams"
	"github.com/smartcontractkit/chainlink-common/pkg/logger"
)

// AuthorizedStreamsCapability wraps StreamsCapability with authorization checks
type AuthorizedStreamsCapability struct {
	StreamsCapability
	authorizer *streams.Authorizer
	lggr       logger.Logger
}

// NewAuthorizedStreamsCapability creates a new capability with authorization enabled
func NewAuthorizedStreamsCapability(capability StreamsCapability, authConfig streams.AuthConfig, lggr logger.Logger) (*AuthorizedStreamsCapability, error) {
	authorizer, err := streams.NewAuthorizer(authConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorizer: %w", err)
	}

	return &AuthorizedStreamsCapability{
		StreamsCapability: capability,
		authorizer:        authorizer,
		lggr:              logger.Named(lggr, "AuthorizedStreamsCapability"),
	}, nil
}

// NewDefaultDataFeedsCapability creates a capability with default Data Feeds authorization
// Only workflows with IDs starting with "df-" or names containing "data-feed" will be allowed
func NewDefaultDataFeedsCapability(capability StreamsCapability, lggr logger.Logger) (*AuthorizedStreamsCapability, error) {
	authConfig := streams.AuthConfig{
		Enabled:                    true,
		AllowedWorkflowPattern:     "^df-.*",
		AllowedWorkflowNamePattern: "data-feed",
	}

	return NewAuthorizedStreamsCapability(capability, authConfig, lggr)
}

// RegisterTrigger wraps the base RegisterTrigger with authorization check
func (a *AuthorizedStreamsCapability) RegisterTrigger(ctx context.Context, triggerID string, metadata capabilities.RequestMetadata, input *streams.Config) (<-chan capabilities.TriggerAndId[*streams.Feed], error) {
	// Authorization check
	if err := a.authorizer.IsAuthorized(metadata); err != nil {
		a.lggr.Warnw("Unauthorized trigger registration attempt",
			"workflowID", metadata.WorkflowID,
			"workflowOwner", metadata.WorkflowOwner,
			"error", err,
		)
		return nil, fmt.Errorf("authorization failed: %w", err)
	}

	a.lggr.Debugw("Authorized trigger registration",
		"workflowID", metadata.WorkflowID,
		"triggerID", triggerID,
	)

	// Call the underlying implementation
	return a.StreamsCapability.RegisterTrigger(ctx, triggerID, metadata, input)
}

// UnregisterTrigger wraps the base UnregisterTrigger with authorization check
func (a *AuthorizedStreamsCapability) UnregisterTrigger(ctx context.Context, triggerID string, metadata capabilities.RequestMetadata, input *streams.Config) error {
	// Authorization check
	if err := a.authorizer.IsAuthorized(metadata); err != nil {
		a.lggr.Warnw("Unauthorized trigger unregistration attempt",
			"workflowID", metadata.WorkflowID,
			"error", err,
		)
		return fmt.Errorf("authorization failed: %w", err)
	}

	a.lggr.Debugw("Authorized trigger unregistration",
		"workflowID", metadata.WorkflowID,
		"triggerID", triggerID,
	)

	// Call the underlying implementation
	return a.StreamsCapability.UnregisterTrigger(ctx, triggerID, metadata, input)
}

// NewAuthorizedStreamsServer creates a server wrapping an authorized capability
func NewAuthorizedStreamsServer(capability StreamsCapability, authConfig streams.AuthConfig, lggr logger.Logger) (*StreamsServer, error) {
	authCap, err := NewAuthorizedStreamsCapability(capability, authConfig, lggr)
	if err != nil {
		return nil, err
	}

	return NewStreamsServer(authCap), nil
}

// NewDefaultDataFeedsServer creates a server with default Data Feeds authorization
func NewDefaultDataFeedsServer(capability StreamsCapability, lggr logger.Logger) (*StreamsServer, error) {
	authCap, err := NewDefaultDataFeedsCapability(capability, lggr)
	if err != nil {
		return nil, err
	}

	return NewStreamsServer(authCap), nil
}
