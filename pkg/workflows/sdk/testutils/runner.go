package testutils

// TODO this could share more with the real engine or possibly be the real engine...?
// I am using it to demonstrate a possible shape of unit testing

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/smartcontractkit/chainlink-common/pkg/logger"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/wasm/pb"
)

type runner[T any] struct {
	ran               bool
	config            []byte
	result            any
	err               error
	ctx               context.Context
	workflowId        string
	executionId       string
	registry          *Registry
	strictTriggers    bool
	asyncCapabilities bool
	runtime           T
	logger            logger.Logger
}

func (r *runner[T]) Logger() logger.Logger {
	return r.logger
}

type TestRunner interface {
	Result() (bool, any, error)

	// SetStrictTriggers causes the workflow to fail if a trigger isn't in the registry
	// this is useful for testing the workflow registrations.
	SetStrictTriggers(strict bool)

	// SetCallCapabilitiesAsync creates a go routine to call capabilities
	// Defaults to false
	SetCallCapabilitiesAsync(async bool)
}

type DonRunner interface {
	sdk.DonRunner
	TestRunner
}

type NodeRunner interface {
	sdk.NodeRunner
	TestRunner
}

func NewDonRunner(ctx context.Context, config []byte, registry *Registry, logger logger.Logger) (DonRunner, error) {
	return newRunner[sdk.DonRuntime](ctx, config, registry, &runtime[sdk.DonRuntime]{}, logger)
}

func NewNodeRunner(ctx context.Context, config []byte, registry *Registry, logger logger.Logger) (NodeRunner, error) {
	return newRunner[sdk.NodeRuntime](ctx, config, registry, &runtime[sdk.NodeRuntime]{}, logger)
}

func newRunner[T any](ctx context.Context, config []byte, registry *Registry, t T, logger logger.Logger) (*runner[T], error) {
	r := &runner[T]{
		config:      config,
		ctx:         ctx,
		workflowId:  uuid.NewString(),
		executionId: uuid.NewString(),
		registry:    registry,
		runtime:     t,
		logger:      logger,
	}

	tmp := any(r.runtime).(*runtime[T])
	tmp.runner = r

	return r, nil
}

func (r *runner[T]) SetCallCapabilitiesAsync(async bool) {
	r.asyncCapabilities = async
}

func (r *runner[T]) nodeRunner() *runner[sdk.NodeRuntime] {
	rt := &runtime[sdk.NodeRuntime]{}
	tmp := &runner[sdk.NodeRuntime]{
		ran:               r.ran,
		config:            r.config,
		result:            r.result,
		err:               r.err,
		ctx:               r.ctx,
		workflowId:        r.workflowId,
		executionId:       r.executionId,
		registry:          r.registry,
		strictTriggers:    r.strictTriggers,
		asyncCapabilities: r.asyncCapabilities,
		runtime:           rt,
	}
	rt.runner = tmp
	return tmp
}

func (r *runner[T]) SetStrictTriggers(strict bool) {
	r.strictTriggers = strict
}

func (r *runner[T]) SubscribeToTrigger(id, method string, triggerCfg *anypb.Any, handler func(runtime T, triggerOutputs *anypb.Any) (any, error)) {
	r.ran = true
	if r.err != nil {
		return
	}

	trigger, ok := r.registry.capabilities[id]
	if !ok {
		if r.strictTriggers {
			r.err = fmt.Errorf("trigger %s not found", id)
		}

		return
	}

	request := &pb.TriggerSubscriptionRequest{
		ExecId:  r.executionId,
		Id:      uuid.NewString(),
		Payload: triggerCfg,
		Method:  method,
	}

	// TODO decide if this should be allowed to be async since it's for starting a workflow...
	response, err := trigger.InvokeTrigger(r.ctx, request)
	if err != nil {
		r.err = err
		return
	}

	// trigger did not fire
	if response == nil {
		return
	}

	// TODO multiple results???
	r.result, r.err = handler(r.runtime, response.Payload)
}

func (r *runner[T]) Config() []byte {
	return r.config
}

func (r *runner[T]) Result() (bool, any, error) {
	return r.ran, r.result, r.err
}

var _ sdk.DonRunner = &runner[sdk.DonRuntime]{}
var _ sdk.NodeRunner = &runner[sdk.NodeRuntime]{}
