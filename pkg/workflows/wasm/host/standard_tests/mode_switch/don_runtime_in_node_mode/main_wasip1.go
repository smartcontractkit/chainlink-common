package main

import (
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
)

func main() {
	// The real SDKs do something to capture the runtime.
	// This is to mimic the mode switch calls they would make
	rawsdk.SwitchModes(int32(pb.Mode_MODE_NODE))

	consensus := &pb.SimpleConsensusInputs{
		Observation: &pb.SimpleConsensusInputs_Error{Error: "cannot use Runtime inside RunInNodeMode"},
		Descriptors: &pb.ConsensusDescriptor{Descriptor_: &pb.ConsensusDescriptor_Aggregation{Aggregation: pb.AggregationType_AGGREGATION_TYPE_IDENTICAL}},
	}

	err := rawsdk.DoRequestErr("consensus@1.0.0-alpha", "Simple", pb.Mode_MODE_DON, consensus)
	rawsdk.SwitchModes(int32(pb.Mode_MODE_DON))
	rawsdk.SendError(err)
}
