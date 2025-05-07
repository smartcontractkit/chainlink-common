package wasm

import (
	"io"
)

type writer struct{}

// Write is used to proxy log requests from the WASM binary back to the host
func (w *writer) Write(p []byte) (n int, err error) {
	ptr, ptrlen, err := bufferToPointerLen(p)
	if err != nil {
		return int(ptrlen), err
	}
	log(ptr, ptrlen)
	return int(ptrlen), nil
}

var _ io.Writer = (*writer)(nil)
