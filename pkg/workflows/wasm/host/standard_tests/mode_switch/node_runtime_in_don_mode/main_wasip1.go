package main

import (
	"errors"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/standard_tests/internal/rawsdk"
)

func main() {
	// The real SDKs do something to capture the runtime.
	// This is to mimic the mode switch calls they would make
	rawsdk.SwitchModes(int32(pb.Mode_Node))
	rawsdk.SwitchModes(int32(pb.Mode_DON))
	rawsdk.SendError(errors.New("cannot use NodeRuntime outside RunInNodeMode"))
}
