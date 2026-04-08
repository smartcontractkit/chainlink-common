package main

import (
	"unsafe"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
)

func main() {
	buf := make([]byte, 4)
	rawsdk.Requirements(unsafe.Pointer(&buf[0]), 100)
	rawsdk.SendResponse(0)
}
