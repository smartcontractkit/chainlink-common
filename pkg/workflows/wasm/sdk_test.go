package wasm

import (
	"testing"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"

	"github.com/stretchr/testify/assert"
)

func Test_toEmitLabels(t *testing.T) {
	t.Run("successfully transforms metadata", func(t *testing.T) {
		md := &capabilities.RequestMetadata{
			WorkflowID:    "workflow-id",
			WorkflowName:  "workflow-name",
			WorkflowOwner: "workflow-owner",
		}
		empty := make(map[string]string, 0)

		gotLabels, err := toEmitLabels(md, empty)
		assert.NoError(t, err)

		assert.Equal(t, map[string]string{
			"workflow_id":            "workflow-id",
			"workflow_name":          "workflow-name",
			"workflow_owner_address": "workflow-owner",
			"workflow_execution_id":  "",
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
}
