package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
)

func main() {
	metadata := rawsdk.GetRequest().GetExecutionMetadata()
	rawsdk.SendResponse(map[string]any{
		"WorkflowID":    metadata.GetWorkflowId(),
		"ExecutionID":   metadata.GetExecutionId(),
		"WorkflowOwner": metadata.GetWorkflowOwner(),
		"DonID":         uint64(metadata.GetDonId()),
		"DonF":          uint64(metadata.GetDonF()),
	})
}
