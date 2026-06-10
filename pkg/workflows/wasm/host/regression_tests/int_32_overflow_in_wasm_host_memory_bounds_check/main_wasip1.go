package main

import (
	"math"
	// import for v2
	_ "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
)

//go:wasmimport env call_capability
func callCapability(req int32, reqLen int32) int64

func main() {
	// pointer size needs to be somehwere in memory
	malisiousPtrLoc := int32(100)
	malisiousPtrSize := int32(math.MaxInt32) - malisiousPtrLoc + 1
	callCapability(malisiousPtrLoc, malisiousPtrSize)
}
