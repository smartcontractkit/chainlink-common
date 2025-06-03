package wasm

import (
	"log/slog"
	"unsafe"
)

//go:wasmimport env log
func log(message unsafe.Pointer, messageLen int32)

func init() {
	slog.SetDefault(slog.New(slog.NewTextHandler(&writer{}, nil)))
}
