package sdk

import (
	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/values"
)

type RuntimeBase interface {
	// CallCapability is meant to be called by generated code
	CallCapability(capId string, request capabilities.CapabilityRequest) Promise[values.Value]

	// AwaitCapabilities is meant to be called by generated code
	AwaitCapabilities(calls ...CapabilityCallPromise) error
}

type NodeRuntime interface {
	RuntimeBase
	IsNodeRuntime()
}

type DonRuntime interface {
	RuntimeBase
	RunInNodeModeWithConsensus(fn func(nodeRuntime NodeRuntime) (values.Value, error), consensus Consensus) Promise[values.Value]
}

type SimpleConsensus int32

const (
	IdenticalConsensus SimpleConsensus = iota
	MedianOfFields
	Median
)

func (SimpleConsensus) isConsensus() {}

type Consensus interface {
	isConsensus()
}

func RunInNodeModeWithConsensus[T any](runtime DonRuntime, fn func(nodeRuntime NodeRuntime) (T, error), consensus Consensus) (T, error) {
	wrapped := func(nodeRuntime NodeRuntime) (values.Value, error) {
		result, err := fn(nodeRuntime)
		if err != nil {
			return nil, err
		}

		return values.Wrap(result)
	}

	result := runtime.RunInNodeModeWithConsensus(wrapped, consensus)
	val, err := result.Await()

	var t T
	if err != nil {
		return t, err
	}

	err = val.UnwrapTo(&t)
	return t, err
}

type EmptyPromise interface {
	Await() error
}

type Promise[T any] interface {
	Await() (T, error)

	// TODO this
	CapabilityCallPromise
}

type emptyPromise struct {
	underlying Promise[struct{}]
}

func (e *emptyPromise) Await() error {
	_, err := e.underlying.Await()
	return err
}

func ToEmptyPromise(p Promise[struct{}]) EmptyPromise {
	return &emptyPromise{underlying: p}
}

func Then[I, O any](p Promise[I], fn func(I) (O, error)) Promise[O] {
	return &thenPromise[I, O]{Promise: p, fn: fn}
}

type thenPromise[I, O any] struct {
	Promise[I]
	fn func(I) (O, error)
}

func (t *thenPromise[I, O]) Await() (O, error) {
	awaited, err := t.Promise.Await()
	var o O
	if err != nil {
		return o, err
	}

	return t.fn(awaited)
}

// weakly-typed, for the runtime to fulfill
type CapabilityCallPromise interface {
	CallInfo() (ref int32, capId string, request capabilities.CapabilityRequest)
	Fulfill(response capabilities.CapabilityResponse, err error)
}
