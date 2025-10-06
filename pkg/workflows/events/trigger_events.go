package events

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/smartcontractkit/chainlink-common/pkg/beholder"
	"github.com/smartcontractkit/chainlink-common/pkg/custmsg"
	workflowsevents "github.com/smartcontractkit/chainlink-protos/workflows/go/v2"
)

// Label keys for trigger events
const (
	KeyTriggerID           = "trigger_id"
	KeyWorkflowID          = "workflow_id"
	KeyWorkflowOwner       = "workflow_owner"
	KeyWorkflowName        = "workflow_name"
	KeyWorkflowExecutionID = "workflow_execution_id"
	KeyDonID               = "don_id"
	KeyDonVersion          = "don_version"
	KeyOrganizationID      = "organization_id"
)

// EmitTriggerExecutionStarted emits a TriggerExecutionStarted event using the provided labeler
func EmitTriggerExecutionStarted(ctx context.Context, labeler custmsg.MessageEmitter) error {
	labels := labeler.Labels()

	// Extract required fields
	triggerID, ok := labels[KeyTriggerID]
	if !ok {
		return fmt.Errorf("missing required field: %s", KeyTriggerID)
	}

	workflowID, ok := labels[KeyWorkflowID]
	if !ok {
		return fmt.Errorf("missing required field: %s", KeyWorkflowID)
	}

	workflowExecutionID, ok := labels[KeyWorkflowExecutionID]
	if !ok {
		return fmt.Errorf("missing required field: %s", KeyWorkflowExecutionID)
	}

	event := &workflowsevents.TriggerExecutionStarted{
		TriggerID:           triggerID,
		WorkflowExecutionID: workflowExecutionID,
		Workflow: &workflowsevents.WorkflowKey{
			WorkflowID:     workflowID,
			WorkflowOwner:  labels[KeyWorkflowOwner],
			WorkflowName:   labels[KeyWorkflowName],
			OrganizationID: labels[KeyOrganizationID],
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Optional; downstream consumers could infer from csa public key,
	// as of now Beholder/ChiP autohydrates csa public key
	if donIDStr, exists := labels[KeyDonID]; exists {
		if donID, err := strconv.ParseInt(donIDStr, 10, 32); err == nil {
			event.CreInfo = &workflowsevents.CreInfo{
				DonID:      int32(donID),
				DonVersion: labels[KeyDonVersion],
			}
		}
	}

	b, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal TriggerExecutionStarted event: %w", err)
	}

	return beholder.GetEmitter().Emit(ctx, b,
		"beholder_data_schema", "workflows.v2.trigger_execution_started", // required
		"beholder_domain", "platform", // required
		"beholder_entity", "workflows.v2.TriggerExecutionStarted") // required
}
