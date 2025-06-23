package sdk

import (
	"errors"
	"math/rand"
	"reflect"

	"github.com/smartcontractkit/chainlink-common/pkg/values"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

// RuntimeBase is not thread safe and must not be used concurrently.
type RuntimeBase interface {
	// CallCapability is meant to be called by generated code
	CallCapability(request *pb.CapabilityRequest) Promise[*pb.CapabilityResponse]
	Rand() (*rand.Rand, error)
}

// NodeRuntime is not thread safe and must not be used concurrently.
type NodeRuntime interface {
	RuntimeBase
	IsNodeRuntime()
}

// Runtime is not thread safe and must not be used concurrently.
type Runtime interface {
	RuntimeBase

	// RunInNodeMode is meant to be used by the helper method RunInNodeMode
	RunInNodeMode(fn func(nodeRuntime NodeRuntime) *pb.SimpleConsensusInputs) Promise[values.Value]
}

type ConsensusAggregation[T any] interface {
	Descriptor() *pb.ConsensusDescriptor
	Default() *T
	Err() error
	WithDefault(t T) ConsensusAggregation[T]
}

type consensusDescriptor[T any] pb.ConsensusDescriptor

func (c *consensusDescriptor[T]) Descriptor() *pb.ConsensusDescriptor {
	return (*pb.ConsensusDescriptor)(c)
}

func (c *consensusDescriptor[T]) Default() *T {
	return nil
}
func (c *consensusDescriptor[T]) Err() error {
	return nil
}

func (c *consensusDescriptor[T]) WithDefault(t T) ConsensusAggregation[T] {
	return &consensusWithDefault[T]{
		ConsensusDescriptor: c.Descriptor(),
		DefaultValue:        t,
	}
}

var _ ConsensusAggregation[int] = (*consensusDescriptor[int])(nil)

type consensusWithDefault[T any] struct {
	ConsensusDescriptor *pb.ConsensusDescriptor
	DefaultValue        T
}

func (c *consensusWithDefault[T]) Descriptor() *pb.ConsensusDescriptor {
	return c.ConsensusDescriptor
}

func (c *consensusWithDefault[T]) Default() *T {
	cpy := c.DefaultValue
	return &cpy
}

func (c *consensusWithDefault[T]) Err() error {
	return nil
}

func (c *consensusWithDefault[T]) WithDefault(t T) ConsensusAggregation[T] {
	return &consensusWithDefault[T]{
		ConsensusDescriptor: c.ConsensusDescriptor,
		DefaultValue:        t,
	}
}

type consensusDescriptorError[T any] struct {
	Error error
}

func (d *consensusDescriptorError[T]) Descriptor() *pb.ConsensusDescriptor {
	return nil
}

func (d *consensusDescriptorError[T]) Default() *T {
	return nil
}

func (d *consensusDescriptorError[T]) Err() error {
	return d.Error
}

func (d *consensusDescriptorError[T]) WithDefault(_ T) ConsensusAggregation[T] {
	return d
}

var nodeModeCallInDonMode = errors.New("cannot use NodeRuntime outside RunInNodeMode")

func NodeModeCallInDonMode() error {
	return nodeModeCallInDonMode
}

var donModeCallInNodeMode = errors.New("cannot use the Runtime inside RunInNodeMode")

func DonModeCallInNodeMode() error {
	return donModeCallInNodeMode
}

func RunInNodeMode[C, T any](
	env *Environment[C],
	runtime Runtime,
	fn func(env *NodeEnvironment[C], nodeRuntime NodeRuntime) (T, error),
	cd ConsensusAggregation[T]) Promise[T] {
	observationFn := func(nodeRuntime NodeRuntime) *pb.SimpleConsensusInputs {
		envClone := env.NodeEnvironment
		if cd.Err() != nil {
			return &pb.SimpleConsensusInputs{Observation: &pb.SimpleConsensusInputs_Error{Error: cd.Err().Error()}}
		}

		var defaultValue values.Value
		descriptor := cd.Descriptor()
		var err error
		if d := cd.Default(); d != nil {
			defaultValue, err = values.Wrap(d)
			if err != nil {
				return &pb.SimpleConsensusInputs{Observation: &pb.SimpleConsensusInputs_Error{Error: err.Error()}}
			}
		}

		returnValue := &pb.SimpleConsensusInputs{
			Descriptors: descriptor,
			Default:     values.Proto(defaultValue),
		}

		result, err := fn(&envClone, nodeRuntime)
		if err != nil {
			returnValue.Observation = &pb.SimpleConsensusInputs_Error{Error: err.Error()}
			return returnValue
		}

		wrapped, err := values.Wrap(result)
		if err != nil {
			returnValue.Observation = &pb.SimpleConsensusInputs_Error{Error: err.Error()}
			return returnValue
		}

		returnValue.Observation = &pb.SimpleConsensusInputs_Value{Value: values.Proto(wrapped)}
		return returnValue
	}

	return Then(runtime.RunInNodeMode(observationFn), func(v values.Value) (T, error) {
		var t T
		var err error
		typ := reflect.TypeOf(t)

		// If T is a pointer type, we need to allocate the underlying type and pass its pointer to UnwrapTo
		if typ.Kind() == reflect.Ptr {
			elem := reflect.New(typ.Elem())
			err = v.UnwrapTo(elem.Interface())
			t = elem.Interface().(T)
		} else {
			err = v.UnwrapTo(&t)
		}
		return t, err
	})
}
