package wasm

import (
	"os"
	"unsafe"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
)

//go:wasmimport env send_response
func sendResponse(response unsafe.Pointer, responseLen int32) int32

//go:wasmimport env version_v2
func versionV2()

func NewDonRunner() sdk.DonRunner {
	return newDonRunner(runnerInternalsImpl{}, runtimeInternalsImpl{})
}

func NewNodeRunner() sdk.NodeRunner {
	return newNodeRunner(runnerInternalsImpl{}, runtimeInternalsImpl{})
}

type runnerInternalsImpl struct{}

var _ runnerInternals = runnerInternalsImpl{}

func (r runnerInternalsImpl) args() []string {
	return os.Args
}

func (r runnerInternalsImpl) sendResponse(response unsafe.Pointer, responseLen int32) int32 {
	return sendResponse(response, responseLen)
}

func (r runnerInternalsImpl) versionV2() {
	versionV2()
}
