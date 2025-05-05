//go:build !wasip1

package wasm

import (
	"testing"
	"unsafe"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/testutils"
)

// initRunnerAndRuntimeForTest is NOT thread safe and expects a single test to be running when testing the runner.
func initRunnerAndRuntimeForTest(t testing.TB, execId string) {
	registry = testutils.GetRegistry(t)
	outstandingCalls = map[string]sdk.Promise[*sdkpb.CapabilityResponse]{}
	testTb = t
	executionId = execId
	callCapabilityErr = false
	t.Cleanup(func() {
		testTb = nil
		executionId = ""
		outstandingCalls = nil
	})
}

var testTb testing.TB
var registry *testutils.Registry
var outstandingCalls map[string]sdk.Promise[*sdkpb.CapabilityResponse]
var executionId string

var sentResponse []byte

func sendResponse(response unsafe.Pointer, responseLen int32) int32 {
	sentResponse = unsafe.Slice((*byte)(response), responseLen)
	return 0
}

func versionV2() {}
