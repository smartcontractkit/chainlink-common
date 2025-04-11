package sdk

import (
	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

type RuntimeBase interface {
	// CallCapability is meant to be called by generated code
	CallCapability(request *pb.CapabilityRequest) Promise[*pb.CapabilityResponse]
	Config() []byte
}

type NodeRuntime interface {
	RuntimeBase
	IsNodeRuntime()
}

type DonRuntime interface {
	RuntimeBase

	// RunInNodeModeWithBuiltInConsensus is meant to be used by the helper method RunInNodeModeWithBuiltInConsensus
	RunInNodeModeWithBuiltInConsensus(fn func(nodeRuntime NodeRuntime) *pb.BuiltInConsensusRequest) Promise[values.Value]
}

type PrimitiveConsensusWithDefault[T any] struct {
	pb.SimpleConsensusType
	DefaultValue T
}

type BuiltInConsensus[T any] interface {
	pb.SimpleConsensusType | *PrimitiveConsensusWithDefault[T]
}

func RunInNodeModeWithBuiltInConsensus[T any, C BuiltInConsensus[T]](runtime DonRuntime, fn func(nodeRuntime NodeRuntime) (T, error), consensus C) Promise[T] {
	observationFn := func(nodeRuntime NodeRuntime) *pb.BuiltInConsensusRequest {
		result, err := fn(nodeRuntime)
		if err != nil {
			return &pb.BuiltInConsensusRequest{
				Observation: &pb.BuiltInConsensusRequest_Error{Error: err.Error()},
			}
		}

		wrapped, err := values.Wrap(result)
		if err != nil {
			return &pb.BuiltInConsensusRequest{
				Observation: &pb.BuiltInConsensusRequest_Error{Error: err.Error()},
			}
		}

		var primitiveConsensus *pb.PrimitiveConsensus
		var defaultValue values.Value
		switch c := any(consensus).(type) {
		case *PrimitiveConsensusWithDefault[T]:
			defaultValue, err = values.Wrap(c.DefaultValue)
			if err != nil {
				return &pb.BuiltInConsensusRequest{Observation: &pb.BuiltInConsensusRequest_Error{Error: err.Error()}}
			}
			primitiveConsensus = &pb.PrimitiveConsensus{
				Consensus: &pb.PrimitiveConsensus_Simple{Simple: c.SimpleConsensusType},
			}
		case pb.SimpleConsensusType:
			primitiveConsensus = &pb.PrimitiveConsensus{Consensus: &pb.PrimitiveConsensus_Simple{Simple: c}}
		}

		return &pb.BuiltInConsensusRequest{
			PrimitiveConsensus: primitiveConsensus,
			Observation:        &pb.BuiltInConsensusRequest_Value{Value: values.Proto(wrapped)},
			DefaultValue:       values.Proto(defaultValue),
		}
	}

	return Then(runtime.RunInNodeModeWithBuiltInConsensus(observationFn), func(v values.Value) (T, error) {
		// TODO this is wrong, but good enough for now...
		var t T
		err := v.UnwrapTo(&t)
		return t, err
	})
}
