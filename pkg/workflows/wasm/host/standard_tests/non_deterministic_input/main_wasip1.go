package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

// This workflow derives its capability request input from the host-provided DON
// random seed. The suspend/resume integrity check requires a resumed (replayed)
// execution to issue exactly the same capability requests as the original run.
// The accompanying test changes the seed between the suspended run and the
// resume, so the replay builds a different request for the same callback id and
// the host rejects it as non-deterministic.
func main() {
	seed := rawsdk.GetSeed(int32(sdk.Mode_MODE_DON))

	// The request payload depends on the seed, so changing the seed across a
	// replay produces a different request.
	input := &basicaction.Inputs{InputThing: seed%2 == 0}
	id := rawsdk.DoRequestAsync("basic-test-action@1.0.0", "PerformAction", sdk.Mode_MODE_DON, input)

	result := &basicaction.Outputs{}
	rawsdk.Await(id, result)

	rawsdk.SendResponse(result.AdaptedThing)
}
