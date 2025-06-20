package main

import (
	"fmt"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/nodeaction"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/standard_tests/internal/rawsdk"
)

func main() {
	dinput := &basicaction.Inputs{InputThing: true}
	doutput := &basicaction.Outputs{}
	rawsdk.MakeRequest("basic-test-action@1.0.0", "PerformAction", pb.Mode_DON, dinput, doutput)

	rawsdk.SwitchModes(int32(pb.Mode_Node))
	ninput := &nodeaction.NodeInputs{InputThing: true}
	noutput := &nodeaction.NodeOutputs{}
	rawsdk.MakeRequest("basic-test-node-action@1.0.0", "PerformAction", pb.Mode_Node, ninput, noutput)
	rawsdk.SwitchModes(int32(pb.Mode_DON))

	dft := &nodeaction.NodeOutputs{OutputThing: 123}
	consensus := &pb.SimpleConsensusInputs{
		Observation: &pb.SimpleConsensusInputs_Value{Value: values.Proto(rawsdk.Must(values.Wrap(noutput)))},
		Descriptors: rawsdk.NodeOutputConsensusDescriptor,
		Default:     values.Proto(rawsdk.Must(values.Wrap(dft))),
	}

	coutput := &nodeaction.NodeOutputs{}
	if err := rawsdk.Must(values.FromProto(consensus.GetValue())).UnwrapTo(coutput); err != nil {
		rawsdk.SendError(fmt.Errorf("failed to unwrap consensus output: %w", err))
	}

	rawsdk.SendResponse(fmt.Sprintf("%s%d", doutput.AdaptedThing, coutput.OutputThing))
}
