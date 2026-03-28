//go:build !wasm3

package engine

import "fmt"

func newWasm3Engine() (Engine, error) {
	return nil, fmt.Errorf("wasm3 engine not available: build with -tags wasm3")
}
