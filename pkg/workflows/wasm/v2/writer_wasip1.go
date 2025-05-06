package wasm

import (
	"unsafe"
)

//go:wasmimport env log
func log(message unsafe.Pointer, messageLen int32)
