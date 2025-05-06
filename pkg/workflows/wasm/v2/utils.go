package wasm

import (
	"fmt"
	"unsafe"
)

// bufferToPointerLen returns a pointer to the first element of the buffer and the length of the buffer.
func bufferToPointerLen(buf []byte) (unsafe.Pointer, int32, error) {
	if len(buf) == 0 {
		return nil, 0, fmt.Errorf("buffer cannot be empty")
	}
	return unsafe.Pointer(&buf[0]), int32(len(buf)), nil
}
