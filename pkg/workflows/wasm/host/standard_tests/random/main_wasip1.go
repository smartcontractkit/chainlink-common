package main

import (
	"math/rand"
	"strconv"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/nodeaction"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
	valuespb "github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
)

func main() {
	dr := rand.New(rand.NewSource(rawsdk.GetSeed(int32(sdk.Mode_MODE_DON))))
	total := dr.Uint64()

	nr := rand.New(rand.NewSource(rawsdk.GetSeed(int32(sdk.Mode_MODE_NODE))))
	rawsdk.SwitchModes(int32(sdk.Mode_MODE_NODE))
	result := &nodeaction.NodeOutputs{}
	input := &nodeaction.NodeInputs{InputThing: true}
	rawsdk.DoRequest("basic-test-node-action@1.0.0", "PerformAction", sdk.Mode_MODE_NODE, input, result)
	if result.OutputThing < 100 {
		msg := []byte("***" + strconv.FormatUint(nr.Uint64(), 10))
		rawsdk.Log(rawsdk.BufferToPointerLen(msg))
	}

	dft := &nodeaction.NodeOutputs{OutputThing: 123}
	consensus := &sdk.SimpleConsensusInputs{
		Observation: &sdk.SimpleConsensusInputs_Value{Value: values.Proto(rawsdk.Must(values.Wrap(result)))},
		Descriptors: rawsdk.NodeOutputConsensusDescriptor,
		Default:     values.Proto(rawsdk.Must(values.Wrap(dft))),
	}

	cresult := &valuespb.Value{}
	rawsdk.DoConsensusRequest("consensus@1.0.0-alpha", consensus, cresult)
	rawsdk.SwitchModes(int32(sdk.Mode_MODE_DON))
	total += dr.Uint64()
	rawsdk.SendResponse(total)
}
