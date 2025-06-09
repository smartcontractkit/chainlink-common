package wasm

import (
	"os"
	"unsafe"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	sdkpb "github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

//go:wasmimport env send_response
func sendResponse(response unsafe.Pointer, responseLen int32) int32

//go:wasmimport env version_v2
func versionV2()

//go:wasmimport env switch_modes
func switchModes(mode int32)

func NewDonRunner() sdk.DonRunner {
	switchModes((int32)(sdkpb.Mode_DON))
	return newDonRunner(runnerInternalsImpl{}, runtimeInternalsImpl{})
}

func NewNodeRunner() sdk.NodeRunner {
	switchModes((int32)(sdkpb.Mode_Node))
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

func (r runnerInternalsImpl) switchModes(mode int32) {
	switchModes(mode)
}
