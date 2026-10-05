package resourcemanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowGasResourceType(t *testing.T) {
	assert.Equal(t, "cre:workflow:gas:421614", WorkflowGasResourceType(421614))
	assert.Equal(t, "cre:workflow:compute", ResourceTypeWorkflowCompute)
}

func TestWorkflowUsageResourceID(t *testing.T) {
	id, err := WorkflowUsageResourceID("wf-1", "exec-1")
	require.NoError(t, err)
	assert.Equal(t, "wf-1:exec-1", id)

	for _, tc := range [][2]string{{"", "exec-1"}, {"wf-1", ""}, {"wf:1", "exec-1"}, {"wf-1", "exec:1"}} {
		_, err := WorkflowUsageResourceID(tc[0], tc[1])
		assert.ErrorIs(t, err, errWorkflowUsageResourceID, "%q %q", tc[0], tc[1])
	}
}

func TestWithWorkflowUsagePool(t *testing.T) {
	base := ResourceIdentity{Product: "cre", Service: EmittingServiceChainWrite}
	gas := WithWorkflowUsagePool(base, WorkflowGasResourceType(421614))
	assert.Equal(t, "cre:workflow:gas", gas.ResourcePool)
	assert.Equal(t, "cre:workflow:gas:421614", gas.ResourcePoolID)
	assert.Equal(t, EmittingServiceChainWrite, gas.Service, "other fields untouched")

	compute := WithWorkflowUsagePool(base, ResourceTypeWorkflowCompute)
	assert.Equal(t, "cre:workflow:compute", compute.ResourcePool)
	assert.Equal(t, "cre:workflow:compute", compute.ResourcePoolID)
}
