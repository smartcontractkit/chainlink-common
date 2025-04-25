package wasm

import (
	"unsafe"
)

//go:wasmimport env call_capability
func callCapability(req unsafe.Pointer, reqLen int32, id unsafe.Pointer) int64

//go:wasmimport env await_capabilities
func awaitCapabilities(awaitRequest unsafe.Pointer, awaitRequestLen int32, responseBuffer unsafe.Pointer, maxResponseLen int32) int64
