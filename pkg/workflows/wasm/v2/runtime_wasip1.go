package wasm

import (
	"unsafe"
)

//go:wasmimport env call_capability
func callCapability(req unsafe.Pointer, reqLen int32) int64

//go:wasmimport env await_capabilities
func awaitCapabilities(awaitRequest unsafe.Pointer, awaitRequestLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64

//go:wasmimport env get_secrets
func getSecrets(req unsafe.Pointer, reqLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64

//go:wasmimport env await_secrets
func awaitSecrets(req unsafe.Pointer, reqLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64

//go:wasmimport env random_seed
func getSeed(mode int32) int64

type runtimeInternalsImpl struct{}

var _ runtimeInternals = runtimeInternalsImpl{}

func (r runtimeInternalsImpl) callCapability(req unsafe.Pointer, reqLen int32) int64 {
	return callCapability(req, reqLen)
}

func (r runtimeInternalsImpl) awaitCapabilities(awaitRequest unsafe.Pointer, awaitRequestLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64 {
	return awaitCapabilities(awaitRequest, awaitRequestLen, responseBuffer, maxResponseLen)
}

func (r runtimeInternalsImpl) getSecrets(req unsafe.Pointer, reqLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64 {
	return getSecrets(req, reqLen, responseBuffer, maxResponseLen)
}

func (r runtimeInternalsImpl) awaitSecrets(awaitRequest unsafe.Pointer, awaitRequestLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64 {
	return awaitSecrets(awaitRequest, awaitRequestLen, responseBuffer, maxResponseLen)
}

func (r runtimeInternalsImpl) switchModes(mode int32) {
	switchModes(mode)
}

func (r runtimeInternalsImpl) getSeed(mode int32) int64 {
	return getSeed(mode)
}
