package resourcemanager

import (
	"errors"
	"strconv"
	"strings"
)

// Workflow capability usage records: one METER_ACTION_USAGE MeterRecord per
// billable capability event of a workflow execution (compute duration, gas per
// chain write). The billing service turns each into a billing_records row with
// event id "cre:workflow:<workflow_id>:<execution_id>:<capability_event_id>",
// built from Utilization.ResourceId ("<workflow_id>:<execution_id>") and
// Utilization.EventId (the capability event id). Both must be identical on every
// node of the DON: they are the quorum and dedup key.
const (
	// WorkflowRecordType is the billing record type shared by all workflow
	// capability usage resource types.
	WorkflowRecordType = "cre:workflow"

	// ResourceTypeWorkflowCompute meters ordinary workflow compute in
	// milliseconds. The capability event id is the workflow execution id.
	ResourceTypeWorkflowCompute = WorkflowRecordType + ":compute"

	// ResourceTypeWorkflowGasPrefix is the prefix of the per-chain gas resource
	// type; see WorkflowGasResourceType. Values are in the chain's native
	// smallest unit (wei, lamports). The capability event id is the tx hash.
	ResourceTypeWorkflowGasPrefix = WorkflowRecordType + ":gas:"

	// EmittingServiceChainWrite is the Identity.Service used by chain-write
	// capabilities for gas usage records.
	EmittingServiceChainWrite = "chain-write"
	// EmittingServiceWorkflowEngine is the Identity.Service used by the
	// workflow engine for compute usage records.
	EmittingServiceWorkflowEngine = "workflow-engine"

	// WorkflowUsageResourcePool is the Identity.ResourcePool for workflow
	// capability usage records.
	WorkflowUsageResourcePool = "workflow_usage"
)

var errWorkflowUsageResourceID = errors.New("workflow usage resource id: workflow id and execution id must be non-empty and contain no ':'")

// WorkflowGasResourceType returns the gas resource type for a chain selector,
// e.g. "cre:workflow:gas:421614".
func WorkflowGasResourceType(chainSelector uint64) string {
	return ResourceTypeWorkflowGasPrefix + strconv.FormatUint(chainSelector, 10)
}

// WorkflowUsageResourceID returns the Utilization.ResourceId for a workflow
// execution: "<workflow_id>:<execution_id>". Neither component may be empty or
// contain ':', since the consumer splits on the first ':'.
func WorkflowUsageResourceID(workflowID, executionID string) (string, error) {
	if workflowID == "" || executionID == "" || strings.Contains(workflowID, ":") || strings.Contains(executionID, ":") {
		return "", errWorkflowUsageResourceID
	}
	return workflowID + ":" + executionID, nil
}
