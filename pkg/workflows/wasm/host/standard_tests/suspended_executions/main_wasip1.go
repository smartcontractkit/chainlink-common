package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

// This workflow exercises the host's suspend/resume behaviour. It dispatches a
// single async capability call and awaits it. When suspension is enabled the
// host has no response available at the await, so it suspends the guest; the
// guest re-runs main() from the top once the host resumes it with the response.
//
// A log line is emitted on every run so the test can observe how many times the
// guest executed: once when suspension is disabled (the await blocks in the
// host), twice when it is enabled (the initial run that suspends plus the
// resumed run that completes).
func main() {
	// Switch to DON mode so the host attributes (and does not drop) the log below.
	rawsdk.SwitchModes(int32(sdk.Mode_MODE_DON))
	logRun()

	id := rawsdk.DoRequestAsync(
		"basic-test-action@1.0.0",
		"PerformAction",
		sdk.Mode_MODE_DON,
		&basicaction.Inputs{InputThing: true},
	)

	result := &basicaction.Outputs{}
	rawsdk.Await(id, result)

	rawsdk.SendResponse(result.AdaptedThing)
}

func logRun() {
	msg := []byte("suspended_executions:run")
	ptr, length := rawsdk.BufferToPointerLen(msg)
	rawsdk.Log(ptr, length)
}
