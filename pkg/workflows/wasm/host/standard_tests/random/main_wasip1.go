package main

import (
	"math/rand"
	"strconv"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/nodeaction"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/standard_tests/internal/rawsdk"
)

func main() {
	dr := rand.New(rand.NewSource(rawsdk.GetSeed(int32(pb.Mode_DON))))
	total := dr.Uint64()

	nr := rand.New(rand.NewSource(rawsdk.GetSeed(int32(pb.Mode_Node))))
	rawsdk.SwitchModes(int32(pb.Mode_Node))
	result := &nodeaction.NodeOutputs{}
	input := &nodeaction.NodeInputs{InputThing: true}
	rawsdk.MakeRequest("basic-test-node-action@1.0.0", "PerformAction", pb.Mode_Node, input, result)
	if result.OutputThing < 100 {
		msg := []byte(strconv.FormatUint(nr.Uint64(), 10))
		rawsdk.Log(rawsdk.BufferToPointerLen(msg))
	}

	dft := &nodeaction.NodeOutputs{OutputThing: 123}
	consensus := &pb.SimpleConsensusInputs{
		Observation: &pb.SimpleConsensusInputs_Value{Value: values.Proto(rawsdk.Must(values.Wrap(result)))},
		Descriptors: rawsdk.NodeOutputConsensusDescriptor,
		Default:     values.Proto(rawsdk.Must(values.Wrap(dft))),
	}

	cresult := &valuespb.Value{}
	rawsdk.MakeRequest("consensus@1.0.0", "Simple", pb.Mode_DON, consensus, cresult)
	rawsdk.SwitchModes(int32(pb.Mode_DON))
	total += dr.Uint64()
	rawsdk.SendResponse(total)
}
