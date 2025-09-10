package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
)

func main() {
	// The real SDKs do something to capture the runtime.
	// This is to mimic the mode switch calls they would make
	rawsdk.SwitchModes(int32(sdk.Mode_MODE_NODE))

	consensus := &sdk.SimpleConsensusInputs{
		Observation: &sdk.SimpleConsensusInputs_Error{Error: "cannot use Runtime inside RunInNodeMode"},
		Descriptors: &sdk.ConsensusDescriptor{Descriptor_: &sdk.ConsensusDescriptor_Aggregation{Aggregation: sdk.AggregationType_AGGREGATION_TYPE_IDENTICAL}},
	}

	err := rawsdk.DoRequestErr("consensus@1.0.0-alpha", "Simple", sdk.Mode_MODE_DON, consensus)
	rawsdk.SwitchModes(int32(sdk.Mode_MODE_DON))
	rawsdk.SendError(err)
}
