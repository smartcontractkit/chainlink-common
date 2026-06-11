package main

import (
	"math"
	// import for v2
	_ "github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
)

//go:wasmimport env call_capability
func callCapability(req int32, reqLen int32) int64

func main() {
	// pointer location needs to be somewhere in the real memory range
	// WASM uses flat memory so 0-max memory should work
	malisiousPtrLoc := int32(100)

	// Overflow an int32 with the prl location + length.
	malisiousPtrSize := int32(math.MaxInt32) - malisiousPtrLoc + 1
	callCapability(malisiousPtrLoc, malisiousPtrSize)
}
