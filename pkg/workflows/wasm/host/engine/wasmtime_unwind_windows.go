//go:build windows

package engine

import "github.com/bytecodealliance/wasmtime-go/v28"

func setUnwinding(_ *wasmtime.Config) {
	// Unwinding cannot be disabled on Windows.
}
