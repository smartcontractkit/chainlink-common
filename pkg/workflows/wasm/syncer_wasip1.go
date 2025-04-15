package wasm

import "unsafe"

//go:wasmimport env log
func log(message unsafe.Pointer, messageLen int32)

type wasmWriteSyncer struct{}

// Write is used to proxy log requests from the WASM binary back to the host
func (wws *wasmWriteSyncer) Write(p []byte) (n int, err error) {
	ptr, ptrlen, err := bufferToPointerLen(p)
	if err != nil {
		return int(ptrlen), err
	}
	log(ptr, ptrlen)
	return int(ptrlen), nil
}
