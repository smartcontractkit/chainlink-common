package main

import (
	"fmt"
	"time"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/basicaction"
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pkg/test_capabilities/nodeaction"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
)

func main() {
	rawsdk.SwitchModes(int32(pb.Mode_MODE_DON))
	ignoreTimeCall()

	dinput := &basicaction.Inputs{InputThing: true}
	doutput := &basicaction.Outputs{}
	rawsdk.DoRequest("basic-test-action@1.0.0", "PerformAction", pb.Mode_MODE_DON, dinput, doutput)

	rawsdk.SwitchModes(int32(pb.Mode_MODE_NODE))
	ignoreTimeCall()

	ninput := &nodeaction.NodeInputs{InputThing: true}
	noutput := &nodeaction.NodeOutputs{}
	rawsdk.DoRequest("basic-test-node-action@1.0.0", "PerformAction", pb.Mode_MODE_NODE, ninput, noutput)

	rawsdk.SwitchModes(int32(pb.Mode_MODE_DON))
	ignoreTimeCall()

	dft := &nodeaction.NodeOutputs{OutputThing: 123}
	consensus := &pb.SimpleConsensusInputs{
		Observation: &pb.SimpleConsensusInputs_Value{Value: values.Proto(rawsdk.Must(values.Wrap(noutput)))},
		Descriptors: rawsdk.NodeOutputConsensusDescriptor,
		Default:     values.Proto(rawsdk.Must(values.Wrap(dft))),
	}

	cresult := &valuespb.Value{}
	rawsdk.DoConsensusRequest("consensus@1.0.0-alpha", consensus, cresult)

	coutput := &nodeaction.NodeOutputs{}
	if err := rawsdk.Must(values.FromProto(cresult)).UnwrapTo(coutput); err != nil {
		rawsdk.SendError(fmt.Errorf("failed to unwrap consensus output: %w", err))
	}

	rawsdk.SendResponse(fmt.Sprintf("%s%d", doutput.AdaptedThing, coutput.OutputThing))
}

// ignoreTimeCall makes a rawsdk now call and forces the compiler not to optimize it away.
func ignoreTimeCall() {
	t, err := rawsdk.Now()
	if err != nil {
		panic(err)
	}
	if t.Before(time.Unix(-1, 322)) {
		panic("Test should not run before 1970")
	}
}
