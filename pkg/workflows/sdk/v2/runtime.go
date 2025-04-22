package sdk

import (
	"errors"
	"io"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/v2/pb"
)

type RuntimeBase interface {
	// CallCapability is meant to be called by generated code
	CallCapability(request *pb.CapabilityRequest) Promise[*pb.CapabilityResponse]
	Config() []byte
	LogWriter() io.Writer
}

type NodeRuntime interface {
	RuntimeBase
	IsNodeRuntime()
}

type DonRuntime interface {
	RuntimeBase

	// RunInNodeMode is meant to be used by the helper method RunInNodeMode
	RunInNodeMode(fn func(nodeRuntime NodeRuntime) *pb.BuiltInConsensusRequest) Promise[values.Value]
}

type PrimitiveConsensusWithDefault[T any] struct {
	pb.SimpleConsensusType
	DefaultValue T
}

type BuiltInConsensus[T any] interface {
	pb.SimpleConsensusType | *PrimitiveConsensusWithDefault[T]
}

var nodeModeCallInDonMode = errors.New("cannot use NodeRuntime outside RunInNodeMode")

func NodeModeCallInDonMode() error {
	return nodeModeCallInDonMode
}

var donModeCallInNodeMode = errors.New("cannot use the DonRuntime inside RunInNodeMode")

func DonModeCallInNodeMode() error {
	return donModeCallInNodeMode
}

func RunInNodeMode[T any, C BuiltInConsensus[T]](runtime DonRuntime, fn func(nodeRuntime NodeRuntime) (T, error), consensus C) Promise[T] {
	observationFn := func(nodeRuntime NodeRuntime) *pb.BuiltInConsensusRequest {
		var primitiveConsensus *pb.PrimitiveConsensus
		var defaultValue values.Value
		var err error
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

		consensusRequest := &pb.BuiltInConsensusRequest{
			PrimitiveConsensus: primitiveConsensus,
			DefaultValue:       values.Proto(defaultValue),
		}

		result, err := fn(nodeRuntime)
		if err == nil {
			wrapped, err := values.Wrap(result)
			if err != nil {
				consensusRequest.Observation = &pb.BuiltInConsensusRequest_Error{Error: err.Error()}
			} else {
				consensusRequest.Observation = &pb.BuiltInConsensusRequest_Value{Value: values.Proto(wrapped)}
			}
		} else {
			consensusRequest.Observation = &pb.BuiltInConsensusRequest_Error{Error: err.Error()}
		}

		return consensusRequest
	}

	return Then(runtime.RunInNodeMode(observationFn), func(v values.Value) (T, error) {
		// TODO this is wrong, but good enough for now...
		var t T
		err := v.UnwrapTo(&t)
		return t, err
	})
}
