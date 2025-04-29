package wasm

import (
	"unsafe"
)

//go:wasmimport env send_response
func sendResponse(response unsafe.Pointer, responseLen int32) int32

//go:wasmimport env version_v2
func versionV2()

func NewDonRunner() sdk.DonRunner {
	return newDonRunner()
}

func NewNodeRunner() sdk.NodeRunner {
	return newNodeRunner()
}
