package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
)

func main() {
	input1 := &basicaction.Inputs{InputThing: true}
	input2 := &basicaction.Inputs{InputThing: false}
	r1Id := rawsdk.DoRequestAsync("basic-test-action@1.0.0", "PerformAction", pb.Mode_MODE_DON, input1)
	r2Id := rawsdk.DoRequestAsync("basic-test-action@1.0.0", "PerformAction", pb.Mode_MODE_DON, input2)

	results2 := &basicaction.Outputs{}
	rawsdk.Await(r2Id, results2)
	results1 := &basicaction.Outputs{}

	rawsdk.Await(r1Id, results1)
	rawsdk.SendResponse(results1.AdaptedThing + results2.AdaptedThing)
}
