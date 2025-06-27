package main

import (
	"errors"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities/v2/protoc/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	valuespb "github.com/smartcontractkit/chainlink-common/pkg/values/pb"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
)

func main() {
	// The real SDKs do something to capture the runtime.
	// This is to mimic the mode switch calls they would make
	rawsdk.SwitchModes(int32(pb.Mode_MODE_NODE))

	consensus := &pb.SimpleConsensusInputs{
		Observation: &pb.SimpleConsensusInputs_Value{Value: values.Proto(rawsdk.Must(values.Wrap("hi")))},
		Descriptors: &pb.ConsensusDescriptor{Descriptor_: &pb.ConsensusDescriptor_Aggregation{Aggregation: pb.AggregationType_AGGREGATION_TYPE_IDENTICAL}},
	}

	cresult := &valuespb.Value{}
	rawsdk.DoRequest("consensus@1.0.0-alpha", "Simple", pb.Mode_MODE_DON, consensus, cresult)

	rawsdk.SwitchModes(int32(pb.Mode_MODE_DON))
	rawsdk.SendError(errors.New("cannot use NodeRuntime outside RunInNodeMode"))
}
