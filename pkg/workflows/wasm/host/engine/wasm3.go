//go:build wasm3

package engine

import "fmt"

func newWasm3Engine() (Engine, error) {
	return &wasm3Engine{}, nil
}

type wasm3Engine struct{}

func (e *wasm3Engine) Load(binary []byte, cfg LoadConfig) (Runtime, error) {
	return nil, fmt.Errorf("wasm3 engine: not yet implemented")
}
