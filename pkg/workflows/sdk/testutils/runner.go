package testutils

// TODO this could share more with the real engine or possibly be the real engine...?
// I am using it to demonstrate a possible shape of unit testing

import (
	"context"
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/capabilities"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

type runner[T any] struct {
	ran            bool
	config         []byte
	result         any
	err            error
	ctx            context.Context
	workflowId     string
	executionId    string
	registry       *Registry
	strictTriggers bool
}

type TestRunner interface {
	Result() (bool, any, error)

	// SetStrictTriggers causes the workflow to fail if a trigger isn't in the registry
	// this is useful for testing the workflow registrations.
	SetStrictTriggers(strict bool)
}

type DonRunner interface {
	sdk.DonRunner
	TestRunner
}

type NodeRunner interface {
	sdk.NodeRunner
	TestRunner
}

type Registry struct {
	Capabilities map[string]Capability
	Triggers     map[string]Trigger
}

func NewDonRunner(ctx context.Context, config []byte, registry *Registry) (DonRunner, error) {
	return newRunner[sdk.DonRuntime](ctx, config, registry)
}

func NewNodeRunner(ctx context.Context, config []byte, registry *Registry) (NodeRunner, error) {
	return newRunner[sdk.NodeRuntime](ctx, config, registry)
}

func newRunner[T any](ctx context.Context, config []byte, registry *Registry) (*runner[T], error) {

	return &runner[T]{
		config:      config,
		ctx:         ctx,
		workflowId:  uuid.NewString(),
		executionId: uuid.NewString(),
		registry:    registry,
	}, nil
}

func (r *runner[T]) SetStrictTriggers(strict bool) {
	r.strictTriggers = strict
}

func (r *runner[T]) SubscribeToTrigger(id, method string, triggerCfg *anypb.Any, handler func(runtime T, triggerOutputs *anypb.Any) (any, error)) {
	r.ran = true
	if r.err != nil {
		return
	}

	trigger, ok := r.registry.Triggers[id]
	if !ok {
		if r.strictTriggers {
			r.err = fmt.Errorf("trigger %s not found", id)
		}

		return
	}

	request := capabilities.TriggerRegistrationRequest{
		// TODO I think this is the id for the trigger so we can differenciate them
		TriggerID: id,
		Metadata: capabilities.RequestMetadata{
			WorkflowID:          r.workflowId,
			WorkflowOwner:       "mock",
			WorkflowExecutionID: r.executionId,
			WorkflowName:        "testworkflow",
			ReferenceID:         uuid.NewString(),
			DecodedWorkflowName: "test workflow",
		},
		Request: triggerCfg,
		Method:  method,
	}
	ch, err := r.trigger.RegisterTrigger(r.ctx, request)
	if err != nil {
		r.err = err
		return
	}
	// TODO multiple registrations to the same trigger, need some way to know if we expect anything form this registration or not
	trigger := <-ch
	_ = trigger
	r.result, r.err = handler(r.ctx, trigger.Event.Value)
	//handler()
}

func (r *runner[T]) Config() []byte {
	return r.config
}

func (r *runner[T]) Result() (bool, any, error) {
	return r.ran, r.result, r.err
}

var _ sdk.DonRunner = &runner[sdk.DonRuntime]{}
var _ sdk.NodeRunner = &runner[sdk.NodeRuntime]{}
