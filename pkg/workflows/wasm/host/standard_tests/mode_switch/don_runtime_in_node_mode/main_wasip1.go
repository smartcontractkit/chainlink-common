package main

import (
	"errors"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
)

func main() {
	// The real SDKs do something to capture the runtime.
	// This is to mimic the mode switch calls they would make
	rawsdk.SwitchModes(int32(pb.Mode_MODE_NODE))
	rawsdk.SendError(errors.New("cannot use Runtime inside RunInNodeMode"))
}
