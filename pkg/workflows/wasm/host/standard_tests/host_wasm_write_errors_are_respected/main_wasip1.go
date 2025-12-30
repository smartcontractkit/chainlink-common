package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func main() {
	input := &basicaction.Inputs{InputThing: true}
	rId := rawsdk.DoRequestAsync("basic-test-action@1.0.0", "PerformAction", sdk.Mode_MODE_DON, input)

	rawsdk.Await(rId, &basicaction.Outputs{})
	rawsdk.SendResponse("should not get here as Await sends error on errors...")
}
