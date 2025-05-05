package testutils

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2"
	"github.com/smartcontractkit/chainlink-common/pkg/workflows/sdk/v2/pb"
)

type runner[T any] struct {
	tb             testing.TB
	ran            bool
	config         []byte
	result         any
	err            error
	workflowId     string
	executionId    string
	registry       *Registry
	strictTriggers bool
	runtime        T
	writer         *testWriter
}

func (r *runner[T]) Logs() []string {
	logs := make([]string, len(r.writer.logs))
	for i, log := range r.writer.logs {
		logs[i] = string(log)
	}
	return logs
}

func (r *runner[T]) LogWriter() io.Writer {
	return r.writer
}

type TestRunner interface {
	Result() (bool, any, error)

	// SetStrictTriggers causes the workflow to fail if a trigger isn'tb in the registry
	// this is useful for testing the workflow registrations.
	SetStrictTriggers(strict bool)

	// SetDefaultLogger sets the default logger to write to logs.
	// This allows workflows that use the logger to behave as-if they were a WASM.
	SetDefaultLogger()

	Logs() []string
}

type DonRunner interface {
	sdk.DonRunner
	TestRunner
}

type NodeRunner interface {
	sdk.NodeRunner
	TestRunner
}

func NewDonRunner(tb testing.TB, config []byte) DonRunner {
	return newRunner[sdk.DonRuntime](tb, config, &runtime[sdk.DonRuntime]{})
}

func NewNodeRunner(tb testing.TB, config []byte) NodeRunner {
	return newRunner[sdk.NodeRuntime](tb, config, &runtime[sdk.NodeRuntime]{})
}

func newRunner[T any](tb testing.TB, config []byte, t T) *runner[T] {
	r := &runner[T]{
		tb:          tb,
		config:      config,
		workflowId:  uuid.NewString(),
		executionId: uuid.NewString(),
		registry:    GetRegistry(tb),
		runtime:     t,
		writer:      &testWriter{},
	}

	tmp := any(r.runtime).(*runtime[T])
	tmp.runner = r

	return r
}

func (r *runner[T]) nodeRunner() *runner[sdk.NodeRuntime] {
	rt := &runtime[sdk.NodeRuntime]{}
	tmp := &runner[sdk.NodeRuntime]{
		ran:            r.ran,
		config:         r.config,
		result:         r.result,
		err:            r.err,
		workflowId:     r.workflowId,
		executionId:    r.executionId,
		registry:       r.registry,
		strictTriggers: r.strictTriggers,
		runtime:        rt,
		tb:             r.tb,
	}
	rt.runner = tmp
	return tmp
}

func (r *runner[T]) SetStrictTriggers(strict bool) {
	r.strictTriggers = strict
}
func (r *runner[T]) Run(args *sdk.WorkflowArgs[T]) {
	for _, handler := range args.Handlers {
		trigger, err := r.registry.GetCapability(handler.Id())
		if err != nil {
			if r.strictTriggers {
				r.err = err
			}
			return
		}

		request := &pb.TriggerSubscription{
			ExecId:  r.executionId,
			Id:      uuid.NewString(),
			Payload: handler.TriggerCfg(),
			Method:  handler.Method(),
		}

		response, err := trigger.InvokeTrigger(r.tb.Context(), request)

		var nostub ErrNoTriggerStub
		if err != nil && (r.strictTriggers || !errors.As(err, &nostub)) {
			r.err = err
			return
		}

		// trigger did not fire
		if response == nil {
			return
		}

		if r.ran {
			r.err = TooManyTriggers{}
			return
		}

		r.ran = true
		r.result, r.err = handler.Callback()(r.runtime, response.Payload)
	}
}

func (r *runner[T]) Config() []byte {
	return r.config
}

func (r *runner[T]) Result() (bool, any, error) {
	return r.ran, r.result, r.err
}

func (r *runner[T]) SetDefaultLogger() {
	slog.SetDefault(slog.New(slog.NewTextHandler(r.LogWriter(), nil)))
}

var _ sdk.DonRunner = &runner[sdk.DonRuntime]{}
var _ sdk.NodeRunner = &runner[sdk.NodeRuntime]{}

type TooManyTriggers struct{}

func (e TooManyTriggers) Error() string {
	return "too many triggers fired during execution"
}
