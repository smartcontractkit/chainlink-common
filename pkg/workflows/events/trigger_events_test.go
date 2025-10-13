package events

import (
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder/beholdertest"
	"github.com/smartcontractkit/chainlink-common/pkg/custmsg"
	workflowsevents "github.com/smartcontractkit/chainlink-protos/workflows/go/v2"
)

func TestEmitTriggerExecutionStarted(t *testing.T) {
	ctx := context.Background()
	beholderObserver := beholdertest.NewObserver(t)

	// Test data
	expectedTriggerID := "test-trigger"
	expectedWorkflowID := "test-workflow"
	expectedWorkflowExecutionID := "test-execution"
	expectedWorkflowOwner := "test-owner"
	expectedWorkflowName := "test-workflow-name"
	expectedOrganizationID := "test-org"
	expectedDonID := "1"
	expectedDonVersion := "v1.0.0"

	labeler := custmsg.NewLabeler().With(
		KeyTriggerID, expectedTriggerID,
		KeyWorkflowID, expectedWorkflowID,
		KeyWorkflowExecutionID, expectedWorkflowExecutionID,
		KeyWorkflowOwner, expectedWorkflowOwner,
		KeyWorkflowName, expectedWorkflowName,
		KeyOrganizationID, expectedOrganizationID,
		KeyDonID, expectedDonID,
		KeyDonVersion, expectedDonVersion,
	)

	err := EmitTriggerExecutionStarted(ctx, labeler)
	require.NoError(t, err)

	// Verify the message was emitted with correct beholder attributes
	msgs := beholderObserver.Messages(t, "beholder_entity", "workflows.v2.TriggerExecutionStarted")
	require.Len(t, msgs, 1, "Expected exactly one TriggerExecutionStarted message")

	msg := msgs[0]

	// Verify beholder attributes
	assert.Equal(t, "workflows.v2.trigger_execution_started", msg.Attrs["beholder_data_schema"])
	assert.Equal(t, "platform", msg.Attrs["beholder_domain"])
	assert.Equal(t, "workflows.v2.TriggerExecutionStarted", msg.Attrs["beholder_entity"])

	// Unmarshal and verify the protobuf message content
	var event workflowsevents.TriggerExecutionStarted
	require.NoError(t, proto.Unmarshal(msg.Body, &event))

	// Verify required fields
	assert.Equal(t, expectedTriggerID, event.TriggerID)
	assert.Equal(t, expectedWorkflowExecutionID, event.WorkflowExecutionID)

	// Verify WorkflowKey fields
	require.NotNil(t, event.Workflow)
	assert.Equal(t, expectedWorkflowID, event.Workflow.WorkflowID)
	assert.Equal(t, expectedWorkflowOwner, event.Workflow.WorkflowOwner)
	assert.Equal(t, expectedWorkflowName, event.Workflow.WorkflowName)
	assert.Equal(t, expectedOrganizationID, event.Workflow.OrganizationID)

	// Verify CreInfo fields (optional)
	require.NotNil(t, event.CreInfo)
	assert.Equal(t, int32(1), event.CreInfo.DonID)
	assert.Equal(t, expectedDonVersion, event.CreInfo.DonVersion)

	// Verify timestamp format (RFC3339)
	timeMatcher := regexp.MustCompile(`[0-9\-]{10}T[0-9:]{8}[Z\-\+][0-9:]*`)
	assert.True(t, timeMatcher.MatchString(event.Timestamp), "Timestamp should be in RFC3339 format: %s", event.Timestamp)
}

func TestEmitTriggerExecutionStarted_MissingRequiredFields(t *testing.T) {
	ctx := context.Background()

	// Test missing trigger ID
	labeler := custmsg.NewLabeler().With(
		KeyWorkflowID, "test-workflow",
		KeyWorkflowExecutionID, "test-execution",
	)

	err := EmitTriggerExecutionStarted(ctx, labeler)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required field: trigger_id")

	labeler = custmsg.NewLabeler().With(
		KeyTriggerID, "test-trigger",
		KeyWorkflowExecutionID, "test-execution",
	)

	err = EmitTriggerExecutionStarted(ctx, labeler)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required field: workflow_id")

	labeler = custmsg.NewLabeler().With(
		KeyTriggerID, "test-trigger",
		KeyWorkflowID, "test-workflow",
	)

	err = EmitTriggerExecutionStarted(ctx, labeler)
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing required field: workflow_execution_id")
}
