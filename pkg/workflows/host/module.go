//go:generate go run ./requirements_gen

package host

import (
	"context"
	"time"

	sdkpb "github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	wfpb "github.com/smartcontractkit/chainlink-protos/workflows/go/v2"
)

type ModuleBase interface {
	Start()
	Close()
	IsLegacyDAG() bool
}

type Module interface {
	ModuleBase

	// V2/"NoDAG" API - request either the list of Trigger Subscriptions or launch workflow execution
	Execute(ctx context.Context, request *sdkpb.ExecuteRequest, handler ExecutionHelper) (*sdkpb.ExecutionResult, error)
}

type RequirementEnforcingModule interface {
	Module

	// SetRequirements must respect the requirements for the execution until it completes
	SetRequirements(executionId string, requirements *sdkpb.Requirements)
}

// ExecutionHelper Implemented by those running the host, for example the Workflow Engine
type ExecutionHelper interface {
	// CallCapability blocking call to the Workflow Engine
	CallCapability(ctx context.Context, request *sdkpb.CapabilityRequest) (*sdkpb.CapabilityResponse, error)
	GetSecrets(ctx context.Context, request *sdkpb.GetSecretsRequest) ([]*sdkpb.SecretResponse, error)

	GetWorkflowExecutionID() string

	GetNodeTime() time.Time

	GetDONTime() (time.Time, error)

	EmitUserLog(log string) error

	EmitUserMetric(ctx context.Context, metric *wfpb.WorkflowUserMetric) error
}
