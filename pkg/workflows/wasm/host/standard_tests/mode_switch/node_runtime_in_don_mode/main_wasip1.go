package main

import (
	"errors"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/host/internal/rawsdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/sdk"
	"github.com/smartcontractkit/chainlink-protos/cre/go/values"
	valuespb "github.com/smartcontractkit/chainlink-protos/cre/go/values/pb"
)

func main() {
	// The real SDKs do something to capture the runtime.
	// This is to mimic the mode switch calls they would make
	rawsdk.SwitchModes(int32(sdk.Mode_MODE_NODE))

	consensus := &sdk.SimpleConsensusInputs{
		Observation: &sdk.SimpleConsensusInputs_Value{Value: values.Proto(rawsdk.Must(values.Wrap("hi")))},
		Descriptors: &sdk.ConsensusDescriptor{Descriptor_: &sdk.ConsensusDescriptor_Aggregation{Aggregation: sdk.AggregationType_AGGREGATION_TYPE_IDENTICAL}},
	}

	cresult := &valuespb.Value{}
	rawsdk.DoConsensusRequest("consensus@1.0.0-alpha", consensus, cresult)

	rawsdk.SwitchModes(int32(sdk.Mode_MODE_DON))
	rawsdk.SendError(errors.New("cannot use NodeRuntime outside RunInNodeMode"))
}
