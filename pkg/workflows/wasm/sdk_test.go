package wasm

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"

	"github.com/stretchr/testify/assert"
)

func Test_toEmitLabels(t *testing.T) {
	t.Run("successfully transforms metadata", func(t *testing.T) {
		md := &capabilities.RequestMetadata{
			WorkflowID:          "workflow-id",
			WorkflowName:        "workflow-name",
			WorkflowOwner:       "workflow-owner",
			WorkflowExecutionID: "6e2a46e3b6ae611bdb9bcc36ed3f46bb9a30babc3aabdd4eae7f35dd9af0f244",
		}
		empty := make(map[string]string, 0)

		gotLabels, err := toEmitLabels(md, empty)
		assert.NoError(t, err)

		assert.Equal(t, map[string]string{
			"workflow_id":            "workflow-id",
			"workflow_name":          "workflow-name",
			"workflow_owner_address": "workflow-owner",
			"workflow_execution_id":  "6e2a46e3b6ae611bdb9bcc36ed3f46bb9a30babc3aabdd4eae7f35dd9af0f244",
		}, gotLabels)
	})

	t.Run("fails on missing workflow id", func(t *testing.T) {
		md := &capabilities.RequestMetadata{
			WorkflowName:  "workflow-name",
			WorkflowOwner: "workflow-owner",
		}
		empty := make(map[string]string, 0)

		_, err := toEmitLabels(md, empty)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "workflow id")
	})

	t.Run("fails on missing workflow name", func(t *testing.T) {
		md := &capabilities.RequestMetadata{
			WorkflowID:    "workflow-id",
			WorkflowOwner: "workflow-owner",
		}
		empty := make(map[string]string, 0)

		_, err := toEmitLabels(md, empty)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "workflow name")
	})

	t.Run("fails on missing workflow owner", func(t *testing.T) {
		md := &capabilities.RequestMetadata{
			WorkflowID:   "workflow-id",
			WorkflowName: "workflow-name",
		}
		empty := make(map[string]string, 0)

		_, err := toEmitLabels(md, empty)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "workflow owner")
	})

	t.Run("fails on missing workflow execution id", func(t *testing.T) {
		md := &capabilities.RequestMetadata{
			WorkflowID:    "workflow-id",
			WorkflowName:  "workflow-name",
			WorkflowOwner: "workflow-owner",
		}
		empty := make(map[string]string, 0)

		_, err := toEmitLabels(md, empty)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "workflow execution id")
	})
}

func Test_bufferToPointerLen(t *testing.T) {
	t.Run("fails when no buffer", func(t *testing.T) {
		_, _, err := bufferToPointerLen([]byte{})
		assert.Error(t, err)
		assert.ErrorContains(t, err, "buffer cannot be empty")
	})
}
