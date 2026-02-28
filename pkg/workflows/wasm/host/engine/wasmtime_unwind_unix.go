//go:build unix

package engine

import "github.com/bytecodealliance/wasmtime-go/v28"

// Load testing shows that leaving native unwind info enabled causes a very large slowdown when loading multiple modules.
func setUnwinding(cfg *wasmtime.Config) {
	if cfg == nil {
		panic("wasmtime.Config cannot be nil")
	}
	cfg.SetNativeUnwindInfo(false)
}
