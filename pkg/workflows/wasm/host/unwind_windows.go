//go:build windows

package host

import "github.com/bytecodealliance/wasmtime-go/v28"

func SetUnwinding(cfg *wasmtime.Config) {
	// Unwinding cannot be disabled on Windows.
}
